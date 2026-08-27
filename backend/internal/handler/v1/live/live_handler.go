package live

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"danmakustream/backend/internal/handler/response"
	analyticslogic "danmakustream/backend/internal/logic/analytics"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	liveApp = "live"
)

type createLiveRoomReq struct {
	Title    string `json:"title" binding:"required"`
	CoverURL string `json:"coverUrl"`
}

type updateChatSettingsReq struct {
	ChatMode        string `json:"chatMode" binding:"required"`
	SlowModeSeconds int    `json:"slowModeSeconds"`
	PinnedMessage   string `json:"pinnedMessage"`
}

type liveRoomInfo struct {
	ID              uint            `json:"id"`
	Title           string          `json:"title"`
	CoverURL        string          `json:"coverUrl"`
	StreamKey       string          `json:"streamKey,omitempty"`
	PublishURL      string          `json:"publishUrl,omitempty"`
	PlayURL         string          `json:"playUrl"`
	StreamURL       string          `json:"streamUrl"`
	Status          string          `json:"status"`
	ViewerCount     int64           `json:"viewerCount"`
	ViewerPeak      int64           `json:"viewerPeak"`
	LikeCount       int64           `json:"likeCount"`
	GiftValue       int64           `json:"giftValue"`
	ChatMode        string          `json:"chatMode"`
	SlowModeSeconds int             `json:"slowModeSeconds"`
	PinnedMessage   string          `json:"pinnedMessage"`
	Heat            int64           `json:"heat"`
	OwnerID         uint            `json:"ownerId"`
	Owner           *model.UserInfo `json:"owner,omitempty"`
	StartedAt       string          `json:"startedAt,omitempty"`
	EndedAt         string          `json:"endedAt,omitempty"`
	CreatedAt       string          `json:"createdAt"`
}

type liveReplayInfo struct {
	ID         uint            `json:"id"`
	RoomID     uint            `json:"roomId"`
	Title      string          `json:"title"`
	CoverURL   string          `json:"coverUrl"`
	ReplayURL  string          `json:"replayUrl"`
	Status     string          `json:"status"`
	Duration   int             `json:"duration"`
	ViewerPeak int64           `json:"viewerPeak"`
	OwnerID    uint            `json:"ownerId"`
	Owner      *model.UserInfo `json:"owner,omitempty"`
	StartedAt  string          `json:"startedAt"`
	EndedAt    string          `json:"endedAt"`
	CreatedAt  string          `json:"createdAt"`
}

type replayDanmakuItem struct {
	ID       uint   `json:"id"`
	UserID   uint   `json:"userId"`
	Content  string `json:"content"`
	Time     int    `json:"time"`
	Color    string `json:"color"`
	FontSize string `json:"fontSize"`
	Type     string `json:"type"`
}

func ListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := getPage(c)

		var total int64
		var rooms []model.LiveRoom
		db := svcCtx.DB.Model(&model.LiveRoom{}).Preload("Owner").Where("status = ?", "live")

		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播列表加载失败")
			return
		}

		if err := db.Order("started_at DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&rooms).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播列表加载失败")
			return
		}

		list := make([]liveRoomInfo, 0, len(rooms))
		for _, room := range rooms {
			list = append(list, toLiveRoomInfo(room, false, svcCtx.Config.Live.RTMPHost, svcCtx.Config.Live.HTTPHost))
		}

		response.Ok(c, gin.H{
			"list":     list,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

func DetailHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的直播间 ID")
			return
		}

		var room model.LiveRoom
		err = svcCtx.DB.Preload("Owner").
			Where("id = ? AND status = ?", id, "live").
			First(&room).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "直播间不存在或已结束")
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播间加载失败")
			return
		}

		response.Ok(c, toLiveRoomInfo(room, false, svcCtx.Config.Live.RTMPHost, svcCtx.Config.Live.HTTPHost))
	}
}

func ManageDetailHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的直播间 ID")
			return
		}
		userID := c.GetUint(middleware.CtxKeyUserID)
		var room model.LiveRoom
		if err := svcCtx.DB.Preload("Owner").Where("id = ? AND owner_id = ? AND status = ?", id, userID, "live").First(&room).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "没有可管理的直播间")
			return
		} else if err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播间加载失败")
			return
		}
		response.Ok(c, toLiveRoomInfo(room, true, svcCtx.Config.Live.RTMPHost, svcCtx.Config.Live.HTTPHost))
	}
}

