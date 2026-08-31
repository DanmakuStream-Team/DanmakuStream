package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"danmakustream/engagement-service/internal/client"
	"danmakustream/engagement-service/internal/config"
	"danmakustream/engagement-service/internal/middleware"
	"danmakustream/engagement-service/internal/model"
	"danmakustream/engagement-service/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Handler struct {
	db             *gorm.DB
	users, content *client.Client
	hub            *Hub
	cfg            config.Config
}

func New(db *gorm.DB, users, content *client.Client, hub *Hub, cfg config.Config) *Handler {
	return &Handler{db: db, users: users, content: content, hub: hub, cfg: cfg}
}

func idParam(c *gin.Context, name string) (uint, bool) {
	raw, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil || raw == 0 {
		response.Error(c, http.StatusBadRequest, "invalid "+name)
		return 0, false
	}
	return uint(raw), true
}
func page(c *gin.Context) (int, int) {
	p, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if p < 1 {
		p = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return p, size
}
func dependencyError(c *gin.Context, err error, object string) {
	if errors.Is(err, client.ErrNotFound) {
		response.Error(c, http.StatusNotFound, object+"不存在")
		return
	}
	if errors.Is(err, client.ErrTimeout) {
		response.Error(c, http.StatusGatewayTimeout, "依赖服务调用超时")
		return
	}
	if errors.Is(err, client.ErrBadGateway) {
		response.Error(c, http.StatusBadGateway, "依赖服务响应无效")
		return
	}
	response.Error(c, http.StatusServiceUnavailable, "依赖服务暂不可用")
}
func (h *Handler) requirePlayable(c *gin.Context, id uint) (client.VideoSummary, bool) {
	video, err := h.content.Video(c.Request.Context(), id)
	if err != nil {
		dependencyError(c, err, "视频")
		return video, false
	}
	if !video.Playable && video.Status != "approved" {
		response.Error(c, http.StatusConflict, "视频当前不可互动")
		return video, false
	}
	return video, true
}

func (h *Handler) ListDanmaku(c *gin.Context) {
	videoID, ok := idParam(c, "videoId")
	if !ok {
		return
	}
	var rows []model.Danmaku
	if err := h.db.Where("video_id = ? AND scene = ? AND blocked = ?", videoID, "video", false).Order("time asc,id asc").Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询弹幕失败")
		return
	}
	response.OK(c, rows)
}
func (h *Handler) SendDanmaku(c *gin.Context) {
	var req struct {
		VideoID  uint   `json:"videoId"`
		Content  string `json:"content"`
		Time     int    `json:"time"`
		Color    string `json:"color"`
		FontSize string `json:"fontSize"`
		Type     string `json:"type"`
	}
	if c.ShouldBindJSON(&req) != nil || req.VideoID == 0 || strings.TrimSpace(req.Content) == "" || len([]rune(req.Content)) > 200 || req.Time < 0 {
		response.Error(c, 400, "弹幕参数无效")
		return
	}
	if _, ok := h.requirePlayable(c, req.VideoID); !ok {
		return
	}
	row := model.Danmaku{VideoID: req.VideoID, Scene: "video", UserID: middleware.UserID(c), Content: strings.TrimSpace(req.Content), Time: req.Time, Color: req.Color, FontSize: req.FontSize, Type: req.Type}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 500, "保存弹幕失败")
		return
	}
	response.OK(c, row)
}
func (h *Handler) AdminListDanmaku(c *gin.Context) {
	p, size := page(c)
	var rows []model.Danmaku
	q := h.db.Order("id desc")
	countQ := h.db.Model(&model.Danmaku{})
	if blocked := c.Query("blocked"); blocked != "" {
		q = q.Where("blocked = ?", blocked == "true")
		countQ = countQ.Where("blocked = ?", blocked == "true")
	}
	var total int64
	countQ.Count(&total)
	if err := q.Offset((p - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询弹幕失败")
		return
	}
	response.OK(c, gin.H{"items": rows, "list": rows, "page": p, "pageSize": size, "total": total})
}
func (h *Handler) BlockDanmaku(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	var req struct {
		Blocked bool `json:"blocked"`
	}
	if c.ShouldBindJSON(&req) != nil {
		response.Error(c, 400, "参数无效")
		return
	}
	r := h.db.Model(&model.Danmaku{}).Where("id = ?", id).Update("blocked", req.Blocked)
	if r.Error != nil {
		response.Error(c, 500, "更新失败")
		return
	}
	if r.RowsAffected == 0 {
		response.Error(c, 404, "弹幕不存在")
		return
	}
	response.OK(c, gin.H{"id": id, "blocked": req.Blocked})
}

func (h *Handler) ListComments(c *gin.Context) {
	videoID, ok := idParam(c, "videoId")
	if !ok {
		return
	}
	p, size := page(c)
	var rows []model.Comment
	if err := h.db.Where("video_id = ?", videoID).Order("id desc").Offset((p - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询评论失败")
		return
	}
	response.OK(c, gin.H{"list": rows, "page": p, "pageSize": size})
}
func (h *Handler) CreateComment(c *gin.Context) {
	var req struct {
		VideoID  uint   `json:"videoId"`
		ParentID *uint  `json:"parentId"`
		Content  string `json:"content"`
	}
	if c.ShouldBindJSON(&req) != nil || req.VideoID == 0 || strings.TrimSpace(req.Content) == "" || len([]rune(req.Content)) > 1000 {
		response.Error(c, 400, "评论参数无效")
		return
	}
	if _, ok := h.requirePlayable(c, req.VideoID); !ok {
		return
	}
	if req.ParentID != nil {
		var count int64
		h.db.Model(&model.Comment{}).Where("id = ? AND video_id = ?", *req.ParentID, req.VideoID).Count(&count)
		if count == 0 {
			response.Error(c, 404, "父评论不存在")
			return
		}
	}
	row := model.Comment{VideoID: req.VideoID, UserID: middleware.UserID(c), ParentID: req.ParentID, Content: strings.TrimSpace(req.Content)}
	if err := h.db.Create(&row).Error; err != nil {
		response.Error(c, 500, "保存评论失败")
		return
	}
	response.OK(c, row)
}
func (h *Handler) DeleteComment(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	r := h.db.Where("id = ? AND user_id = ?", id, middleware.UserID(c)).Delete(&model.Comment{})
	if r.Error != nil {
		response.Error(c, 500, "删除失败")
		return
	}
	if r.RowsAffected == 0 {
		response.Error(c, 404, "评论不存在或无权删除")
		return
	}
	response.OK(c, gin.H{"deleted": true})
}
func (h *Handler) ToggleCommentLike(c *gin.Context) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	uid := middleware.UserID(c)
	liked := false
	err := h.db.Transaction(func(tx *gorm.DB) error {
		var comment model.Comment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&comment, id).Error; err != nil {
			return err
		}
		var like model.CommentLike
		err := tx.Where("user_id = ? AND comment_id = ?", uid, id).First(&like).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := tx.Create(&model.CommentLike{UserID: uid, CommentID: id}).Error; err != nil {
				return err
			}
			liked = true
		} else if err == nil {
			if err := tx.Unscoped().Delete(&like).Error; err != nil {
				return err
			}
		} else {
			return err
		}
		var count int64
		if err := tx.Model(&model.CommentLike{}).Where("comment_id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		return tx.Model(&comment).Update("like_count", count).Error
	})
	if err != nil {
		response.Error(c, 500, "点赞失败")
		return
	}
	response.OK(c, gin.H{"liked": liked})
}

