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
	collectionhandler "danmakustream/backend/internal/handler/v1/collection"
	commenthandler "danmakustream/backend/internal/handler/v1/comment"
	creatorhandler "danmakustream/backend/internal/handler/v1/creator"
	danmakuhandler "danmakustream/backend/internal/handler/v1/danmaku"
	dynamichandler "danmakustream/backend/internal/handler/v1/dynamic"
	livehandler "danmakustream/backend/internal/handler/v1/live"
	mediahandler "danmakustream/backend/internal/handler/v1/media"
	membershiphandler "danmakustream/backend/internal/handler/v1/membership"
	messagehandler "danmakustream/backend/internal/handler/v1/message"
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
	if err := yaml.Unmarshal(data, c); err != nil {
		return err
	}
	// 容器和 Kubernetes 通过环境变量注入敏感配置，避免把密钥写入
	// ConfigMap 或镜像。未设置时仍兼容本地 YAML 配置。
	if value := os.Getenv("DMS_JWT_SECRET"); value != "" {
		c.Auth.AccessSecret = value
	}
	if value := os.Getenv("DMS_DATABASE_DSN"); value != "" {
		c.Database.DataSource = value
	}
	return nil
}

func main() {
	flag.Parse()

	var c config.Config
	if err := loadConfig(*configFile, &c); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	svcCtx := svc.NewServiceContext(c)
	livehandler.StartScheduleWorker(svcCtx)
	membershiphandler.StartExpirationWorker(svcCtx)

	r := gin.Default()
	r.MaxMultipartMemory = 8 << 20
	r.Use(middleware.VideoConnectionMiddleware())
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
	r.Static("/media/messages", svcCtx.VideoDir+"/messages")
	r.Static("/media/replays", svcCtx.VideoDir+"/live")

	authMW := middleware.AuthMiddleware(c.Auth.AccessSecret)

	v1 := r.Group("/api/v1")
	{
		// 仅验证进程存活，不依赖数据库，供 Kubernetes livenessProbe 使用。
		v1.GET("/livez", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"status": "ok"}})
		})
		// 无鉴权健康检查：供 Docker/K8s 探针与 CI 使用，顺带验证数据库连通性
		v1.GET("/health", func(ctx *gin.Context) {
			if sqlDB, err := svcCtx.DB.DB(); err == nil {
				if err := sqlDB.PingContext(ctx.Request.Context()); err == nil {
					ctx.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"status": "ok", "db": "up"}})
					return
				}
			}
			ctx.JSON(http.StatusServiceUnavailable, gin.H{"code": 503, "data": gin.H{"status": "degraded", "db": "down"}})
		})
		v1.POST("/auth/login", authhandler.LoginHandler(svcCtx))
		v1.POST("/auth/register", authhandler.RegisterHandler(svcCtx))
		// SRS publishes callbacks from the internal container network. The
		// unguessable stream key is validated against an active room.
		v1.POST("/live/hooks/srs", livehandler.SRSStreamHookHandler(svcCtx))

		// 视频列表，支持 sort=hot|date|like|collect 排序。
		v1.GET("/videos", videohandler.ListHandler(svcCtx))
		v1.GET("/videos/:id/collections", collectionhandler.VideoCollectionsHandler(svcCtx))
		v1.GET("/videos/:id", videohandler.DetailHandler(svcCtx))
		v1.GET("/collections/:id", collectionhandler.DetailHandler(svcCtx))
		v1.GET("/danmaku/:videoId", danmakuhandler.ListHandler(svcCtx))

		// 动态列表。
		v1.GET("/dynamics", dynamichandler.ListHandler(svcCtx))

		// 搜索用户，按用户名、昵称、简介匹配。
		v1.GET("/search/users", userhandler.SearchHandler(svcCtx))
		v1.GET("/users/:id", userhandler.ProfileHandler(svcCtx))
		v1.GET("/users/:id/videos", userhandler.VideosHandler(svcCtx))
		v1.GET("/creators/:id/membership-plan", membershiphandler.PlanHandler(svcCtx))

		// 评论列表，支持 sort=date|like 排序。
		v1.GET("/comments/:videoId", commenthandler.ListHandler(svcCtx))

		v1.GET("/live", livehandler.ListHandler(svcCtx))
		// 直播热度榜、礼物目录和观众助力榜。
		v1.GET("/live/rankings/heat", livehandler.HeatRankingHandler(svcCtx))
		v1.GET("/live-schedules", livehandler.ScheduleListHandler(svcCtx))
		// 已结束直播的回放列表、详情和历史弹幕。
		v1.GET("/live-replays", livehandler.ReplayListHandler(svcCtx))
		v1.GET("/live-replays/:id", livehandler.ReplayDetailHandler(svcCtx))
		v1.GET("/live-replays/:id/danmaku", livehandler.ReplayDanmakuHandler(svcCtx))
		v1.GET("/live/:id", livehandler.DetailHandler(svcCtx))
		v1.GET("/live/:id/interaction", livehandler.InteractionHandler(svcCtx))
		v1.GET("/live/:id/danmaku", livehandler.CurrentDanmakuHandler(svcCtx))
	}

	auth := v1.Group("")
	auth.Use(authMW)
	{
		auth.GET("/auth/me", authhandler.MeHandler(svcCtx))
		auth.PUT("/users/me", userhandler.UpdateMeHandler(svcCtx))
		auth.POST("/users/me/avatar", userhandler.UploadAvatarHandler(svcCtx))
		auth.GET("/users/me/videos", userhandler.MeVideosHandler(svcCtx))
		// 跨设备观看历史与稍后再看。
		auth.GET("/users/me/history", userhandler.HistoryListHandler(svcCtx))
		auth.GET("/users/me/history/:videoId", userhandler.HistoryDetailHandler(svcCtx))
		auth.PUT("/users/me/history/:videoId", userhandler.SaveHistoryHandler(svcCtx))
		auth.DELETE("/users/me/history/:videoId", userhandler.DeleteHistoryHandler(svcCtx))
		auth.DELETE("/users/me/history", userhandler.ClearHistoryHandler(svcCtx))
		auth.GET("/users/me/watch-later", userhandler.WatchLaterListHandler(svcCtx))
		auth.GET("/users/me/watch-later/:videoId/status", userhandler.WatchLaterStatusHandler(svcCtx))
		auth.POST("/users/me/watch-later/:videoId", userhandler.ToggleWatchLaterHandler(svcCtx))
		auth.DELETE("/users/me/watch-later/:videoId", userhandler.DeleteWatchLaterHandler(svcCtx))
		auth.DELETE("/users/me/watch-later", userhandler.ClearWatchLaterHandler(svcCtx))
		// 创作者后台按日数据。
		auth.GET("/creator/analytics", creatorhandler.AnalyticsHandler(svcCtx))
		auth.GET("/users/me/collections", collectionhandler.MineHandler(svcCtx))
		auth.GET("/users/following", userhandler.FollowingListHandler(svcCtx))
		auth.POST("/users/:id/follow", userhandler.FollowHandler(svcCtx))
		// 关注分组、特别关注与黑名单管理。
		auth.GET("/users/follow-groups", userhandler.FollowGroupListHandler(svcCtx))
		auth.POST("/users/follow-groups", userhandler.CreateFollowGroupHandler(svcCtx))
		auth.PUT("/users/follow-groups/:id", userhandler.UpdateFollowGroupHandler(svcCtx))
		auth.DELETE("/users/follow-groups/:id", userhandler.DeleteFollowGroupHandler(svcCtx))
		auth.PUT("/users/:id/follow-settings", userhandler.UpdateFollowSettingsHandler(svcCtx))
		auth.GET("/users/blocked", userhandler.BlockedListHandler(svcCtx))
		auth.POST("/users/:id/block", userhandler.BlockHandler(svcCtx))
		// 创作者付费特别关注套餐、订单和订阅状态。
		auth.GET("/creator/membership-plan", membershiphandler.MyPlanHandler(svcCtx))
		auth.PUT("/creator/membership-plan", membershiphandler.UpdateMyPlanHandler(svcCtx))
		auth.GET("/subscriptions", membershiphandler.MineHandler(svcCtx))
		auth.GET("/subscriptions/orders", membershiphandler.OrderListHandler(svcCtx))
		auth.GET("/subscriptions/creators/:id/status", membershiphandler.StatusHandler(svcCtx))
		auth.POST("/subscriptions/orders", membershiphandler.CreateOrderHandler(svcCtx))
		auth.POST("/subscriptions/orders/:orderNo/demo-pay", membershiphandler.DemoPayHandler(svcCtx))
		auth.PUT("/subscriptions/:creatorId/auto-renew", membershiphandler.AutoRenewHandler(svcCtx))
		auth.POST("/images/upload", mediahandler.UploadImageHandler(svcCtx))
		// 私信图片和短视频附件上传，不创建投稿记录。
		auth.POST("/messages/media", mediahandler.UploadMessageMediaHandler(svcCtx))

		auth.POST("/videos/upload", videohandler.UploadHandler(svcCtx))
		auth.PUT("/videos/:id", videohandler.UpdateHandler(svcCtx))
		auth.POST("/videos/:id/cover", videohandler.UpdateCoverHandler(svcCtx))
		auth.GET("/videos/:id/download", videohandler.DownloadHandler(svcCtx))
		auth.DELETE("/videos/:id", videohandler.DeleteHandler(svcCtx))
		auth.POST("/videos/:id/like", videohandler.LikeHandler(svcCtx))
		auth.POST("/videos/:id/collect", videohandler.CollectHandler(svcCtx))
		// 视频共创成员管理。
		auth.POST("/videos/:id/collaborators", collectionhandler.AddCollaboratorHandler(svcCtx))
		auth.DELETE("/videos/:id/collaborators/:userId", collectionhandler.RemoveCollaboratorHandler(svcCtx))

		// 视频合集创建和视频增删。
		auth.POST("/collections", collectionhandler.CreateHandler(svcCtx))
		auth.POST("/collections/:id/videos", collectionhandler.AddVideoHandler(svcCtx))
		auth.DELETE("/collections/:id/videos/:videoId", collectionhandler.RemoveVideoHandler(svcCtx))

		auth.POST("/danmaku", danmakuhandler.SendHandler(svcCtx))
		// 上传 .danmaku 文件，批量创建高级弹幕。
		auth.POST("/danmaku/advanced/upload", danmakuhandler.UploadAdvancedHandler(svcCtx))
		auth.POST("/comments", commenthandler.CreateHandler(svcCtx))
		auth.DELETE("/comments/:id", commenthandler.DeleteHandler(svcCtx))
		auth.POST("/comments/:id/like", commenthandler.LikeHandler(svcCtx))

		// 动态发布和删除。
		auth.POST("/dynamics", dynamichandler.CreateHandler(svcCtx))
		auth.DELETE("/dynamics/:id", dynamichandler.DeleteHandler(svcCtx))

		// 直播创建、预约和结束。
		auth.POST("/live", livehandler.CreateHandler(svcCtx))
		auth.GET("/live/:id/manage", livehandler.ManageDetailHandler(svcCtx))
		// 获取主播悬浮监看窗中的近期评论和 SC。
		auth.GET("/live/:id/monitor", livehandler.MonitorHandler(svcCtx))
		// 配置直播弹幕权限、慢速模式和置顶公告。
		auth.PUT("/live/:id/chat-settings", livehandler.UpdateChatSettingsHandler(svcCtx))
		auth.POST("/live-schedules", livehandler.CreateScheduleHandler(svcCtx))
		auth.DELETE("/live-schedules/:id", livehandler.CancelScheduleHandler(svcCtx))
		auth.POST("/live-schedules/:id/reserve", livehandler.ReserveScheduleHandler(svcCtx))
		auth.PUT("/live/:id/end", livehandler.EndHandler(svcCtx))
		auth.POST("/live/:id/like", livehandler.LikeHandler(svcCtx))
		auth.GET("/live/:id/like/status", livehandler.LikeStatusHandler(svcCtx))
		// 赠送普通礼物，附带留言时作为定时悬浮 SC。
		auth.POST("/live/:id/gifts", livehandler.GiftHandler(svcCtx))

		// 通知列表和已读状态。
		auth.GET("/notifications", notificationhandler.ListHandler(svcCtx))
		auth.PUT("/notifications", notificationhandler.ReadAllHandler(svcCtx))
		auth.PUT("/notifications/:id/read", notificationhandler.ReadHandler(svcCtx))

		// 私信会话、历史消息和已读状态。
		auth.GET("/messages/conversations", messagehandler.ConversationListHandler(svcCtx))
		auth.GET("/messages/unread", messagehandler.UnreadHandler(svcCtx))
		auth.GET("/messages/:userId", messagehandler.HistoryHandler(svcCtx))
		auth.POST("/messages", messagehandler.SendHandler(svcCtx))
		auth.PUT("/messages/:userId/read", messagehandler.ReadHandler(svcCtx))
	}

	staff := v1.Group("")
	staff.Use(authMW, middleware.StaffMiddleware)
	{
		staff.GET("/admin/videos", videohandler.AdminListHandler(svcCtx))
		staff.PUT("/admin/videos/:id/status", videohandler.AdminUpdateStatusHandler(svcCtx))
		staff.GET("/admin/danmaku", danmakuhandler.AdminListHandler(svcCtx))
		staff.PUT("/admin/danmaku/:id/block", danmakuhandler.BlockHandler(svcCtx))
	}

	admin := v1.Group("")
	admin.Use(authMW, middleware.AdminMiddleware)
	{
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
	r.GET("/ws/live-publish/:id", wshandler.LivePublishWebSocketHandler(svcCtx))
	r.GET("/ws/chat", wshandler.ChatWebSocketHandler(svcCtx))

	addr := fmt.Sprintf("%s:%d", c.Host, c.Port)
	fmt.Printf("DanmakuStream API server starting on %s\n", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func notImplemented(c *gin.Context) {
	response.Fail(c, http.StatusNotImplemented, "not implemented yet")
}