// UpdateChatSettingsHandler updates the live chat access, rate limit and pinned notice.
func UpdateChatSettingsHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的直播间 ID")
			return
		}
		var req updateChatSettingsReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		if req.ChatMode != "everyone" && req.ChatMode != "followers" && req.ChatMode != "members" {
			response.Fail(c, http.StatusBadRequest, "不支持的聊天模式")
			return
		}
		if req.SlowModeSeconds < 0 || req.SlowModeSeconds > 120 {
			response.Fail(c, http.StatusBadRequest, "慢速模式需要在 0 到 120 秒之间")
			return
		}
		req.PinnedMessage = strings.TrimSpace(req.PinnedMessage)
		if len([]rune(req.PinnedMessage)) > 200 {
			response.Fail(c, http.StatusBadRequest, "置顶公告不能超过 200 个字")
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)
		var room model.LiveRoom
		if err := svcCtx.DB.Preload("Owner").First(&room, uint(id)).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "直播间不存在")
			return
		}
		if room.OwnerID != userID && role != "admin" {
			response.Fail(c, http.StatusForbidden, "无权修改这个直播间")
			return
		}
		if err := svcCtx.DB.Model(&room).Updates(map[string]any{
			"chat_mode": req.ChatMode, "slow_mode_seconds": req.SlowModeSeconds, "pinned_message": req.PinnedMessage,
		}).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "聊天设置保存失败")
			return
		}
		room.ChatMode = req.ChatMode
		room.SlowModeSeconds = req.SlowModeSeconds
		room.PinnedMessage = req.PinnedMessage
		response.Ok(c, toLiveRoomInfo(room, true, svcCtx.Config.Live.RTMPHost, svcCtx.Config.Live.HTTPHost))
	}
}

func ReplayListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := getPage(c)
		db := svcCtx.DB.Model(&model.LiveReplay{}).Preload("Owner")

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播回放加载失败")
			return
		}

		replays := make([]model.LiveReplay, 0, pageSize)
		if err := db.Order("ended_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&replays).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播回放加载失败")
			return
		}

		list := make([]liveReplayInfo, 0, len(replays))
		for i := range replays {
			refreshReplayStatus(svcCtx, &replays[i])
			list = append(list, toLiveReplayInfo(replays[i]))
		}
		response.Ok(c, gin.H{"list": list, "total": total, "page": page, "pageSize": pageSize})
	}
}

func ReplayDetailHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的回放 ID")
			return
		}

		var replay model.LiveReplay
		if err := svcCtx.DB.Preload("Owner").First(&replay, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "直播回放不存在")
			return
		} else if err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播回放加载失败")
			return
		}
		refreshReplayStatus(svcCtx, &replay)
		response.Ok(c, toLiveReplayInfo(replay))
	}
}

func ReplayDanmakuHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的回放 ID")
			return
		}

		var replay model.LiveReplay
		if err := svcCtx.DB.First(&replay, id).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "直播回放不存在")
			return
		} else if err != nil {
			response.Fail(c, http.StatusInternalServerError, "回放弹幕加载失败")
			return
		}

		var danmakus []model.Danmaku
		if err := svcCtx.DB.Where(
			"video_id = ? AND scene = ? AND blocked = ? AND created_at >= ? AND created_at <= ?",
			replay.RoomID, "live", false, replay.StartedAt, replay.EndedAt,
		).Order("created_at ASC").Find(&danmakus).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "回放弹幕加载失败")
			return
		}

		list := make([]replayDanmakuItem, 0, len(danmakus))
		for _, item := range danmakus {
			offset := int(item.CreatedAt.Sub(replay.StartedAt).Seconds())
			if offset < 0 {
				offset = 0
			}
			list = append(list, replayDanmakuItem{
				ID: item.ID, UserID: item.UserID, Content: item.Content, Time: offset,
				Color: item.Color, FontSize: item.FontSize, Type: item.Type,
			})
		}
		response.Ok(c, list)
	}
}

func CreateHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req createLiveRoomReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		title := strings.TrimSpace(req.Title)
		if title == "" {
			response.Fail(c, http.StatusBadRequest, "直播间标题不能为空")
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		if userID == 0 {
			response.Fail(c, http.StatusUnauthorized, "未授权")
			return
		}

		now := time.Now()
		streamKey, err := generateStreamKey()
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播密钥生成失败")
			return
		}

		var room model.LiveRoom
		err = svcCtx.DB.Where("owner_id = ?", userID).First(&room).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			room = model.LiveRoom{
				Title:       title,
				CoverURL:    strings.TrimSpace(req.CoverURL),
				StreamKey:   streamKey,
				Status:      "live",
				ViewerCount: 0,
				ViewerPeak:  0,
				OwnerID:     userID,
				StartedAt:   &now,
				EndedAt:     nil,
			}
			if err := svcCtx.DB.Create(&room).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "直播间创建失败")
				return
			}
		} else if err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播间加载失败")
			return
		} else if room.Status == "live" && room.StreamKey != "" {
			if err := svcCtx.DB.Preload("Owner").First(&room, room.ID).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "live room load failed")
				return
			}
			response.Ok(c, toLiveRoomInfo(room, true, svcCtx.Config.Live.RTMPHost, svcCtx.Config.Live.HTTPHost))
			return
		} else {
			if err := svcCtx.DB.Unscoped().Where("room_id = ?", room.ID).Delete(&model.LiveLike{}).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "直播互动数据重置失败")
				return
			}
			if err := svcCtx.DB.Unscoped().Where("room_id = ?", room.ID).Delete(&model.LiveGift{}).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "直播互动数据重置失败")
				return
			}
			updates := map[string]any{
				"title":        title,
				"cover_url":    strings.TrimSpace(req.CoverURL),
				"stream_key":   streamKey,
				"status":       "live",
				"viewer_count": 0,
				"viewer_peak":  0,
				"like_count":   0,
				"gift_value":   0,
				"started_at":   &now,
				"ended_at":     nil,
			}
			if err := svcCtx.DB.Model(&room).Updates(updates).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "直播间创建失败")
				return
			}
		}

		if err := analyticslogic.AddCreatorDailyStat(svcCtx.DB, userID, 0, 0, 1); err != nil {
			fmt.Printf("record creator stream metric failed: %v\n", err)
		}

		_ = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			if err := notifyScheduleReservations(tx, userID, title); err != nil {
				return err
			}
			return notifyLiveFollowers(tx, userID, "live_start", "你关注的主播开播了", title, "/live")
		})

		if err := svcCtx.DB.Preload("Owner").First(&room, room.ID).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播间加载失败")
			return
		}

		response.Ok(c, toLiveRoomInfo(room, true, svcCtx.Config.Live.RTMPHost, svcCtx.Config.Live.HTTPHost))
	}
}

func EndHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的直播间 ID")
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)

		var room model.LiveRoom
		err = svcCtx.DB.Preload("Owner").First(&room, id).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "直播间不存在")
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播间加载失败")
			return
		}

		if room.OwnerID != userID && role != "admin" {
			response.Fail(c, http.StatusForbidden, "无权结束该直播")
			return
		}

		now := time.Now()
		startedAt := now
		if room.StartedAt != nil {
			startedAt = *room.StartedAt
		}
		duration := int(now.Sub(startedAt).Seconds())
		if duration < 0 {
			duration = 0
		}
		replayURL := fmt.Sprintf("/media/replays/%s.m3u8", room.StreamKey)
		replay := model.LiveReplay{
			RoomID: room.ID, Title: room.Title, CoverURL: room.CoverURL,
			StreamKey: room.StreamKey, ReplayURL: replayURL, Status: "processing",
			Duration: duration, ViewerPeak: room.ViewerPeak, OwnerID: room.OwnerID,
			StartedAt: startedAt, EndedAt: now,
		}
		if err := svcCtx.DB.Where("stream_key = ?", room.StreamKey).FirstOrCreate(&replay).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播回放记录创建失败")
			return
		}

		if err := svcCtx.DB.Model(&room).Updates(map[string]any{
			"status":       "ended",
			"viewer_count": 0,
			"ended_at":     &now,
		}).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "直播结束失败")
			return
		}

		room.Status = "ended"
		room.ViewerCount = 0
		room.EndedAt = &now
		go waitForReplay(svcCtx, replay.ID)
		response.Ok(c, toLiveRoomInfo(room, true, svcCtx.Config.Live.RTMPHost, svcCtx.Config.Live.HTTPHost))
	}
}