func (h *Handler) toggleVideoRelation(c *gin.Context, collect bool) {
	id, ok := idParam(c, "id")
	if !ok {
		return
	}
	if _, ok := h.requirePlayable(c, id); !ok {
		return
	}
	uid := middleware.UserID(c)
	active := false
	var err error
	if collect {
		var row model.VideoCollection
		err = h.db.Where("user_id = ? AND video_id = ?", uid, id).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = h.db.Create(&model.VideoCollection{UserID: uid, VideoID: id}).Error
			active = err == nil
		} else if err == nil {
			err = h.db.Unscoped().Delete(&row).Error
		}
	} else {
		var row model.VideoLike
		err = h.db.Where("user_id = ? AND video_id = ?", uid, id).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = h.db.Create(&model.VideoLike{UserID: uid, VideoID: id}).Error
			active = err == nil
		} else if err == nil {
			err = h.db.Unscoped().Delete(&row).Error
		}
	}
	if err != nil {
		response.Error(c, 500, "更新互动状态失败")
		return
	}
	if collect {
		response.OK(c, gin.H{"collected": active})
		return
	}
	response.OK(c, gin.H{"liked": active})
}
func (h *Handler) ToggleVideoLike(c *gin.Context)       { h.toggleVideoRelation(c, false) }
func (h *Handler) ToggleVideoCollection(c *gin.Context) { h.toggleVideoRelation(c, true) }

