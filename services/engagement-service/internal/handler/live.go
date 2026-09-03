package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"danmakustream/engagement-service/internal/middleware"
	"danmakustream/engagement-service/internal/model"
	"danmakustream/engagement-service/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (h *Handler) ListLiveRooms(c *gin.Context) {
	p, size := page(c)
	var rows []model.LiveRoom
	var total int64
	h.db.Model(&model.LiveRoom{}).Where("status = ?", "live").Count(&total)
	if err := h.db.Where("status = ?", "live").Order("id desc").Offset((p - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询直播失败")
		return
	}
	response.OK(c, gin.H{"items": rows, "list": rows, "page": p, "pageSize": size, "total": total})
}
func (h *Handler) GetLiveRoom(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var row model.LiveRoom
	if err := h.db.Where("id = ? AND status = ?", id, "live").First(&row).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "直播间不存在")
		return
	} else if err != nil {
		response.Error(c, 500, "查询直播失败")
		return
	}
	response.OK(c, row)
}
func (h *Handler) ManageLiveRoom(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var room model.LiveRoom
	if err := h.db.Where("id = ? AND owner_id = ?", id, middleware.UserID(c)).First(&room).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "直播间不存在或无权管理")
		return
	} else if err != nil {
		response.Error(c, 500, "查询直播间失败")
		return
	}
	response.OK(c, room)
}
func (h *Handler) UpdateChatSettings(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		ChatMode        string `json:"chatMode"`
		SlowModeSeconds int    `json:"slowModeSeconds"`
		PinnedMessage   string `json:"pinnedMessage"`
	}
	if c.ShouldBindJSON(&req) != nil || !map[string]bool{"everyone": true, "followers": true, "members": true}[req.ChatMode] || req.SlowModeSeconds < 0 || req.SlowModeSeconds > 300 || len([]rune(req.PinnedMessage)) > 500 {
		response.Error(c, 400, "聊天设置无效")
		return
	}
	r := h.db.Model(&model.LiveRoom{}).Where("id = ? AND owner_id = ?", id, middleware.UserID(c)).Updates(map[string]any{"chat_mode": req.ChatMode, "slow_mode_seconds": req.SlowModeSeconds, "pinned_message": req.PinnedMessage})
	if r.Error != nil {
		response.Error(c, 500, "更新聊天设置失败")
		return
	}
	if r.RowsAffected == 0 {
		response.Error(c, 404, "直播间不存在或无权管理")
		return
	}
	var room model.LiveRoom
	h.db.First(&room, id)
	response.OK(c, room)
}
func (h *Handler) LiveInteraction(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var room model.LiveRoom
	if err := h.db.Where("id = ? AND status = ?", id, "live").First(&room).Error; err != nil {
		response.Error(c, 404, "直播间不存在")
		return
	}
	response.OK(c, gin.H{
		"viewerCount": room.ViewerCount, "viewerPeak": room.ViewerPeak,
		"likeCount": room.LikeCount, "giftValue": room.GiftValue,
		"heat":  room.ViewerCount*10 + room.LikeCount*2 + room.GiftValue,
		"gifts": giftDefinitions(), "supportRank": []any{}, "superChats": []any{},
		"chatSettings": gin.H{"chatMode": room.ChatMode, "slowModeSeconds": room.SlowModeSeconds, "pinnedMessage": room.PinnedMessage},
	})
}
func (h *Handler) CurrentLiveDanmaku(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var rows []model.Danmaku
	if err := h.db.Where("video_id = ? AND scene = ? AND blocked = ?", id, "live", false).Order("id desc").Limit(100).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询直播弹幕失败")
		return
	}
	response.OK(c, rows)
}
func (h *Handler) HeatRanking(c *gin.Context) {
	var rooms []model.LiveRoom
	if err := h.db.Where("status = ?", "live").Find(&rooms).Error; err != nil {
		response.Error(c, 500, "查询热度榜失败")
		return
	}
	items := make([]gin.H, 0, len(rooms))
	for _, room := range rooms {
		items = append(items, gin.H{"room": room, "heat": room.ViewerCount*10 + room.LikeCount*2 + room.GiftValue})
	}
	response.OK(c, items)
}
func (h *Handler) LiveLikeStatus(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var count int64
	if err := h.db.Model(&model.LiveLike{}).Where("room_id = ? AND user_id = ?", id, middleware.UserID(c)).Count(&count).Error; err != nil {
		response.Error(c, 500, "查询点赞状态失败")
		return
	}
	response.OK(c, gin.H{"liked": count > 0})
}
func (h *Handler) MonitorLiveRoom(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var room model.LiveRoom
	if err := h.db.Where("id = ? AND owner_id = ? AND status = ?", id, middleware.UserID(c), "live").First(&room).Error; err != nil {
		response.Error(c, 404, "直播间不存在或无权监控")
		return
	}
	var chats []model.Danmaku
	h.db.Where("video_id = ? AND scene = ? AND blocked = ?", id, "live", false).Order("id desc").Limit(50).Find(&chats)
	var gifts []model.LiveGift
	h.db.Where("room_id = ?", id).Order("id desc").Limit(20).Find(&gifts)
	response.OK(c, gin.H{"room": room, "recentDanmaku": chats, "recentGifts": gifts, "heat": room.ViewerCount*10 + room.LikeCount*2 + room.GiftValue})
}
func (h *Handler) CreateLiveRoom(c *gin.Context) {
	var req struct {
		Title    string `json:"title"`
		CoverURL string `json:"coverUrl"`
	}
	if c.ShouldBindJSON(&req) != nil || strings.TrimSpace(req.Title) == "" || len([]rune(req.Title)) > 200 {
		response.Error(c, 400, "直播标题无效")
		return
	}
	uid := middleware.UserID(c)
	now := time.Now()
	var row model.LiveRoom
	err := h.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Where("owner_id = ?", uid).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			row = model.LiveRoom{Title: strings.TrimSpace(req.Title), CoverURL: req.CoverURL, StreamKey: streamKey(), Status: "live", OwnerID: uid, StartedAt: &now}
			return tx.Create(&row).Error
		}
		if err != nil {
			return err
		}
		if err := tx.Model(&row).Updates(map[string]any{"title": strings.TrimSpace(req.Title), "cover_url": req.CoverURL, "stream_key": streamKey(), "status": "live", "started_at": &now, "ended_at": nil, "viewer_count": 0, "viewer_peak": 0, "like_count": 0, "gift_value": 0}).Error; err != nil {
			return err
		}
		// Reopening an owner's single reusable room starts a fresh session.
		// Session-scoped likes/gifts/chat must not leak into the new broadcast.
		if err := tx.Unscoped().Where("room_id = ?", row.ID).Delete(&model.LiveLike{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("room_id = ?", row.ID).Delete(&model.LiveGift{}).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Where("room_id = ?", row.ID).Delete(&model.SuperChat{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Where("video_id = ? AND scene = ?", row.ID, "live").Delete(&model.Danmaku{}).Error
	})
	if err != nil {
		response.Error(c, 500, "创建直播失败")
		return
	}
	h.db.First(&row, row.ID)
	response.OK(c, row)
}
func (h *Handler) EndLiveRoom(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	now := time.Now()
	r := h.db.Model(&model.LiveRoom{}).Where("id = ? AND owner_id = ? AND status = ?", id, middleware.UserID(c), "live").Updates(map[string]any{"status": "ended", "ended_at": &now, "viewer_count": 0})
	if r.Error != nil {
		response.Error(c, 500, "结束直播失败")
		return
	}
	if r.RowsAffected == 0 {
		response.Error(c, 404, "直播间不存在或无权操作")
		return
	}
	var room model.LiveRoom
	if err := h.db.First(&room, id).Error; err != nil {
		response.Error(c, 500, "读取结束状态失败")
		return
	}
	response.OK(c, room)
}
func (h *Handler) ToggleLiveLike(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserID(c)
	liked := false
	var count int64
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var room model.LiveRoom
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", id, "live").First(&room).Error; err != nil {
			return err
		}
		var row model.LiveLike
		err := tx.Where("room_id = ? AND user_id = ?", id, uid).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&model.LiveLike{RoomID: id, UserID: uid}).Error; err != nil {
				return err
			}
			liked = true
		} else if err == nil {
			if err := tx.Unscoped().Delete(&row).Error; err != nil {
				return err
			}
		} else {
			return err
		}
		if err := tx.Model(&model.LiveLike{}).Where("room_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		return tx.Model(&room).Update("like_count", count).Error
	})
	if err != nil {
		response.Error(c, 500, "直播点赞失败")
		return
	}
	var room model.LiveRoom
	_ = h.db.First(&room, id).Error
	heat := room.ViewerCount*10 + count*2 + room.GiftValue
	payload := gin.H{"userId": uid, "liked": liked, "likeCount": count, "heat": heat}
	h.hub.Broadcast(id, gin.H{"type": "live_like", "payload": payload})
	response.OK(c, payload)
}