func getPage(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return page, pageSize
}

func generateStreamKey() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func toLiveRoomInfo(room model.LiveRoom, includeStreamKey bool, rtmpHost, httpHost string) liveRoomInfo {
	playURL := fmt.Sprintf("http://%s/%s/%s.m3u8", httpHost, liveApp, room.StreamKey)
	info := liveRoomInfo{
		ID:              room.ID,
		Title:           room.Title,
		CoverURL:        room.CoverURL,
		PlayURL:         playURL,
		StreamURL:       playURL,
		Status:          room.Status,
		ViewerCount:     room.ViewerCount,
		ViewerPeak:      room.ViewerPeak,
		LikeCount:       room.LikeCount,
		GiftValue:       room.GiftValue,
		ChatMode:        room.ChatMode,
		SlowModeSeconds: room.SlowModeSeconds,
		PinnedMessage:   room.PinnedMessage,
		Heat:            roomHeat(room),
		OwnerID:         room.OwnerID,
		CreatedAt:       room.CreatedAt.Format("2006-01-02 15:04:05"),
		Owner: &model.UserInfo{
			ID:       room.Owner.ID,
			Username: room.Owner.Username,
			Nickname: room.Owner.Nickname,
			Avatar:   room.Owner.Avatar,
			Role:     room.Owner.Role,
		},
	}
	if room.StartedAt != nil {
		info.StartedAt = room.StartedAt.Format("2006-01-02 15:04:05")
	}
	if room.EndedAt != nil {
		info.EndedAt = room.EndedAt.Format("2006-01-02 15:04:05")
	}
	if includeStreamKey {
		info.StreamKey = room.StreamKey
		info.PublishURL = fmt.Sprintf("rtmp://%s/%s/%s", rtmpHost, liveApp, room.StreamKey)
	}
	return info
}

func toLiveReplayInfo(replay model.LiveReplay) liveReplayInfo {
	info := liveReplayInfo{
		ID: replay.ID, RoomID: replay.RoomID, Title: replay.Title, CoverURL: replay.CoverURL,
		ReplayURL: replay.ReplayURL, Status: replay.Status, Duration: replay.Duration,
		ViewerPeak: replay.ViewerPeak, OwnerID: replay.OwnerID,
		StartedAt: replay.StartedAt.Format("2006-01-02 15:04:05"),
		EndedAt:   replay.EndedAt.Format("2006-01-02 15:04:05"),
		CreatedAt: replay.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if replay.Owner.ID > 0 {
		info.Owner = &model.UserInfo{
			ID: replay.Owner.ID, Username: replay.Owner.Username, Nickname: replay.Owner.Nickname,
			Avatar: replay.Owner.Avatar, Role: replay.Owner.Role,
		}
	}
	return info
}

func waitForReplay(svcCtx *svc.ServiceContext, replayID uint) {
	for attempt := 0; attempt < 15; attempt++ {
		time.Sleep(time.Second)
		var replay model.LiveReplay
		if err := svcCtx.DB.First(&replay, replayID).Error; err != nil {
			return
		}
		if refreshReplayStatus(svcCtx, &replay) {
			return
		}
	}
}

func refreshReplayStatus(svcCtx *svc.ServiceContext, replay *model.LiveReplay) bool {
	if replay.Status == "ready" {
		return true
	}
	playlistPath := filepath.Join(svcCtx.VideoDir, "live", replay.StreamKey+".m3u8")
	info, err := os.Stat(playlistPath)
	if err != nil || info.Size() == 0 || time.Since(info.ModTime()) < 2*time.Second {
		return false
	}
	content, err := os.ReadFile(playlistPath)
	if err != nil || !strings.Contains(string(content), "#EXTM3U") {
		return false
	}
	if !strings.Contains(string(content), "#EXT-X-ENDLIST") {
		content = append(content, []byte("\n#EXT-X-ENDLIST\n")...)
		if err := os.WriteFile(playlistPath, content, 0644); err != nil {
			return false
		}
	}
	replay.Status = "ready"
	svcCtx.DB.Model(replay).Update("status", "ready")
	return true
}