func (h *Handler) ListHistory(c *gin.Context) {
	p, size := page(c)
	var rows []model.WatchHistory
	var total int64
	h.db.Model(&model.WatchHistory{}).Where("user_id = ?", middleware.UserID(c)).Count(&total)
	if err := h.db.Where("user_id = ?", middleware.UserID(c)).Order("updated_at desc").Offset((p - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询历史失败")
		return
	}
	items, err := h.historyRecords(c, rows)
	if err != nil {
		return
	}
	response.OK(c, gin.H{"items": items, "list": items, "page": p, "pageSize": size, "total": total})
}

type libraryRecord struct {
	Video    client.VideoSummary `json:"video"`
	SavedAt  time.Time           `json:"savedAt"`
	Progress int                 `json:"progress"`
	Position int                 `json:"position"`
}

func (h *Handler) historyRecords(c *gin.Context, rows []model.WatchHistory) ([]libraryRecord, error) {
	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.VideoID
	}
	videos, err := h.content.Videos(c.Request.Context(), ids)
	if err != nil {
		dependencyError(c, err, "视频")
		return nil, err
	}
	byID := map[uint]client.VideoSummary{}
	for _, video := range videos {
		byID[video.ID] = video
	}
	items := make([]libraryRecord, 0, len(rows))
	for _, row := range rows {
		video, ok := byID[row.VideoID]
		if !ok {
			continue
		}
		progress := 0
		if video.Duration > 0 {
			progress = row.Position * 100 / video.Duration
			if progress > 100 {
				progress = 100
			}
		}
		items = append(items, libraryRecord{Video: video, SavedAt: row.UpdatedAt, Progress: progress, Position: row.Position})
	}
	return items, nil
}
func (h *Handler) GetHistory(c *gin.Context) {
	videoID, ok := idParam(c, "videoId")
	if !ok {
		return
	}
	var row model.WatchHistory
	if err := h.db.Where("user_id = ? AND video_id = ?", middleware.UserID(c), videoID).First(&row).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		response.Error(c, 404, "观看记录不存在")
		return
	} else if err != nil {
		response.Error(c, 500, "查询历史失败")
		return
	}
	items, err := h.historyRecords(c, []model.WatchHistory{row})
	if err != nil {
		return
	}
	if len(items) == 0 {
		response.Error(c, 404, "视频不存在")
		return
	}
	response.OK(c, items[0])
}
func (h *Handler) SaveHistory(c *gin.Context) {
	videoID, ok := idParam(c, "videoId")
	if !ok {
		return
	}
	var req struct {
		Position int `json:"position"`
	}
	if c.ShouldBindJSON(&req) != nil || req.Position < 0 {
		response.Error(c, 400, "播放进度无效")
		return
	}
	video, ok := h.requirePlayable(c, videoID)
	if !ok {
		return
	}
	if video.Duration > 0 && req.Position > video.Duration {
		req.Position = video.Duration
	}
	row := model.WatchHistory{UserID: middleware.UserID(c), VideoID: videoID, Position: req.Position}
	if err := h.db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}, {Name: "video_id"}}, DoUpdates: clause.Assignments(map[string]any{"position": req.Position, "updated_at": time.Now()})}).Create(&row).Error; err != nil {
		response.Error(c, 500, "保存进度失败")
		return
	}
	response.OK(c, gin.H{"videoId": videoID, "position": req.Position})
}
func (h *Handler) DeleteHistory(c *gin.Context) {
	videoID, ok := idParam(c, "videoId")
	if !ok {
		return
	}
	if err := h.db.Unscoped().Where("user_id = ? AND video_id = ?", middleware.UserID(c), videoID).Delete(&model.WatchHistory{}).Error; err != nil {
		response.Error(c, 500, "删除历史失败")
		return
	}
	response.OK(c, gin.H{"videoId": videoID})
}
func (h *Handler) ClearHistory(c *gin.Context) {
	if err := h.db.Unscoped().Where("user_id = ?", middleware.UserID(c)).Delete(&model.WatchHistory{}).Error; err != nil {
		response.Error(c, 500, "清空历史失败")
		return
	}
	response.OK(c, gin.H{"cleared": true})
}
func (h *Handler) ListWatchLater(c *gin.Context) {
	p, size := page(c)
	var rows []model.WatchLater
	var total int64
	h.db.Model(&model.WatchLater{}).Where("user_id = ?", middleware.UserID(c)).Count(&total)
	if err := h.db.Where("user_id = ?", middleware.UserID(c)).Order("created_at desc").Offset((p - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询稍后再看失败")
		return
	}
	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.VideoID
	}
	videos, err := h.content.Videos(c.Request.Context(), ids)
	if err != nil {
		dependencyError(c, err, "视频")
		return
	}
	byID := map[uint]client.VideoSummary{}
	for _, video := range videos {
		byID[video.ID] = video
	}
	items := make([]libraryRecord, 0, len(rows))
	for _, row := range rows {
		if video, ok := byID[row.VideoID]; ok {
			items = append(items, libraryRecord{Video: video, SavedAt: row.CreatedAt})
		}
	}
	response.OK(c, gin.H{"items": items, "list": items, "page": p, "pageSize": size, "total": total})
}
func (h *Handler) WatchLaterStatus(c *gin.Context) {
	videoID, ok := idParam(c, "videoId")
	if !ok {
		return
	}
	var count int64
	if err := h.db.Model(&model.WatchLater{}).Where("user_id = ? AND video_id = ?", middleware.UserID(c), videoID).Count(&count).Error; err != nil {
		response.Error(c, 500, "查询稍后再看失败")
		return
	}
	response.OK(c, gin.H{"saved": count > 0})
}
func (h *Handler) ToggleWatchLater(c *gin.Context) {
	videoID, ok := idParam(c, "videoId")
	if !ok {
		return
	}
	if _, ok := h.requirePlayable(c, videoID); !ok {
		return
	}
	uid := middleware.UserID(c)
	var row model.WatchLater
	err := h.db.Where("user_id = ? AND video_id = ?", uid, videoID).First(&row).Error
	active := false
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = h.db.Create(&model.WatchLater{UserID: uid, VideoID: videoID}).Error
		active = err == nil
	} else if err == nil {
		err = h.db.Unscoped().Delete(&row).Error
	}
	if err != nil {
		response.Error(c, 500, "更新稍后再看失败")
		return
	}
	response.OK(c, gin.H{"saved": active})
}
func (h *Handler) DeleteWatchLater(c *gin.Context) {
	videoID, ok := idParam(c, "videoId")
	if !ok {
		return
	}
	if err := h.db.Unscoped().Where("user_id = ? AND video_id = ?", middleware.UserID(c), videoID).Delete(&model.WatchLater{}).Error; err != nil {
		response.Error(c, 500, "删除稍后再看失败")
		return
	}
	response.OK(c, gin.H{"videoId": videoID})
}
func (h *Handler) ClearWatchLater(c *gin.Context) {
	if err := h.db.Unscoped().Where("user_id = ?", middleware.UserID(c)).Delete(&model.WatchLater{}).Error; err != nil {
		response.Error(c, 500, "清空稍后再看失败")
		return
	}
	response.OK(c, gin.H{"cleared": true})
}
func (h *Handler) ListVideoCollections(c *gin.Context) {
	p, size := page(c)
	uid := middleware.UserID(c)
	var rows []model.VideoCollection
	var total int64
	h.db.Model(&model.VideoCollection{}).Where("user_id = ?", uid).Count(&total)
	if err := h.db.Where("user_id = ?", uid).Order("id desc").Offset((p - 1) * size).Limit(size).Find(&rows).Error; err != nil {
		response.Error(c, 500, "查询收藏失败")
		return
	}
	ids := make([]uint, len(rows))
	for i, row := range rows {
		ids[i] = row.VideoID
	}
	videos, err := h.content.Videos(c.Request.Context(), ids)
	if err != nil {
		dependencyError(c, err, "视频")
		return
	}
	response.OK(c, gin.H{"items": videos, "list": videos, "page": p, "pageSize": size, "total": total})
}

