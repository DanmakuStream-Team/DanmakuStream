package handler

import (
	"errors"
	"net/http"
	"strconv"

	"danmakustream/user-service/internal/handler/response"
	model "danmakustream/user-service/internal/model/mysql"
	"danmakustream/user-service/internal/svc"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Livez(c *gin.Context) { response.Ok(c, gin.H{"status": "up"}) }
func Version(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Ok(c, gin.H{"service": ctx.Config.Name, "version": ctx.Config.Version, "commit": ctx.Config.Commit, "buildTime": ctx.Config.BuildTime})
	}
}
func Health(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		if ctx == nil || ctx.DB == nil {
			response.Fail(c, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		sqlDB, err := ctx.DB.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			response.Fail(c, http.StatusServiceUnavailable, "database unavailable")
			return
		}
		response.Ok(c, gin.H{"status": "up", "database": "up", "dependencies": gin.H{}})
	}
}

func InternalUsers(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		ids := c.QueryArray("id")
		var users []model.User
		q := ctx.DB.Select("id", "username", "nickname", "avatar", "role")
		if len(ids) > 0 {
			q = q.Where("id IN ?", ids)
		}
		if err := q.Find(&users).Error; err != nil {
			response.Fail(c, 500, "query users failed")
			return
		}
		items := make([]model.UserInfo, 0, len(users))
		for _, u := range users {
			items = append(items, model.UserInfo{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Avatar: u.Avatar, Role: u.Role})
		}
		response.Ok(c, gin.H{"items": items})
	}
}
func InternalUserExists(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := positiveID(c.Param("id"))
		if !ok {
			response.Fail(c, http.StatusBadRequest, "invalid user id")
			return
		}
		var n int64
		if err := ctx.DB.Model(&model.User{}).Where("id = ?", id).Count(&n).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "query user failed")
			return
		}
		response.Ok(c, gin.H{"exists": n > 0})
	}
}
func InternalUser(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := positiveID(c.Param("id"))
		if !ok {
			response.Fail(c, http.StatusBadRequest, "invalid user id")
			return
		}
		var user model.User
		if err := ctx.DB.Select("id", "username", "nickname", "avatar", "role").First(&user, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Fail(c, http.StatusNotFound, "user not found")
			} else {
				response.Fail(c, http.StatusInternalServerError, "query user failed")
			}
			return
		}
		response.Ok(c, model.UserInfo{ID: user.ID, Username: user.Username, Nickname: user.Nickname, Avatar: user.Avatar, Role: user.Role})
	}
}
func InternalBlocked(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		firstID, secondID := c.Query("firstId"), c.Query("secondId")
		if firstID == "" {
			firstID = c.Query("blockerId")
		}
		if secondID == "" {
			secondID = c.Query("blockedId")
		}
		first, firstOK := positiveID(firstID)
		second, secondOK := positiveID(secondID)
		if !firstOK || !secondOK {
			response.Fail(c, http.StatusBadRequest, "invalid relationship ids")
			return
		}
		var n int64
		if err := ctx.DB.Model(&model.UserBlock{}).Where("(blocker_id=? AND blocked_id=?) OR (blocker_id=? AND blocked_id=?)", first, second, second, first).Count(&n).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "query blocked relationship failed")
			return
		}
		response.Ok(c, gin.H{"blocked": n > 0})
	}
}
func InternalMembership(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		subscriberID := c.Query("subscriberId")
		if subscriberID == "" {
			subscriberID = c.Query("userId")
		}
		subscriber, subscriberOK := positiveID(subscriberID)
		creator, creatorOK := positiveID(c.Query("creatorId"))
		if !subscriberOK || !creatorOK {
			response.Fail(c, http.StatusBadRequest, "invalid membership ids")
			return
		}
		var n int64
		if err := ctx.DB.Model(&model.CreatorSubscription{}).Where("subscriber_id=? AND creator_id=? AND status='active' AND expires_at > NOW()", subscriber, creator).Count(&n).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "query membership failed")
			return
		}
		response.Ok(c, gin.H{"active": n > 0})
	}
}
func InternalFollowing(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		follower, followerOK := positiveID(c.Query("followerId"))
		followee, followeeOK := positiveID(c.Query("followeeId"))
		if !followerOK || !followeeOK {
			response.Fail(c, http.StatusBadRequest, "invalid following ids")
			return
		}
		var n int64
		if err := ctx.DB.Model(&model.Follow{}).Where("follower_id=? AND followee_id=?", follower, followee).Count(&n).Error; err != nil {
			response.Fail(c, http.StatusInternalServerError, "query following relationship failed")
			return
		}
		response.Ok(c, gin.H{"following": n > 0})
	}
}

func positiveID(value string) (uint64, bool) {
	id, err := strconv.ParseUint(value, 10, 64)
	return id, err == nil && id > 0
}
