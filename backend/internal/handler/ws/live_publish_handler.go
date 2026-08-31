package handler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var activeBrowserPublishers = struct {
	sync.Mutex
	rooms map[uint]bool
}{rooms: make(map[uint]bool)}

var browserPublisherRecoveryDelay = 15 * time.Second

// LivePublishWebSocketHandler receives a continuous WebM stream produced by
// MediaRecorder, then lets FFmpeg publish it to the room's RTMP endpoint.
func LivePublishWebSocketHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || roomID64 == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
			return
		}
		userID, ok := getUserIDFromLiveRequest(c, svcCtx)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		roomID := uint(roomID64)
		var room model.LiveRoom
		if err := svcCtx.DB.Where("id = ? AND owner_id = ? AND status = ?", roomID, userID, "live").First(&room).Error; err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "room is not manageable"})
			return
		}
		if _, err := exec.LookPath("ffmpeg"); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ffmpeg is unavailable"})
			return
		}

		activeBrowserPublishers.Lock()
		if activeBrowserPublishers.rooms[roomID] {
			activeBrowserPublishers.Unlock()
			c.JSON(http.StatusConflict, gin.H{"error": "room already has a browser publisher"})
			return
		}
		activeBrowserPublishers.rooms[roomID] = true
		activeBrowserPublishers.Unlock()
		defer releaseBrowserPublisher(svcCtx, roomID)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		ctx, cancel := context.WithCancel(c.Request.Context())
		defer cancel()
		target := fmt.Sprintf("rtmp://%s/%s/%s", svcCtx.Config.Live.RTMPHost, liveAppName, room.StreamKey)
		cmd := exec.CommandContext(ctx, "ffmpeg",
			"-hide_banner", "-loglevel", "warning",
			"-f", "webm", "-i", "pipe:0",
			"-map", "0:v:0", "-map", "0:a?",
			"-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency",
			"-pix_fmt", "yuv420p", "-g", "60",
			"-c:a", "aac", "-ar", "44100", "-b:a", "128k",
			"-f", "flv", target,
		)
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return
		}
		cmd.Stdout = io.Discard
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			return
		}
		defer func() {
			stdin.Close()
			cancel()
			_ = cmd.Wait()
		}()

		for {
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			if messageType != websocket.BinaryMessage || len(data) == 0 {
				continue
			}
			if _, err := stdin.Write(data); err != nil {
				return
			}
		}
	}
}

func releaseBrowserPublisher(svcCtx *svc.ServiceContext, roomID uint) {
	activeBrowserPublishers.Lock()
	delete(activeBrowserPublishers.rooms, roomID)
	activeBrowserPublishers.Unlock()
	time.AfterFunc(browserPublisherRecoveryDelay, func() {
		_ = finalizeAbandonedBrowserPublisher(svcCtx, roomID, time.Now())
	})
}

func finalizeAbandonedBrowserPublisher(svcCtx *svc.ServiceContext, roomID uint, endedAt time.Time) error {
	activeBrowserPublishers.Lock()
	active := activeBrowserPublishers.rooms[roomID]
	activeBrowserPublishers.Unlock()
	if active {
		return nil
	}
	return svcCtx.DB.Model(&model.LiveRoom{}).
		Where("id = ? AND status = ?", roomID, "live").
		Updates(map[string]any{"status": "ended", "viewer_count": 0, "ended_at": &endedAt}).Error
}

const liveAppName = "live"