var wsUpgrader = websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}

func (h *Handler) LiveWebSocket(c *gin.Context) {
	roomID, ok := idParam(c, "id")
	if !ok {
		return
	}
	var room model.LiveRoom
	if h.db.Where("id = ? AND status = ?", roomID, "live").First(&room).Error != nil {
		response.Error(c, 404, "直播间不存在")
		return
	}
	uid, authenticated := h.wsUserID(c)
	monitor := c.Query("monitor") == "1"
	if monitor && (!authenticated || uid != room.OwnerID) {
		response.Error(c, 403, "仅主播可连接监控")
		return
	}
	if !monitor && room.ChatMode == "members" && uid != room.OwnerID {
		active, err := h.users.IsMember(c.Request.Context(), uid, room.OwnerID)
		if err != nil {
			dependencyError(c, err, "会员状态")
			return
		}
		if !active {
			response.Error(c, 403, "仅会员可参与直播聊天")
			return
		}
	}
	if !monitor && room.ChatMode == "followers" && uid != room.OwnerID {
		following, err := h.users.IsFollowing(c.Request.Context(), uid, room.OwnerID)
		if err != nil {
			dependencyError(c, err, "关注关系")
			return
		}
		if !following {
			response.Error(c, 403, "仅关注者可参与直播聊天")
			return
		}
	}
	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	viewers := h.hub.Add(roomID, conn, uid, !monitor)
	h.db.Model(&model.LiveRoom{}).Where("id = ?", roomID).Updates(map[string]any{"viewer_count": viewers, "viewer_peak": gorm.Expr("GREATEST(viewer_peak, ?)", viewers)})
	h.hub.Broadcast(roomID, gin.H{"type": "viewer_count", "payload": viewers})
	defer func() {
		left := h.hub.Remove(roomID, conn)
		h.db.Model(&model.LiveRoom{}).Where("id = ?", roomID).Update("viewer_count", left)
		_ = conn.Close()
		h.hub.Broadcast(roomID, gin.H{"type": "viewer_count", "payload": left})
	}()
	for {
		var msg struct {
			Type    string `json:"type"`
			Content string `json:"content"`
			Time    int    `json:"time"`
		}
		if conn.ReadJSON(&msg) != nil {
			return
		}
		if msg.Type != "danmaku" || strings.TrimSpace(msg.Content) == "" || len([]rune(msg.Content)) > 200 {
			h.hub.Write(conn, gin.H{"type": "chat_error", "payload": gin.H{"message": "消息无效", "retryAfter": 0}})
			continue
		}
		if monitor {
			h.hub.Write(conn, gin.H{"type": "chat_error", "payload": gin.H{"message": "监控连接不能发送弹幕", "retryAfter": 0}})
			continue
		}
		row := model.Danmaku{VideoID: roomID, Scene: "live", UserID: uid, Content: strings.TrimSpace(msg.Content), Time: msg.Time}
		if err := h.db.Create(&row).Error; err != nil {
			h.hub.Write(conn, gin.H{"type": "chat_error", "payload": gin.H{"message": "保存失败", "retryAfter": 0}})
			continue
		}
		h.hub.Broadcast(roomID, gin.H{"type": "danmaku", "payload": row})
	}
}

func streamKey() string { return fmt.Sprintf("live-%d", time.Now().UnixNano()) }