var gifts = map[string]struct {
	Name  string
	Value int64
}{"flower": {"鲜花", 10}, "star": {"星光", 50}, "rocket": {"火箭", 200}}

func giftDefinitions() []gin.H {
	return []gin.H{
		{"key": "flower", "name": "鲜花", "value": 10},
		{"key": "star", "name": "星光", "value": 50},
		{"key": "rocket", "name": "火箭", "value": 200},
	}
}

func (h *Handler) SendGift(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		GiftKey string `json:"giftKey"`
		Count   int    `json:"count"`
		Message string `json:"message"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Count < 1 || req.Count > 99 || len([]rune(req.Message)) > 200 {
		response.Error(c, 400, "礼物参数无效")
		return
	}
	gift, ok := gifts[req.GiftKey]
	if !ok {
		response.Error(c, 400, "礼物不存在")
		return
	}
	value := gift.Value * int64(req.Count)
	record := model.LiveGift{RoomID: id, UserID: middleware.UserID(c), GiftKey: req.GiftKey, Name: gift.Name, Count: req.Count, Value: value, Message: strings.TrimSpace(req.Message)}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var room model.LiveRoom
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", id, "live").First(&room).Error; err != nil {
			return err
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		if record.Message != "" {
			if err := tx.Create(&model.SuperChat{RoomID: id, UserID: record.UserID, GiftID: record.ID, Content: record.Message, DisplaySeconds: superChatSeconds(value)}).Error; err != nil {
				return err
			}
		}
		return tx.Model(&room).Update("gift_value", gorm.Expr("gift_value + ?", value)).Error
	})
	if err != nil {
		response.Error(c, 500, "赠送礼物失败")
		return
	}
	var room model.LiveRoom
	_ = h.db.First(&room, id).Error
	payload := gin.H{
		"id": record.ID, "userId": record.UserID,
		"gift":  gin.H{"key": record.GiftKey, "name": record.Name, "value": gift.Value},
		"count": record.Count, "value": record.Value, "message": record.Message,
		"giftValue": room.GiftValue, "heat": room.ViewerCount*10 + room.LikeCount*2 + room.GiftValue,
		"supportRank": []any{}, "createdAt": record.CreatedAt,
		"displaySeconds": superChatSeconds(record.Value),
	}
	h.hub.Broadcast(id, gin.H{"type": "live_gift", "payload": payload})
	response.OK(c, payload)
}
func superChatSeconds(value int64) int {
	switch {
	case value >= 1000:
		return 120
	case value >= 500:
		return 90
	case value >= 200:
		return 60
	case value >= 50:
		return 30
	default:
		return 15
	}
}

func (h *Handler) ListSchedules(c *gin.Context) {
	p, size := page(c)
	var rows []model.LiveSchedule
	q := h.db.Order("scheduled_at asc")
	countQ := h.db.Model(&model.LiveSchedule{})
	if status := c.Query("status"); status != "" {
		q = q.Where("status = ?", status)
		countQ = countQ.Where("status = ?", status)
	}
	var total int64
	countQ.Count(&total)
	if err := q.Offset((p - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询直播计划失败")
		return
	}
	type scheduleInfo struct {
		model.LiveSchedule
		Reserved bool `json:"reserved"`
	}
	items := make([]scheduleInfo, 0, len(rows))
	userID := middleware.UserID(c)
	reserved := map[uint]bool{}
	if userID > 0 && len(rows) > 0 {
		ids := make([]uint, 0, len(rows))
		for _, row := range rows {
			ids = append(ids, row.ID)
		}
		var reservations []model.LiveReservation
		if err := h.db.Where("user_id = ? AND schedule_id IN ?", userID, ids).Find(&reservations).Error; err != nil {
			response.Error(c, 500, "查询预约状态失败")
			return
		}
		for _, item := range reservations {
			reserved[item.ScheduleID] = true
		}
	}
	for _, row := range rows {
		items = append(items, scheduleInfo{LiveSchedule: row, Reserved: reserved[row.ID]})
	}
	response.OK(c, gin.H{"items": items, "list": items, "page": p, "pageSize": size, "total": total})
}
func (h *Handler) CreateSchedule(c *gin.Context) {
	var req struct {
		Title       string `json:"title"`
		CoverURL    string `json:"coverUrl"`
		ScheduledAt string `json:"scheduledAt"`
	}
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		response.Error(c, 400, "参数无效")
		return
	}
	at, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil || !at.After(time.Now()) || strings.TrimSpace(req.Title) == "" {
		response.Error(c, 400, "直播时间或标题无效")
		return
	}
	uid := middleware.UserID(c)
	var count int64
	h.db.Model(&model.LiveSchedule{}).Where("owner_id = ? AND scheduled_at = ? AND status = ?", uid, at, "pending").Count(&count)
	if count > 0 {
		response.Error(c, 409, "相同时间的直播计划已存在")
		return
	}
	row := model.LiveSchedule{Title: strings.TrimSpace(req.Title), CoverURL: req.CoverURL, ScheduledAt: at, Status: "pending", OwnerID: uid}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 500, "创建直播计划失败")
		return
	}
	response.OK(c, row)
}
func (h *Handler) CancelSchedule(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	r := h.db.Model(&model.LiveSchedule{}).Where("id = ? AND owner_id = ? AND status = ?", id, middleware.UserID(c), "pending").Update("status", "canceled")
	if r.Error != nil {
		response.Error(c, 500, "取消失败")
		return
	}
	if r.RowsAffected == 0 {
		response.Error(c, 404, "计划不存在或无权取消")
		return
	}
	response.OK(c, gin.H{"canceled": true})
}
func (h *Handler) ToggleReservation(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserID(c)
	reserved := false
	var count int64
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var schedule model.LiveSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND status = ?", id, "pending").First(&schedule).Error; err != nil {
			return err
		}
		if schedule.OwnerID == uid {
			return fmt.Errorf("cannot reserve own schedule")
		}
		var row model.LiveReservation
		err := tx.Where("schedule_id = ? AND user_id = ?", id, uid).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&model.LiveReservation{ScheduleID: id, UserID: uid}).Error; err != nil {
				return err
			}
			reserved = true
		} else if err == nil {
			if err := tx.Unscoped().Delete(&row).Error; err != nil {
				return err
			}
		} else {
			return err
		}
		if err := tx.Model(&model.LiveReservation{}).Where("schedule_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		return tx.Model(&schedule).Update("reminder_count", count).Error
	})
	if err != nil {
		if strings.Contains(err.Error(), "own schedule") {
			response.Error(c, 409, "不能预约自己的直播")
			return
		}
		response.Error(c, 500, "预约操作失败")
		return
	}
	response.OK(c, gin.H{"reserved": reserved, "reminderCount": count})
}

func (h *Handler) SRSHook(c *gin.Context) {
	var req struct {
		Action string `json:"action"`
		Stream string `json:"stream"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Stream == "" {
		c.JSON(200, gin.H{"code": 0})
		return
	}
	updates := map[string]any{}
	switch req.Action {
	case "on_publish":
		updates["status"] = "live"
		updates["started_at"] = time.Now()
	case "on_unpublish":
		updates["status"] = "ended"
		updates["ended_at"] = time.Now()
		updates["viewer_count"] = 0
	default:
		c.JSON(200, gin.H{"code": 0})
		return
	}
	r := h.db.Model(&model.LiveRoom{}).Where("stream_key = ?", req.Stream).Updates(updates)
	if r.Error != nil || r.RowsAffected == 0 {
		c.JSON(200, gin.H{"code": 1})
		return
	}
	c.JSON(200, gin.H{"code": 0})
}

