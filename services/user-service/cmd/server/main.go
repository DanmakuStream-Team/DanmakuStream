package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"danmakustream/user-service/internal/config"
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

func main() {
	configFile := flag.String("f", "etc/config.example.yaml", "configuration file")
	flag.Parse()
	c, err := config.Load(*configFile)
	if err != nil {
		log.Fatal(err)
	}
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatal(err)
	}
	membershiphandler.StartExpirationWorker(ctx)
	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestContext(c.Name))
	r.MaxMultipartMemory = 8 << 20
	r.Static("/media/avatars", ctx.VideoDir+"/avatars")
	r.Static("/media/messages", ctx.VideoDir+"/messages")
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
	internal := r.Group("/internal/v1")
	internal.Use(middleware.InternalAuth(c.InternalToken))
	internal.GET("/users", platform.InternalUsers(ctx))
	internal.GET("/users/:id/exists", platform.InternalUserExists(ctx))
	internal.GET("/relationships/blocked", platform.InternalBlocked(ctx))
	internal.GET("/memberships/status", platform.InternalMembership(ctx))
	r.GET("/ws/chat", wshandler.Chat(ctx))
	server := &http.Server{Addr: fmt.Sprintf("%s:%d", c.Host, c.Port), Handler: r, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 15 * time.Second, WriteTimeout: 30 * time.Second, IdleTimeout: 60 * time.Second}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdown)
}
