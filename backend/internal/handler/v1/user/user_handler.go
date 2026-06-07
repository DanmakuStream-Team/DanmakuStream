package user

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"danmakustream/backend/internal/handler/response"
	videologic "danmakustream/backend/internal/logic/video"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type profileInfo struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	Role        string `json:"role"`
	FollowCount int64  `json:"followCount"`
	FanCount    int64  `json:"fanCount"`
	VideoCount  int64  `json:"videoCount"`
	Followed    bool   `json:"followed"`
	CreatedAt   string `json:"createdAt"`
}

type publicUserVideoListReq struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

type userListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Keyword  string `form:"keyword"`
	Role     string `form:"role"`
}

type updateUserRoleReq struct {
	Role string `json:"role" binding:"required"`
}

type meVideoListReq struct {
	Page     int    `form:"page"`
	PageSize int    `form:"pageSize"`
	Status   string `form:"status"`
}

type updateMeReq struct {
	Nickname string `json:"nickname"`
	Bio      string `json:"bio"`
}

type userListItem struct {
	ID          uint   `json:"id"`
	Username    string `json:"username"`
	Nickname    string `json:"nickname"`
	Avatar      string `json:"avatar"`
	Bio         string `json:"bio"`
	Role        string `json:"role"`
	FollowCount int64  `json:"followCount"`
	FanCount    int64  `json:"fanCount"`
	VideoCount  int64  `json:"videoCount"`
	CreatedAt   string `json:"createdAt"`
}

func isValidVideoStatus(status string) bool {
	return status == "pending" || status == "approved" || status == "rejected"
}

func isValidUserRole(role string) bool {
	return role == "user" || role == "creator" || role == "admin"
}

func ProfileHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || userID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的用户 ID")
			return
		}

		var u model.User
		if err := svcCtx.DB.First(&u, userID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "用户不存在")
			return
		}

		var videoCount int64
		if err := svcCtx.DB.Model(&model.Video{}).
			Where("author_id = ? AND status = ?", userID, "approved").
			Count(&videoCount).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "用户投稿数量加载失败")
			return
		}

		currentUserID := getOptionalUserID(c, svcCtx)
		followed := false
		if currentUserID != 0 && currentUserID != uint(userID) {
			var count int64
			if err := svcCtx.DB.Model(&model.Follow{}).
				Where("follower_id = ? AND followee_id = ?", currentUserID, userID).
				Count(&count).Error; err != nil {
				response.Fail(c, http.StatusInternalServerError, "关注状态加载失败")
				return
			}
			followed = count > 0
		}

		response.Ok(c, profileInfo{
			ID:          u.ID,
			Username:    u.Username,
			Nickname:    u.Nickname,
			Avatar:      u.Avatar,
			Bio:         u.Bio,
			Role:        u.Role,
			FollowCount: u.FollowCount,
			FanCount:    u.FanCount,
			VideoCount:  videoCount,
			Followed:    followed,
			CreatedAt:   u.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
}

func getOptionalUserID(c *gin.Context, svcCtx *svc.ServiceContext) uint {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return 0
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(svcCtx.Config.Auth.AccessSecret), nil
	})
	if err != nil || !token.Valid {
		return 0
	}
	return claims.UserID
}

func FollowHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentUserID := c.GetUint(middleware.CtxKeyUserID)

		targetUserID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || targetUserID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的用户 ID")
			return
		}

		if currentUserID == uint(targetUserID) {
			response.Fail(c, http.StatusBadRequest, "不能关注自己")
			return
		}

		var target model.User
		if err := svcCtx.DB.Select("id").First(&target, targetUserID).Error; err != nil {
			response.Fail(c, http.StatusNotFound, "用户不存在")
			return
		}

		var followed bool

		err = svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var follow model.Follow
			err := tx.Where("follower_id = ? AND followee_id = ?", currentUserID, targetUserID).
				First(&follow).Error

			if err == nil {
				if err := tx.Unscoped().Delete(&follow).Error; err != nil {
					return err
				}
				followed = false

				if err := tx.Model(&model.User{}).
					Where("id = ? AND follow_count > 0", currentUserID).
					UpdateColumn("follow_count", gorm.Expr("follow_count - ?", 1)).Error; err != nil {
					return err
				}

				return tx.Model(&model.User{}).
					Where("id = ? AND fan_count > 0", targetUserID).
					UpdateColumn("fan_count", gorm.Expr("fan_count - ?", 1)).Error
			}

			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			if err := tx.Create(&model.Follow{
				FollowerID: currentUserID,
				FolloweeID: uint(targetUserID),
			}).Error; err != nil {
				return err
			}

			followed = true

			if err := tx.Model(&model.User{}).
				Where("id = ?", currentUserID).
				UpdateColumn("follow_count", gorm.Expr("follow_count + ?", 1)).Error; err != nil {
				return err
			}

			return tx.Model(&model.User{}).
				Where("id = ?", targetUserID).
				UpdateColumn("fan_count", gorm.Expr("fan_count + ?", 1)).Error
		})

		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "关注操作失败")
			return
		}

		response.Ok(c, gin.H{
			"followed": followed,
		})
	}
}

func VideosHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || userID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的用户 ID")
			return
		}

		var req publicUserVideoListReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 10
		}
		if req.PageSize > 100 {
			req.PageSize = 100
		}

		var userCount int64
		if err := svcCtx.DB.Model(&model.User{}).
			Where("id = ?", userID).
			Count(&userCount).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "用户加载失败")
			return
		}
		if userCount == 0 {
			response.Fail(c, http.StatusNotFound, "用户不存在")
			return
		}

		db := svcCtx.DB.Model(&model.Video{}).
			Preload("Author").
			Where("author_id = ? AND status = ?", userID, "approved")

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "视频列表加载失败")
			return
		}

		var videos []model.Video
		if err := db.
			Order("created_at DESC").
			Offset((req.Page - 1) * req.PageSize).
			Limit(req.PageSize).
			Find(&videos).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "视频列表加载失败")
			return
		}

		list := make([]videologic.VideoInfo, 0, len(videos))
		for _, video := range videos {
			list = append(list, videologic.VideoInfo{
				ID:           video.ID,
				Title:        video.Title,
				Description:  video.Description,
				CoverURL:     video.CoverURL,
				VideoURL:     video.VideoURL,
				Duration:     video.Duration,
				ViewCount:    video.ViewCount,
				LikeCount:    video.LikeCount,
				CollectCount: video.CollectCount,
				DanmakuCount: video.DanmakuCount,
				Status:       video.Status,
				Tags:         video.Tags,
				CreatedAt:    video.CreatedAt.Format("2006-01-02 15:04:05"),
				Author: &model.UserInfo{
					ID:       video.Author.ID,
					Username: video.Author.Username,
					Nickname: video.Author.Nickname,
					Avatar:   video.Author.Avatar,
					Role:     video.Author.Role,
				},
			})
		}

		response.Ok(c, videologic.PageResult[videologic.VideoInfo]{
			List:     list,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		})
	}
}

func UpdateMeHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		var req updateMeReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		updates := map[string]interface{}{}

		if req.Nickname != "" {
			updates["nickname"] = req.Nickname
		}
		if req.Bio != "" {
			updates["bio"] = req.Bio
		}

		if len(updates) == 0 {
			response.Fail(c, http.StatusBadRequest, "没有需要更新的内容")
			return
		}

		if err := svcCtx.DB.Model(&model.User{}).
			Where("id = ?", userID).
			Updates(updates).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "用户资料更新失败")
			return
		}

		response.Ok(c, nil)
	}
}

func FollowingListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		var followees []model.User
		err := svcCtx.DB.
			Table("users").
			Select("users.id, users.nickname, users.avatar, users.role").
			Joins("INNER JOIN follows ON follows.followee_id = users.id AND follows.deleted_at IS NULL").
			Where("follows.follower_id = ?", userID).
			Find(&followees).Error
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "关注列表加载失败")
			return
		}

		type followeeInfo struct {
			ID       uint   `json:"id"`
			Nickname string `json:"nickname"`
			Avatar   string `json:"avatar"`
			Role     string `json:"role"`
		}

		list := make([]followeeInfo, 0, len(followees))
		for _, u := range followees {
			list = append(list, followeeInfo{
				ID:       u.ID,
				Nickname: u.Nickname,
				Avatar:   u.Avatar,
				Role:     u.Role,
			})
		}

		response.Ok(c, gin.H{
			"list": list,
		})
	}
}

func PublicFollowingListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return publicFollowListHandler(svcCtx, "following")
}

func FollowersListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return publicFollowListHandler(svcCtx, "followers")
}

func publicFollowListHandler(svcCtx *svc.ServiceContext, listType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || userID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的用户 ID")
			return
		}

		var req publicUserVideoListReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		page, pageSize := normalizePage(req.Page, req.PageSize)

		var userCount int64
		if err := svcCtx.DB.Model(&model.User{}).Where("id = ?", userID).Count(&userCount).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "用户加载失败")
			return
		}
		if userCount == 0 {
			response.Fail(c, http.StatusNotFound, "用户不存在")
			return
		}

		var users []model.User
		var total int64
		db := svcCtx.DB.Table("users").
			Where("users.deleted_at IS NULL")
		if listType == "followers" {
			db = db.Joins("INNER JOIN follows ON follows.follower_id = users.id AND follows.deleted_at IS NULL").
				Where("follows.followee_id = ?", userID)
		} else {
			db = db.Joins("INNER JOIN follows ON follows.followee_id = users.id AND follows.deleted_at IS NULL").
				Where("follows.follower_id = ?", userID)
		}

		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "关注列表加载失败")
			return
		}

		if err := db.
			Select("users.*").
			Order("users.id DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&users).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "关注列表加载失败")
			return
		}

		response.Ok(c, gin.H{
			"list":     toUserListItems(svcCtx, users),
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

func AdminListHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req userListReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		page, pageSize := normalizePage(req.Page, req.PageSize)

		db := svcCtx.DB.Model(&model.User{})
		if req.Keyword != "" {
			keyword := "%" + strings.TrimSpace(req.Keyword) + "%"
			db = db.Where("username LIKE ? OR nickname LIKE ? OR bio LIKE ?", keyword, keyword, keyword)
		}
		if req.Role != "" {
			if !isValidUserRole(req.Role) {
				response.Fail(c, http.StatusBadRequest, "无效的用户角色")
				return
			}
			db = db.Where("role = ?", req.Role)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "用户列表加载失败")
			return
		}

		var users []model.User
		if err := db.
			Order("created_at DESC").
			Offset((page - 1) * pageSize).
			Limit(pageSize).
			Find(&users).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "用户列表加载失败")
			return
		}

		response.Ok(c, gin.H{
			"list":     toUserListItems(svcCtx, users),
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		})
	}
}

func AdminUpdateRoleHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || userID == 0 {
			response.Fail(c, http.StatusBadRequest, "无效的用户 ID")
			return
		}

		var req updateUserRoleReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		if !isValidUserRole(req.Role) {
			response.Fail(c, http.StatusBadRequest, "无效的用户角色")
			return
		}

		result := svcCtx.DB.Model(&model.User{}).
			Where("id = ?", userID).
			Update("role", req.Role)
		if result.Error != nil {
			response.Fail(c, http.StatusInternalServerError, "用户角色更新失败")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusNotFound, "用户不存在")
			return
		}

		response.Ok(c, gin.H{
			"id":   userID,
			"role": req.Role,
		})
	}
}

func normalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

