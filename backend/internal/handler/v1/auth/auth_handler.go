package handler

import (
	"net/http"

	"danmakustream/backend/internal/handler/response"
	authlogic "danmakustream/backend/internal/logic/auth"
	"danmakustream/backend/internal/middleware"
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

<<<<<<< HEAD

func MeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        l := authlogic.NewMeLogic(r.Context(), svcCtx)
        resp, err := l.Me()
        if err != nil {
            httpx.ErrorCtx(r.Context(), w, err)
            return
        }
        httpx.OkJsonCtx(r.Context(), w, resp)
    }
}



func RegisterHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
=======
func RegisterHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
>>>>>>> origin/dev
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

func MeHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint(middleware.CtxKeyUserID)
		if userID == 0 {
			response.Fail(c, http.StatusUnauthorized, "未授权")
			return
		}
		l := authlogic.NewMeLogic(c.Request.Context(), svcCtx)
		resp, err := l.Me(userID)
		if err != nil {
			response.Fail(c, http.StatusNotFound, err.Error())
			return
		}
		response.Ok(c, resp)
	}
}
