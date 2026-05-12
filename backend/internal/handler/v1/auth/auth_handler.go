package handler

import (
	"net/http"

	"danmakustream/backend/internal/handler/response"
	authlogic "danmakustream/backend/internal/logic/auth"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
)

func LoginHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req authlogic.LoginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		l := authlogic.NewLoginLogic(c.Request.Context(), svcCtx)
		resp, err := l.Login(&req)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Ok(c, resp)
	}
}

func RegisterHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req authlogic.RegisterReq
		if err := c.ShouldBindJSON(&req); err != nil {
			response.Fail(c, http.StatusBadRequest, "参数错误")
			return
		}
		l := authlogic.NewRegisterLogic(c.Request.Context(), svcCtx)
		resp, err := l.Register(&req)
		if err != nil {
			response.Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Ok(c, resp)
	}
}
