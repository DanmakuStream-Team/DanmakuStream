package user

import (
	"errors"
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

type followGroupRequest struct {
	Name string `json:"name"`
}

type followSettingsRequest struct {
	Special *bool `json:"special"`
	GroupID *uint `json:"groupId"`
}

type followGroupItem struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

func FollowGroupListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		var groups []model.FollowGroup
		if err := svcCtx.DB.Where("owner_id = ?", userID).Order("created_at ASC").Find(&groups).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "关注分组加载失败")
			return
		}

		list := make([]followGroupItem, 0, len(groups))
		for _, group := range groups {
			var count int64
			if err := svcCtx.DB.Model(&model.Follow{}).
				Where("follower_id = ? AND group_id = ?", userID, group.ID).
				Count(&count).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "关注分组统计失败")
				return
			}
			list = append(list, followGroupItem{ID: group.ID, Name: group.Name, Count: count})
		}
		response.Ok(c, gin.H{"list": list})
	}
}

func CreateFollowGroupHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		var req followGroupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" || len([]rune(name)) > 20 {
			response.Fail(c, http.StatusBadRequest, "分组名称应为 1 到 20 个字符")
			return
		}
		var count int64
		if err := svcCtx.DB.Model(&model.FollowGroup{}).Where("owner_id = ? AND name = ?", userID, name).Count(&count).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "分组状态检查失败")
			return
		}
		if count > 0 {
			response.Fail(c, http.StatusConflict, "同名分组已存在")
			return
		}
		group := model.FollowGroup{OwnerID: userID, Name: name}
		if err := svcCtx.DB.Create(&group).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "创建分组失败")
			return
		}
		response.Ok(c, followGroupItem{ID: group.ID, Name: group.Name, Count: 0})
	}
}

func UpdateFollowGroupHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		groupID, ok := parseRelationID(c, "id")
		if !ok {
			return
		}
		var req followGroupRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		name := strings.TrimSpace(req.Name)
		if name == "" || len([]rune(name)) > 20 {
			response.Fail(c, http.StatusBadRequest, "分组名称应为 1 到 20 个字符")
			return
		}
		var duplicateCount int64
		if err := svcCtx.DB.Model(&model.FollowGroup{}).
			Where("owner_id = ? AND name = ? AND id <> ?", userID, name, groupID).
			Count(&duplicateCount).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "分组状态检查失败")
			return
		}
		if duplicateCount > 0 {
			response.Fail(c, http.StatusConflict, "同名分组已存在")
			return
		}
		result := svcCtx.DB.Model(&model.FollowGroup{}).
			Where("id = ? AND owner_id = ?", groupID, userID).
			Update("name", name)
		if result.Error != nil {
			response.Fail(c, http.StatusInternalServerError, "修改分组失败")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusNotFound, "分组不存在")
			return
		}
		response.Ok(c, gin.H{"id": groupID, "name": name})
	}
}

func DeleteFollowGroupHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		groupID, ok := parseRelationID(c, "id")
		if !ok {
			return
		}
		err := svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var group model.FollowGroup
			if err := tx.Where("id = ? AND owner_id = ?", groupID, userID).First(&group).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Follow{}).
				Where("follower_id = ? AND group_id = ?", userID, groupID).
				Update("group_id", nil).Error; err != nil {
				return err
			}
			return tx.Unscoped().Delete(&group).Error
		})
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, http.StatusNotFound, "分组不存在")
			return
		}
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "删除分组失败")
			return
		}
		response.Ok(c, nil)
	}
}

func UpdateFollowSettingsHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		targetID, ok := parseRelationID(c, "id")
		if !ok {
			return
		}
		var req followSettingsRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		updates := map[string]interface{}{}
		if req.Special != nil {
			response.Fail(c, http.StatusForbidden, "特别关注由付费订阅状态管理")
			return
		}
		if req.GroupID != nil {
			if *req.GroupID == 0 {
				updates["group_id"] = nil
			} else {
				var count int64
				if err := svcCtx.DB.Model(&model.FollowGroup{}).
					Where("id = ? AND owner_id = ?", *req.GroupID, userID).
					Count(&count).Error; err != nil || count == 0 {
					response.Fail(c, http.StatusBadRequest, "关注分组不存在")
					return
				}
				updates["group_id"] = *req.GroupID
			}
		}
		if len(updates) == 0 {
			response.Fail(c, http.StatusBadRequest, "没有需要更新的设置")
			return
		}
		result := svcCtx.DB.Model(&model.Follow{}).
			Where("follower_id = ? AND followee_id = ?", userID, targetID).
			Updates(updates)
		if result.Error != nil {
			response.Fail(c, http.StatusInternalServerError, "关注设置保存失败")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusBadRequest, "请先关注该用户")
			return
		}
		response.Ok(c, nil)
	}
}

func BlockedListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		type blockedUser struct {
			ID        uint   `json:"id"`
			Nickname  string `json:"nickname"`
			Avatar    string `json:"avatar"`
			Role      string `json:"role"`
			BlockedAt string `json:"blockedAt"`
		}
		var list []blockedUser
		if err := svcCtx.DB.Table("users").
			Select("users.id, users.nickname, users.avatar, users.role, DATE_FORMAT(user_blocks.created_at, '%Y-%m-%d %H:%i:%s') AS blocked_at").
			Joins("INNER JOIN user_blocks ON user_blocks.blocked_id = users.id AND user_blocks.deleted_at IS NULL").
			Where("user_blocks.blocker_id = ?", userID).
			Order("user_blocks.created_at DESC").
			Scan(&list).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "黑名单加载失败")
			return
		}
		if list == nil {
			list = []blockedUser{}
		}
		response.Ok(c, gin.H{"list": list})
	}
}

func BlockHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		targetID, ok := parseRelationID(c, "id")
		if !ok {
			return
		}
		if userID == targetID {
			response.Fail(c, http.StatusBadRequest, "不能拉黑自己")
			return
		}
		var target model.User
		if err := svcCtx.DB.Select("id").First(&target, targetID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "用户不存在")
			return
		}

		blocked := false
		err := svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var relation model.UserBlock
			err := tx.Where("blocker_id = ? AND blocked_id = ?", userID, targetID).First(&relation).Error
			if err == nil {
				blocked = false
				return tx.Unscoped().Delete(&relation).Error
			}
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if err := tx.Create(&model.UserBlock{BlockerID: userID, BlockedID: targetID}).Error; err != nil {
				return err
			}
			blocked = true
			if err := removeFollow(tx, userID, targetID); err != nil {
				return err
			}
			return removeFollow(tx, targetID, userID)
		})
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "拉黑操作失败")
			return
		}
		response.Ok(c, gin.H{"blocked": blocked})
	}
}

func removeFollow(tx *gorm.DB, followerID, followeeID uint) error {
	var follow model.Follow
	err := tx.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&follow).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := tx.Unscoped().Delete(&follow).Error; err != nil {
		return err
	}
	if err := tx.Model(&model.User{}).Where("id = ? AND follow_count > 0", followerID).
		UpdateColumn("follow_count", gorm.Expr("follow_count - 1")).Error; err != nil {
		return err
	}
	return tx.Model(&model.User{}).Where("id = ? AND fan_count > 0", followeeID).
		UpdateColumn("fan_count", gorm.Expr("fan_count - 1")).Error
}

func parseRelationID(c *gin.Context, key string) (uint, bool) {
	value, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil || value == 0 {
		response.Fail(c, http.StatusBadRequest, "无效的 ID")
		return 0, false
	}
	return uint(value), true
}
