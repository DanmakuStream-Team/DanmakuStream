package admin

import (
	"danmakustream/user-service/internal/handler/response"
	model "danmakustream/user-service/internal/model/mysql"
	"danmakustream/user-service/internal/svc"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

func UserList(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []model.User
		var total int64
		ctx.DB.Model(&model.User{}).Count(&total)
		if err := ctx.DB.Order("id desc").Limit(100).Find(&users).Error; err != nil {
			response.Fail(c, 500, "query users failed")
			return
		}
		items := make([]model.UserInfo, 0, len(users))
		for _, u := range users {
			items = append(items, model.UserInfo{ID: u.ID, Username: u.Username, Nickname: u.Nickname, Avatar: u.Avatar, Role: u.Role})
		}
		response.Ok(c, gin.H{"items": items, "page": 1, "pageSize": 100, "total": total})
	}
}
func UpdateRole(ctx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			response.Fail(c, 400, "invalid user id")
			return
		}
		var req struct {
			Role string `json:"role"`
		}
		if c.ShouldBindJSON(&req) != nil || (req.Role != "user" && req.Role != "creator" && req.Role != "moderator" && req.Role != "admin") {
			response.Fail(c, 400, "invalid role")
			return
		}
		result := ctx.DB.Model(&model.User{}).Where("id=?", id).Update("role", req.Role)
		if result.Error != nil {
			response.Fail(c, 500, "update role failed")
			return
		}
		if result.RowsAffected == 0 {
			response.Fail(c, http.StatusNotFound, "user not found")
			return
		}
		response.Ok(c, gin.H{"id": id, "role": req.Role})
	}
}
