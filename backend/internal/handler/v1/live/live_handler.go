package live

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"danmakustream/backend/internal/handler/response"
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

type liveRoomInfo struct {
	ID          uint            `json:"id"`
	Title       string          `json:"title"`
	CoverURL    string          `json:"coverUrl"`
	StreamKey   string          `json:"streamKey,omitempty"`
	PublishURL  string          `json:"publishUrl,omitempty"`
	PlayURL     string          `json:"playUrl"`
	StreamURL   string          `json:"streamUrl"`
	Status      string          `json:"status"`
	ViewerCount int64           `json:"viewerCount"`
	OwnerID     uint            `json:"ownerId"`
	Owner       *model.UserInfo `json:"owner,omitempty"`
	StartedAt   string          `json:"startedAt,omitempty"`
	EndedAt     string          `json:"endedAt,omitempty"`
	CreatedAt   string          `json:"createdAt"`
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
		} else {
			updates := map[string]any{
				"title":        title,
				"cover_url":    strings.TrimSpace(req.CoverURL),
				"stream_key":   streamKey,
				"status":       "live",
				"viewer_count": 0,
				"started_at":   &now,
				"ended_at":     nil,
			}
			if err := svcCtx.DB.Model(&room).Updates(updates).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "直播间创建失败")
				return
			}
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
		ID:          room.ID,
		Title:       room.Title,
		CoverURL:    room.CoverURL,
		PlayURL:     playURL,
		StreamURL:   playURL,
		Status:      room.Status,
		ViewerCount: room.ViewerCount,
		OwnerID:     room.OwnerID,
		CreatedAt:   room.CreatedAt.Format("2006-01-02 15:04:05"),
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
