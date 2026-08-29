package live

import (
	"errors"
	"net/http"
	"strings"
	"sync"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type srsHookRequest struct {
	Action string `json:"action"`
	App    string `json:"app"`
	Stream string `json:"stream"`
}

type srsStreamState struct {
	Generation uint64
	Active     bool
}

var srsStreams = struct {
	sync.Mutex
	items map[string]srsStreamState
}{items: make(map[string]srsStreamState)}

var srsUnpublishRecoveryDelay = 15 * time.Second

// SRSStreamHookHandler receives SRS on_publish/on_unpublish callbacks. A short
// grace period prevents a transient encoder reconnect from incorrectly ending
// the room, while an abandoned stream eventually restores the room to ended.
func SRSStreamHookHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request srsHookRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 1})
			return
		}
		request.Action = strings.TrimSpace(request.Action)
		request.App = strings.TrimSpace(request.App)
		request.Stream = strings.TrimSpace(request.Stream)
		if request.App != liveApp || request.Stream == "" {
			c.JSON(http.StatusOK, gin.H{"code": 1})
			return
		}

		switch request.Action {
		case "on_publish":
			var count int64
			if err := svcCtx.DB.Model(&model.LiveRoom{}).
				Where("stream_key = ? AND status = ?", request.Stream, "live").Count(&count).Error; err != nil || count == 0 {
				c.JSON(http.StatusOK, gin.H{"code": 1})
				return
			}
			markSRSStream(request.Stream, true)
		case "on_unpublish":
			generation := markSRSStream(request.Stream, false)
			time.AfterFunc(srsUnpublishRecoveryDelay, func() {
				_ = finalizeUnpublishedSRSStream(svcCtx, request.Stream, generation, time.Now())
			})
		default:
			c.JSON(http.StatusOK, gin.H{"code": 1})
			return
		}

		c.JSON(http.StatusOK, gin.H{"code": 0})
	}
}

func markSRSStream(stream string, active bool) uint64 {
	srsStreams.Lock()
	defer srsStreams.Unlock()
	state := srsStreams.items[stream]
	state.Generation++
	state.Active = active
	srsStreams.items[stream] = state
	return state.Generation
}

func finalizeUnpublishedSRSStream(svcCtx *svc.ServiceContext, stream string, generation uint64, endedAt time.Time) error {
	srsStreams.Lock()
	state, ok := srsStreams.items[stream]
	if ok && state.Generation == generation && !state.Active {
		delete(srsStreams.items, stream)
		srsStreams.Unlock()
	} else {
		srsStreams.Unlock()
		return nil
	}

	result := svcCtx.DB.Model(&model.LiveRoom{}).
		Where("stream_key = ? AND status = ?", stream, "live").
		Updates(map[string]any{"status": "ended", "viewer_count": 0, "ended_at": &endedAt})
	if result.Error != nil && !errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return result.Error
	}
	return nil
}