func toUserListItems(svcCtx *svc.ServiceContext, users []model.User) []userListItem {
	userIDs := make([]uint, 0, len(users))
	for _, user := range users {
		userIDs = append(userIDs, user.ID)
	}

	videoCounts := make(map[uint]int64, len(users))
	if len(userIDs) > 0 {
		type videoCountRow struct {
			AuthorID uint
			Count    int64
		}
		var rows []videoCountRow
		_ = svcCtx.DB.Model(&model.Video{}).
			Select("author_id, COUNT(*) AS count").
			Where("author_id IN ? AND status = ?", userIDs, "approved").
			Group("author_id").
			Scan(&rows).Error
		for _, row := range rows {
			videoCounts[row.AuthorID] = row.Count
		}
	}

	list := make([]userListItem, 0, len(users))
	for _, user := range users {
		list = append(list, userListItem{
			ID:          user.ID,
			Username:    user.Username,
			Nickname:    user.Nickname,
			Avatar:      user.Avatar,
			Bio:         user.Bio,
			Role:        user.Role,
			FollowCount: user.FollowCount,
			FanCount:    user.FanCount,
			VideoCount:  videoCounts[user.ID],
			CreatedAt:   user.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return list
}

func UploadAvatarHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		avatarFile, err := c.FormFile("avatar")
		if err != nil {
			response.Fail(c, http.StatusBadRequest, "请上传头像文件")
			return
		}

		fileExt := filepath.Ext(avatarFile.Filename)
		if fileExt == "" {
			response.Fail(c, http.StatusBadRequest, "头像文件格式无效")
			return
		}

		avatarDir := filepath.Join(svcCtx.VideoDir, "avatars", fmt.Sprintf("%d", userID))
		if err := os.MkdirAll(avatarDir, 0755); err != nil {
			response.Fail(c, http.StatusInternalServerError, "创建头像目录失败")
			return
		}

		fileName := "avatar" + fileExt
		avatarPath := filepath.Join(avatarDir, fileName)

		src, err := avatarFile.Open()
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "读取头像文件失败")
			return
		}
		defer src.Close()

		dst, err := os.Create(avatarPath)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存头像文件失败")
			return
		}
		defer dst.Close()

		if _, err := io.Copy(dst, src); err != nil {
			response.Fail(c, http.StatusInternalServerError, "保存头像文件失败")
			return
		}

		avatarURL := fmt.Sprintf("/media/avatars/%d/%s", userID, fileName)

		if err := svcCtx.DB.Model(&model.User{}).
			Where("id = ?", userID).
			Update("avatar", avatarURL).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "头像更新失败")
			return
		}

		response.Ok(c, gin.H{
			"avatar": avatarURL,
		})
	}
}

func MeVideosHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)

		var req meVideoListReq
		if err := c.ShouldBindQuery(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}

		if req.Page <= 0 {
			req.Page = 1
		}
		if req.PageSize <= 0 {
			req.PageSize = 10
		}
		if req.PageSize > 100 {
			req.PageSize = 100
		}

		db := svcCtx.DB.Model(&model.Video{}).
			Preload("Author").
			Where("author_id = ?", userID)

		if req.Status != "" {
			if !isValidVideoStatus(req.Status) {
				response.Fail(c, http.StatusBadRequest, "无效的视频状态")
				return
			}
			db = db.Where("status = ?", req.Status)
		}

		var total int64
		if err := db.Count(&total).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "视频列表加载失败")
			return
		}

		var videos []model.Video
		if err := db.
			Order("created_at DESC").
			Offset((req.Page - 1) * req.PageSize).
			Limit(req.PageSize).
			Find(&videos).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "视频列表加载失败")
			return
		}

		list := make([]videologic.VideoInfo, 0, len(videos))
		for _, video := range videos {
			list = append(list, videologic.VideoInfo{
				ID:           video.ID,
				Title:        video.Title,
				Description:  video.Description,
				CoverURL:     video.CoverURL,
				VideoURL:     video.VideoURL,
				Duration:     video.Duration,
				ViewCount:    video.ViewCount,
				LikeCount:    video.LikeCount,
				CollectCount: video.CollectCount,
				DanmakuCount: video.DanmakuCount,
				Status:       video.Status,
				Tags:         video.Tags,
				CreatedAt:    video.CreatedAt.Format("2006-01-02 15:04:05"),
				Author: &model.UserInfo{
					ID:       video.Author.ID,
					Username: video.Author.Username,
					Nickname: video.Author.Nickname,
					Avatar:   video.Author.Avatar,
					Role:     video.Author.Role,
				},
			})
		}

		response.Ok(c, videologic.PageResult[videologic.VideoInfo]{
			List:     list,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		})
	}
}
