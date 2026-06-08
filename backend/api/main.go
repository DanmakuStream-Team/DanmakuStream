package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"danmakustream/backend/internal/config"
	"danmakustream/backend/internal/handler/response"
	adminhandler "danmakustream/backend/internal/handler/v1/admin"
	authhandler "danmakustream/backend/internal/handler/v1/auth"
	commenthandler "danmakustream/backend/internal/handler/v1/comment"
	danmakuhandler "danmakustream/backend/internal/handler/v1/danmaku"
	dynamichandler "danmakustream/backend/internal/handler/v1/dynamic"
	livehandler "danmakustream/backend/internal/handler/v1/live"
	mediahandler "danmakustream/backend/internal/handler/v1/media"
	notificationhandler "danmakustream/backend/internal/handler/v1/notification"
	userhandler "danmakustream/backend/internal/handler/v1/user"
	videohandler "danmakustream/backend/internal/handler/v1/video"
	wshandler "danmakustream/backend/internal/handler/ws"
	"danmakustream/backend/internal/middleware"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func loadConfig(path string, c *config.Config) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, c)
}

func main() {
	flag.Parse()

	var c config.Config
	if err := loadConfig(*configFile, &c); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	svcCtx := svc.NewServiceContext(c)

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20
	r.Use(middleware.TrafficMiddleware(svcCtx))

	r.Use(func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		ctx.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(http.StatusNoContent)
			return
		}
		ctx.Next()
	})

	r.Static("/media/videos", svcCtx.VideoDir+"/videos")
	r.Static("/media/covers", svcCtx.VideoDir+"/covers")
	r.Static("/media/avatars", svcCtx.VideoDir+"/avatars")
	r.Static("/media/images", svcCtx.VideoDir+"/images")

	authMW := middleware.AuthMiddleware(c.Auth.AccessSecret)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/login", authhandler.LoginHandler(svcCtx))
		v1.POST("/auth/register", authhandler.RegisterHandler(svcCtx))

		// 视频列表，支持 sort=hot|date|like|collect 排序。
		v1.GET("/videos", videohandler.ListHandler(svcCtx))
		v1.GET("/videos/:id", videohandler.DetailHandler(svcCtx))
		v1.GET("/danmaku/:videoId", danmakuhandler.ListHandler(svcCtx))

		// 动态列表。
		v1.GET("/dynamics", dynamichandler.ListHandler(svcCtx))

		// 搜索用户，按用户名、昵称、简介匹配。
		v1.GET("/search/users", userhandler.SearchHandler(svcCtx))
		v1.GET("/users/:id", userhandler.ProfileHandler(svcCtx))
		v1.GET("/users/:id/videos", userhandler.VideosHandler(svcCtx))

		// 评论列表，支持 sort=date|like 排序。
		v1.GET("/comments/:videoId", commenthandler.ListHandler(svcCtx))

		v1.GET("/live", livehandler.ListHandler(svcCtx))
		v1.GET("/live-schedules", livehandler.ScheduleListHandler(svcCtx))
		v1.GET("/live/:id", livehandler.DetailHandler(svcCtx))
	}

	auth := v1.Group("")
	auth.Use(authMW)
	{
		auth.GET("/auth/me", authhandler.MeHandler(svcCtx))
		auth.PUT("/users/me", userhandler.UpdateMeHandler(svcCtx))
		auth.POST("/users/me/avatar", userhandler.UploadAvatarHandler(svcCtx))
		auth.GET("/users/me/videos", userhandler.MeVideosHandler(svcCtx))
		auth.GET("/users/following", userhandler.FollowingListHandler(svcCtx))
		auth.POST("/users/:id/follow", userhandler.FollowHandler(svcCtx))
		auth.POST("/images/upload", mediahandler.UploadImageHandler(svcCtx))

		auth.POST("/videos/upload", videohandler.UploadHandler(svcCtx))
		auth.PUT("/videos/:id", videohandler.UpdateHandler(svcCtx))
		auth.POST("/videos/:id/cover", videohandler.UpdateCoverHandler(svcCtx))
		auth.GET("/videos/:id/download", videohandler.DownloadHandler(svcCtx))
		auth.DELETE("/videos/:id", videohandler.DeleteHandler(svcCtx))
		auth.POST("/videos/:id/like", videohandler.LikeHandler(svcCtx))
		auth.POST("/videos/:id/collect", videohandler.CollectHandler(svcCtx))

		auth.POST("/danmaku", danmakuhandler.SendHandler(svcCtx))
		auth.POST("/comments", commenthandler.CreateHandler(svcCtx))
		auth.DELETE("/comments/:id", commenthandler.DeleteHandler(svcCtx))
		auth.POST("/comments/:id/like", commenthandler.LikeHandler(svcCtx))

		// 动态发布和删除。
		auth.POST("/dynamics", dynamichandler.CreateHandler(svcCtx))
		auth.DELETE("/dynamics/:id", dynamichandler.DeleteHandler(svcCtx))

		// 直播创建、预约和结束。
		auth.POST("/live", livehandler.CreateHandler(svcCtx))
		auth.POST("/live-schedules", livehandler.CreateScheduleHandler(svcCtx))
		auth.DELETE("/live-schedules/:id", livehandler.CancelScheduleHandler(svcCtx))
		auth.POST("/live-schedules/:id/reserve", livehandler.ReserveScheduleHandler(svcCtx))
		auth.PUT("/live/:id/end", livehandler.EndHandler(svcCtx))

		// 通知列表和已读状态。
		auth.GET("/notifications", notificationhandler.ListHandler(svcCtx))
		auth.PUT("/notifications", notificationhandler.ReadAllHandler(svcCtx))
		auth.PUT("/notifications/:id/read", notificationhandler.ReadHandler(svcCtx))
	}

	admin := v1.Group("")
	admin.Use(authMW, middleware.AdminMiddleware)
	{
		admin.GET("/admin/videos", videohandler.AdminListHandler(svcCtx))
		admin.PUT("/admin/videos/:id/status", videohandler.AdminUpdateStatusHandler(svcCtx))
		admin.GET("/admin/danmaku", danmakuhandler.AdminListHandler(svcCtx))
		admin.PUT("/admin/danmaku/:id/block", danmakuhandler.BlockHandler(svcCtx))
		admin.GET("/admin/infrastructure", adminhandler.InfrastructureHandler(svcCtx))
		admin.GET("/admin/users", adminhandler.UserListHandler(svcCtx))
		admin.PUT("/admin/users/:id/role", adminhandler.UpdateUserRoleHandler(svcCtx))
		admin.GET("/admin/banners", adminhandler.BannerListHandler(svcCtx))
		admin.POST("/admin/banners", adminhandler.CreateBannerHandler(svcCtx))
		admin.PUT("/admin/banners/:id", adminhandler.UpdateBannerHandler(svcCtx))
		admin.DELETE("/admin/banners/:id", adminhandler.DeleteBannerHandler(svcCtx))
		admin.GET("/admin/announcements", adminhandler.AnnouncementListHandler(svcCtx))
		admin.POST("/admin/announcements", adminhandler.CreateAnnouncementHandler(svcCtx))
		admin.PUT("/admin/announcements/:id", adminhandler.UpdateAnnouncementHandler(svcCtx))
		admin.DELETE("/admin/announcements/:id", adminhandler.DeleteAnnouncementHandler(svcCtx))
	}

	r.GET("/ws/live/:id", wshandler.LiveWebSocketHandler(svcCtx))

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	fmt.Printf("DanmakuStream API server starting on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func notImplemented(c *gin.Context) {
	response.Fail(c, http.StatusNotImplemented, "not implemented yet")
}
