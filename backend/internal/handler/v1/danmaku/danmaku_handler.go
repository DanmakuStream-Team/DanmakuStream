package danmaku

import (
	"net/http"
	"strconv"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type sendDanmakuReq struct {
	VideoID  uint   `json:"videoId" binding:"required"`
	Content  string `json:"content" binding:"required,max=200"`
	Time     int    `json:"time"`
	Color    string `json:"color"`
	FontSize string `json:"fontSize"`
	Type     string `json:"type"`
}

type danmakuItem struct {
	ID       uint   `json:"id"`
	VideoID  uint   `json:"videoId"`
	UserID   uint   `json:"userId"`
	Content  string `json:"content"`
	Time     int    `json:"time"`
	Color    string `json:"color"`
	FontSize string `json:"fontSize"`
	Type     string `json:"type"`
}

type adminDanmakuListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	VideoID  uint   `form:"videoId"`
	Scene    string `form:"scene"`
	Keyword  string `form:"keyword"`
	Blocked  *bool  `form:"blocked"`
}

type adminDanmakuItem struct {
	ID        uint   `json:"id"`
	VideoID   uint   `json:"videoId"`
	Scene     string `json:"scene"`
	UserID    uint   `json:"userId"`
	Content   string `json:"content"`
	Time      int    `json:"time"`
	Color     string `json:"color"`
	FontSize  string `json:"fontSize"`
	Type      string `json:"type"`
	Blocked   bool   `json:"blocked"`
	CreatedAt string `json:"createdAt"`
}

type pageResult[T any] struct {
	List     []T   `json:"list"`
	Total    int64 `json:"total"`
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
}

func ListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.ParseUint(c.Param("videoId"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var danmakus []model.Danmaku
		if err := svcCtx.DB.
			Where("video_id = ? AND scene = ? AND blocked = ?", videoID, "video", false).
			Order("time ASC").
			Find(&danmakus).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "弹幕加载失败")
			return
		}

		list := make([]danmakuItem, 0, len(danmakus))
		for _, d := range danmakus {
			list = append(list, danmakuItem{
				ID:       d.ID,
				VideoID:  d.VideoID,
				UserID:   d.UserID,
				Content:  d.Content,
				Time:     d.Time,
				Color:    d.Color,
				FontSize: d.FontSize,
				Type:     d.Type,
			})
		}
		response.Ok(c, list)
	}
}

func SendHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req sendDanmakuReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		userID := c.GetUint(middleware.CtxKeyUserID)
		if userID == 0 {
			response.Fail(c, http.StatusUnauthorized, "未授权")
			return
		}

		var video model.Video
		if err := svcCtx.DB.Select("id", "status").
			Where("id = ?", req.VideoID).
			First(&video).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "视频不存在")
			return
		}
		if video.Status != "approved" {
			response.Fail(c, http.StatusForbidden, "视频未通过审核")
			return
		}

		danmaku := model.Danmaku{
			VideoID:  req.VideoID,
			Scene:    "video",
			UserID:   userID,
			Content:  req.Content,
			Time:     req.Time,
			Color:    defaultStr(req.Color, "#FFFFFF"),
			FontSize: defaultStr(req.FontSize, "medium"),
			Type:     defaultStr(req.Type, "scroll"),
		}

		if err := svcCtx.DB.Create(&danmaku).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "弹幕发送失败")
			return
		}

		svcCtx.DB.Model(&model.Video{}).
			Where("id = ?", req.VideoID).
			UpdateColumn("danmaku_count", gorm.Expr("danmaku_count + ?", 1))

		response.Ok(c, danmakuItem{
			ID:       danmaku.ID,
			VideoID:  danmaku.VideoID,
			UserID:   danmaku.UserID,
			Content:  danmaku.Content,
			Time:     danmaku.Time,
			Color:    danmaku.Color,
			FontSize: danmaku.FontSize,
			Type:     danmaku.Type,
		})
	}
}

func AdminListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req adminDanmakuListReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 20
		}
		if req.PageSize > 100 {
			req.PageSize = 100
		}

		db := svcCtx.DB.Model(&model.Danmaku{})

		if req.VideoID > 0 {
			db = db.Where("video_id = ?", req.VideoID)
		}

		if req.Scene != "" {
			if req.Scene != "video" && req.Scene != "live" {
				response.Fail(c, http.StatusBadRequest, "无效的弹幕场景")
				return
			}
			db = db.Where("scene = ?", req.Scene)
		}

		if req.Keyword != "" {
			db = db.Where("content LIKE ?", "%"+req.Keyword+"%")
		}

		if req.Blocked != nil {
			db = db.Where("blocked = ?", *req.Blocked)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "弹幕列表加载失败")
			return
		}

		var danmakus []model.Danmaku
		if err := db.
			Order("created_at DESC").
			Offset((req.Page - 1) * req.PageSize).
			Limit(req.PageSize).
			Find(&danmakus).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "弹幕列表加载失败")
			return
		}

		list := make([]adminDanmakuItem, 0, len(danmakus))
		for _, d := range danmakus {
			list = append(list, adminDanmakuItem{
				ID:        d.ID,
				VideoID:   d.VideoID,
				Scene:     d.Scene,
				UserID:    d.UserID,
				Content:   d.Content,
				Time:      d.Time,
				Color:     d.Color,
				FontSize:  d.FontSize,
				Type:      d.Type,
				Blocked:   d.Blocked,
				CreatedAt: d.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}

		response.Ok(c, pageResult[adminDanmakuItem]{
			List:     list,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		})
	}
}

func BlockHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的弹幕 ID")
			return
		}

		result := svcCtx.DB.Model(&model.Danmaku{}).
			Where("id = ?", id).
			Update("blocked", true)
		if result.Error != nil {
			response.Fail(c, http.StatusInternalServerError, "屏蔽失败")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusNotFound, "弹幕不存在")
			return
		}
		response.Ok(c, nil)
	}
}

func defaultStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
