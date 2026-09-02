package svc

import (
	"log/slog"
	"net/http"

	"danmakustream/content-service/internal/handler"
	"danmakustream/content-service/internal/middleware"
	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
)

func Router(ctx *Context, logger *slog.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.MaxMultipartMemory = 16 << 20
	router.Use(middleware.RequestID(), middleware.AccessLog(logger))
	router.Use(gin.CustomRecovery(func(c *gin.Context, _ any) {
		response.Error(c, http.StatusInternalServerError, 50000, "internal server error")
		c.Abort()
	}))
	h := &handler.Handler{DB: ctx.DB, Logic: ctx.Logic, Config: ctx.Config}

	api := router.Group("/api/v1")
	api.GET("/livez", h.Livez)
	api.GET("/health", h.Health)
	api.GET("/version", h.Version)
	api.GET("/videos", h.ListVideos)
	api.GET("/videos/:id", h.VideoDetail)
	api.GET("/users/:id/videos", h.UserVideos)
	api.GET("/dynamics", h.ListDynamics)
	api.GET("/banners", h.PublicBanners)
	api.GET("/announcements", h.PublicAnnouncements)

	auth := api.Group("")
	auth.Use(middleware.Authenticate(ctx.Config.JWTSecret))
	auth.GET("/users/me/videos", h.MyVideos)
	auth.GET("/creator/analytics", h.CreatorAnalytics)
	auth.POST("/images/upload", h.UploadImage)
	auth.POST("/videos/upload", h.UploadVideo)
	auth.PUT("/videos/:id", h.UpdateVideo)
	auth.POST("/videos/:id/cover", h.UpdateCover)
	auth.GET("/videos/:id/download", h.DownloadVideo)
	auth.DELETE("/videos/:id", h.DeleteVideo)
	auth.POST("/videos/:id/collaborators", h.AddCollaborator)
	auth.DELETE("/videos/:id/collaborators/:userId", h.RemoveCollaborator)
	auth.POST("/dynamics", h.CreateDynamic)
	auth.DELETE("/dynamics/:id", h.DeleteDynamic)

	staff := auth.Group("/admin")
	staff.Use(middleware.RequireRoles("admin", "moderator"))
	staff.GET("/videos", h.AdminListVideos)
	staff.PUT("/videos/:id/status", h.ReviewVideo)

	internal := router.Group("/internal/v1")
	internal.Use(middleware.InternalAuth(ctx.Config.InternalAPIToken))
	internal.GET("/videos/batch", h.InternalVideos)
	internal.GET("/videos/:id", h.InternalVideo)

	admin := auth.Group("/admin")
	admin.Use(middleware.RequireRoles("admin"))
	admin.GET("/banners", h.AdminBanners)
	admin.POST("/banners", h.CreateBanner)
	admin.PUT("/banners/:id", h.UpdateBanner)
	admin.DELETE("/banners/:id", h.DeleteBanner)
	admin.GET("/announcements", h.AdminAnnouncements)
	admin.POST("/announcements", h.CreateAnnouncement)
	admin.PUT("/announcements/:id", h.UpdateAnnouncement)
	admin.DELETE("/announcements/:id", h.DeleteAnnouncement)

	router.StaticFS("/media", http.Dir(ctx.Config.StorageDir))
	router.NoRoute(func(c *gin.Context) { response.Error(c, http.StatusNotFound, 40400, "route not found") })
	return router
}