var publishers = struct {
	sync.Mutex
	rooms map[uint]bool
}{rooms: map[uint]bool{}}

func (h *Handler) LivePublishWebSocket(c *gin.Context) {
	roomID, ok := idParam(c, "id")
	if !ok {
		return
	}
	uid, ok := h.wsUserID(c)
	if !ok {
		response.Error(c, 401, "未授权")
		return
	}
	var room model.LiveRoom
	if err := h.db.Where("id = ? AND owner_id = ? AND status = ?", roomID, uid, "live").First(&room).Error; err != nil {
		response.Error(c, 403, "直播间不可管理")
		return
	}
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		response.Error(c, 503, "ffmpeg 不可用")
		return
	}
	publishers.Lock()
	if publishers.rooms[roomID] {
		publishers.Unlock()
		response.Error(c, 409, "直播间已有浏览器推流")
		return
	}
	publishers.rooms[roomID] = true
	publishers.Unlock()
	defer func() { publishers.Lock(); delete(publishers.rooms, roomID); publishers.Unlock() }()
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()
	target := fmt.Sprintf("rtmp://%s/live/%s", defaultHost(h.cfg.Live.RTMPHost, "srs:1935"), room.StreamKey)
	cmd := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-loglevel", "warning", "-f", "webm", "-i", "pipe:0", "-map", "0:v:0", "-map", "0:a?", "-c:v", "libx264", "-preset", "ultrafast", "-tune", "zerolatency", "-pix_fmt", "yuv420p", "-c:a", "aac", "-f", "flv", target)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	cmd.Stdout = io.Discard
	cmd.Stderr = os.Stderr
	if cmd.Start() != nil {
		return
	}
	defer func() { _ = stdin.Close(); cancel(); _ = cmd.Wait() }()
	for {
		kind, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if kind == websocket.BinaryMessage && len(data) > 0 {
			if _, err := stdin.Write(data); err != nil {
				return
			}
		}
	}
}
func (h *Handler) wsUserID(c *gin.Context) (uint, bool) {
	raw := c.Query("token")
	if raw == "" {
		raw = strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	}
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(h.cfg.Auth.AccessSecret), nil
	})
	return claims.UserID, err == nil && token.Valid && claims.UserID > 0
}
func defaultHost(value, fallback string) string {
	if value == "" {
		return fallback
	}
	if strings.Contains(value, "://") {
		return fallback
	}
	return value
}

