package danmaku

import (
	"net/http"
<<<<<<< HEAD

	danmakulogic "danmakustream/backend/internal/logic/danmaku"
	"danmakustream/backend/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ListHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req danmakulogic.DanmakuListReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := danmakulogic.NewListDanmakuLogic(r.Context(), svcCtx)
		resp, err := l.List(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}

func SendHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req danmakulogic.SendDanmakuReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := danmakulogic.NewSendDanmakuLogic(r.Context(), svcCtx)
		resp, err := l.Send(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		httpx.OkJsonCtx(r.Context(), w, resp)
	}
}
=======
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

func ListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		videoID, err := strconv.ParseUint(c.Param("videoId"), 10, 64)
		if err != nil || videoID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的视频 ID")
			return
		}

		var danmakus []model.Danmaku
		if err := svcCtx.DB.
			Where("video_id = ? AND blocked = ?", videoID, false).
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
>>>>>>> origin/dev
