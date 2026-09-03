package server

import (
	platform "danmakustream/user-service/internal/handler"
	adminhandler "danmakustream/user-service/internal/handler/v1/admin"
	authhandler "danmakustream/user-service/internal/handler/v1/auth"
	mediahandler "danmakustream/user-service/internal/handler/v1/media"
	membershiphandler "danmakustream/user-service/internal/handler/v1/membership"
	messagehandler "danmakustream/user-service/internal/handler/v1/message"
	notificationhandler "danmakustream/user-service/internal/handler/v1/notification"
	userhandler "danmakustream/user-service/internal/handler/v1/user"
	wshandler "danmakustream/user-service/internal/handler/ws"
	"danmakustream/user-service/internal/middleware"
	"danmakustream/user-service/internal/svc"

	"github.com/gin-gonic/gin"
)

// Router is the single route registry used by production and API regression tests.
func Router(ctx *svc.ServiceContext) *gin.Engine {
	c := ctx.Config
	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestContext(c.Name))
	r.MaxMultipartMemory = 8 << 20
	r.Static("/media/avatars", ctx.VideoDir+"/avatars")
	r.Static("/media/messages", ctx.VideoDir+"/messages")
	r.Static("/media/images", ctx.VideoDir+"/images")
	v1 := r.Group("/api/v1")
	v1.GET("/livez", platform.Livez)
	v1.GET("/health", platform.Health(ctx))
	v1.GET("/version", platform.Version(ctx))
	v1.POST("/auth/login", authhandler.LoginHandler(ctx))
	v1.POST("/auth/register", authhandler.RegisterHandler(ctx))
	v1.GET("/search/users", userhandler.SearchHandler(ctx))
	v1.GET("/users/:id", userhandler.ProfileHandler(ctx))
	v1.GET("/creators/:id/membership-plan", membershiphandler.PlanHandler(ctx))
	auth := v1.Group("")
	auth.Use(middleware.AuthMiddleware(c.Auth.AccessSecret))
	auth.GET("/auth/me", authhandler.MeHandler(ctx))
	auth.PUT("/users/me", userhandler.UpdateMeHandler(ctx))
	auth.POST("/users/me/avatar", userhandler.UploadAvatarHandler(ctx))
	auth.GET("/users/following", userhandler.FollowingListHandler(ctx))
	auth.POST("/users/:id/follow", userhandler.FollowHandler(ctx))
	auth.GET("/users/follow-groups", userhandler.FollowGroupListHandler(ctx))
	auth.POST("/users/follow-groups", userhandler.CreateFollowGroupHandler(ctx))
	auth.PUT("/users/follow-groups/:id", userhandler.UpdateFollowGroupHandler(ctx))
	auth.DELETE("/users/follow-groups/:id", userhandler.DeleteFollowGroupHandler(ctx))
	auth.PUT("/users/:id/follow-settings", userhandler.UpdateFollowSettingsHandler(ctx))
	auth.GET("/users/blocked", userhandler.BlockedListHandler(ctx))
	auth.POST("/users/:id/block", userhandler.BlockHandler(ctx))
	auth.GET("/creator/membership-plan", membershiphandler.MyPlanHandler(ctx))
	auth.PUT("/creator/membership-plan", membershiphandler.UpdateMyPlanHandler(ctx))
	auth.GET("/subscriptions", membershiphandler.MineHandler(ctx))
	auth.GET("/subscriptions/orders", membershiphandler.OrderListHandler(ctx))
	auth.GET("/subscriptions/creators/:id/status", membershiphandler.StatusHandler(ctx))
	auth.POST("/subscriptions/orders", membershiphandler.CreateOrderHandler(ctx))
	auth.POST("/subscriptions/orders/:orderNo/demo-pay", membershiphandler.DemoPayHandler(ctx))
	auth.PUT("/subscriptions/:creatorId/auto-renew", membershiphandler.AutoRenewHandler(ctx))
	auth.POST("/messages/media", mediahandler.UploadMessageMediaHandler(ctx))
	auth.POST("/images/upload", mediahandler.UploadImageHandler(ctx))
	auth.GET("/messages/conversations", messagehandler.ConversationListHandler(ctx))
	auth.GET("/messages/unread", messagehandler.UnreadHandler(ctx))
	auth.GET("/messages/:userId", messagehandler.HistoryHandler(ctx))
	auth.POST("/messages", messagehandler.SendHandler(ctx))
	auth.PUT("/messages/:userId/read", messagehandler.ReadHandler(ctx))
	auth.GET("/notifications", notificationhandler.ListHandler(ctx))
	auth.PUT("/notifications", notificationhandler.ReadAllHandler(ctx))
	auth.PUT("/notifications/:id/read", notificationhandler.ReadHandler(ctx))
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthMiddleware(c.Auth.AccessSecret), middleware.AdminMiddleware)
	admin.GET("/users", adminhandler.UserList(ctx))
	admin.PUT("/users/:id/role", adminhandler.UpdateRole(ctx))
	admin.GET("/infrastructure", adminhandler.Infrastructure(ctx))
	internal := r.Group("/internal/v1")
	internal.Use(middleware.InternalAuth(c.InternalToken))
	internal.GET("/users", platform.InternalUsers(ctx))
	internal.GET("/users/:id", platform.InternalUser(ctx))
	internal.GET("/users/:id/exists", platform.InternalUserExists(ctx))
	internal.GET("/relationships/blocked", platform.InternalBlocked(ctx))
	internal.GET("/relationships/following", platform.InternalFollowing(ctx))
	internal.GET("/memberships/status", platform.InternalMembership(ctx))
	internal.POST("/users/:id/followers/notifications", notificationhandler.FollowersHandler(ctx))
	r.GET("/ws/chat", wshandler.Chat(ctx))
	return r
}
