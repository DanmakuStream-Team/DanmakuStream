package dynamic

import (
	"net/http"
	"strconv"
	"strings"

	"danmakustream/backend/internal/handler/response"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type createDynamicReq struct {
	Content string `json:"content" binding:"required,max=1000"`
	Images  string `json:"images"`
}

type dynamicInfo struct {
	ID        uint            `json:"id"`
	UserID    uint            `json:"userId"`
	Content   string          `json:"content"`
	Images    string          `json:"images"`
	Author    *model.UserInfo `json:"author"`
	CreatedAt string          `json:"createdAt"`
}

func ListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		page, pageSize := getPage(c)

		db := svcCtx.DB.Model(&model.DynamicPost{}).Preload("User")
		if userID := c.Query("userId"); userID != "" {
			parsed, err := strconv.ParseUint(userID, 10, 64)
			if err != nil || parsed == 0 {
				response.Fail(c, http.StatusBadRequest, "无效的用户 ID")
				return
			}
			db = db.Where("user_id = ?", parsed)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "动态列表加载失败")
			return
		}

		var posts []model.DynamicPost
		if err := db.Order("created_at DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&posts).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "动态列表加载失败")
			return
		}

		list := make([]dynamicInfo, 0, len(posts))
		for _, post := range posts {
			list = append(list, toDynamicInfo(post))
		}

		response.Ok(c, gin.H{
			"list":     list,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

func CreateHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		var req createDynamicReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		content := strings.TrimSpace(req.Content)
		if content == "" {
			response.Fail(c, http.StatusBadRequest, "动态内容不能为空")
			return
		}

		post := model.DynamicPost{
			UserID:  userID,
			Content: content,
			Images:  strings.TrimSpace(req.Images),
		}

		err := svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&post).Error; err != nil {
				return err
			}
			return notifyFollowers(tx, userID, "dynamic", "你关注的用户发布了新动态", content, "/dynamic")
		})
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "动态发布失败")
			return
		}

		if err := svcCtx.DB.Preload("User").First(&post, post.ID).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "动态加载失败")
			return
		}

		response.Ok(c, toDynamicInfo(post))
	}
}

func DeleteHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		role := c.GetString(middleware.CtxKeyRole)

		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的动态 ID")
			return
		}

		var post model.DynamicPost
		if err := svcCtx.DB.First(&post, id).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "动态不存在")
			return
		}

		if post.UserID != userID && role != "admin" {
			response.Fail(c, http.StatusForbidden, "无权删除该动态")
			return
		}

		if err := svcCtx.DB.Delete(&post).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "动态删除失败")
			return
		}

		response.Ok(c, gin.H{"id": post.ID})
	}
}

func notifyFollowers(tx *gorm.DB, actorID uint, noticeType, title, content, link string) error {
	var follows []model.Follow
	if err := tx.Where("followee_id = ?", actorID).Find(&follows).Error; err != nil {
		return err
	}

	if len(follows) == 0 {
		return nil
	}

	notifications := make([]model.Notification, 0, len(follows))
	for _, follow := range follows {
		if follow.FollowerID == actorID {
			continue
		}
		actor := actorID
		notifications = append(notifications, model.Notification{
			UserID:  follow.FollowerID,
			ActorID: &actor,
			Type:    noticeType,
			Title:   title,
			Content: content,
			Link:    link,
		})
	}

	if len(notifications) == 0 {
		return nil
	}
	return tx.Create(&notifications).Error
}

func toDynamicInfo(post model.DynamicPost) dynamicInfo {
	return dynamicInfo{
		ID:      post.ID,
		UserID:  post.UserID,
		Content: post.Content,
		Images:  post.Images,
		Author: &model.UserInfo{
			ID:       post.User.ID,
			Username: post.User.Username,
			Nickname: post.User.Nickname,
			Avatar:   post.User.Avatar,
			Role:     post.User.Role,
		},
		CreatedAt: post.CreatedAt.Format("2006-01-02 15:04:05"),
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
