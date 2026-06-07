package notification

import (
	"net/http"
	"strconv"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
)

type notificationInfo struct {
	ID        uint            `json:"id"`
	Type      string          `json:"type"`
	Title     string          `json:"title"`
	Content   string          `json:"content"`
	Link      string          `json:"link"`
	Read      bool            `json:"read"`
	Actor     *model.UserInfo `json:"actor,omitempty"`
	CreatedAt string          `json:"createdAt"`
}

func ListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		page, pageSize := getPage(c)

		db := svcCtx.DB.Model(&model.Notification{}).
			Preload("Actor").
			Where("user_id = ?", userID)

		if c.Query("read") != "" {
			read, err := strconv.ParseBool(c.Query("read"))
			if err != nil {
				response.Fail(c, http.StatusBadRequest, "无效的已读状态")
				return
			}
			db = db.Where("`read` = ?", read)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "通知列表加载失败")
			return
		}

		var notifications []model.Notification
		if err := db.Order("created_at DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&notifications).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "通知列表加载失败")
			return
		}

		list := make([]notificationInfo, 0, len(notifications))
		for _, item := range notifications {
			list = append(list, toNotificationInfo(item))
		}

		var unreadCount int64
		if err := svcCtx.DB.Model(&model.Notification{}).
			Where("user_id = ? AND `read` = ?", userID, false).
			Count(&unreadCount).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "未读通知数量加载失败")
			return
		}

		response.Ok(c, gin.H{
			"list":        list,
			"total":       total,
			"unreadCount": unreadCount,
			"page":        page,
			"pageSize":    pageSize,
		})
	}
}

func ReadHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的通知 ID")
			return
		}

		result := svcCtx.DB.Model(&model.Notification{}).
			Where("id = ? AND user_id = ?", id, userID).
			Update("read", true)
		if result.Error != nil {
			response.Fail(c, http.StatusInternalServerError, "通知更新失败")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusNotFound, "通知不存在")
			return
		}

		response.Ok(c, gin.H{"id": id, "read": true})
	}
}

func ReadAllHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		if err := svcCtx.DB.Model(&model.Notification{}).
			Where("user_id = ? AND `read` = ?", userID, false).
			Update("read", true).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "通知更新失败")
			return
		}

		response.Ok(c, gin.H{"read": true})
	}
}

func toNotificationInfo(item model.Notification) notificationInfo {
	info := notificationInfo{
		ID:        item.ID,
		Type:      item.Type,
		Title:     item.Title,
		Content:   item.Content,
		Link:      item.Link,
		Read:      item.Read,
		CreatedAt: item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if item.ActorID != nil {
		info.Actor = &model.UserInfo{
			ID:       item.Actor.ID,
			Username: item.Actor.Username,
			Nickname: item.Actor.Nickname,
			Avatar:   item.Actor.Avatar,
			Role:     item.Actor.Role,
		}
	}
	return info
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