// StartScheduleWorker promotes due plans using engagement-owned tables only.
// Cross-service notifications are intentionally deferred to the integration
// step so a user-service outage cannot roll back the owned state transition.
func StartScheduleWorker(db *gorm.DB) {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var schedules []model.LiveSchedule
			if err := db.Where("status = ? AND scheduled_at <= ?", "pending", time.Now()).Order("scheduled_at asc").Limit(20).Find(&schedules).Error; err != nil {
				continue
			}
			for _, schedule := range schedules {
				_ = db.Transaction(func(tx *gorm.DB) error {
					result := tx.Model(&model.LiveSchedule{}).Where("id = ? AND status = ?", schedule.ID, "pending").Update("status", "live")
					if result.Error != nil || result.RowsAffected == 0 {
						return result.Error
					}
					var room model.LiveRoom
					err := tx.Where("owner_id = ?", schedule.OwnerID).First(&room).Error
					now := time.Now()
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return tx.Create(&model.LiveRoom{Title: schedule.Title, CoverURL: schedule.CoverURL, StreamKey: streamKey(), Status: "live", OwnerID: schedule.OwnerID, StartedAt: &now}).Error
					}
					if err != nil {
						return err
					}
					return tx.Model(&room).Updates(map[string]any{"title": schedule.Title, "cover_url": schedule.CoverURL, "stream_key": streamKey(), "status": "live", "started_at": &now, "ended_at": nil}).Error
				})
			}
		}
	}()
}
