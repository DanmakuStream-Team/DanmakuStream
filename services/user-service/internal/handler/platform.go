package handler

import (
	"danmakustream/user-service/internal/handler/response"
	model "danmakustream/user-service/internal/model/mysql"
	"danmakustream/user-service/internal/svc"
	"github.com/gin-gonic/gin"
	"net/http"
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
		var n int64
		ctx.DB.Model(&model.User{}).Where("id = ?", c.Param("id")).Count(&n)
		response.Ok(c, gin.H{"exists": n > 0})
	}
}
func InternalBlocked(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var n int64
		ctx.DB.Model(&model.UserBlock{}).Where("(blocker_id=? AND blocked_id=?) OR (blocker_id=? AND blocked_id=?)", c.Query("firstId"), c.Query("secondId"), c.Query("secondId"), c.Query("firstId")).Count(&n)
		response.Ok(c, gin.H{"blocked": n > 0})
	}
}
func InternalMembership(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var n int64
		ctx.DB.Model(&model.CreatorSubscription{}).Where("subscriber_id=? AND creator_id=? AND status='active' AND expires_at > NOW()", c.Query("subscriberId"), c.Query("creatorId")).Count(&n)
		response.Ok(c, gin.H{"active": n > 0})
	}
}
