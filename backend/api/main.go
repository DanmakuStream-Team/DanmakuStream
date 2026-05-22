package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"danmakustream/backend/internal/config"
	"danmakustream/backend/internal/handler/response"
	authhandler "danmakustream/backend/internal/handler/v1/auth"
	commenthandler "danmakustream/backend/internal/handler/v1/comment"
	danmakuhandler "danmakustream/backend/internal/handler/v1/danmaku"
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

	// CORS
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

	// Serve uploaded media files
	r.Static("/media/videos", svcCtx.VideoDir+"/videos")
	r.Static("/media/covers", svcCtx.VideoDir+"/covers")

	authMW := middleware.AuthMiddleware(c.Auth.AccessSecret)

	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/auth/login", authhandler.LoginHandler(svcCtx))
		v1.POST("/auth/register", authhandler.RegisterHandler(svcCtx))
		v1.GET("/videos", videohandler.ListHandler(svcCtx))
		v1.GET("/videos/:id", videohandler.DetailHandler(svcCtx))
		v1.GET("/danmaku/:videoId", danmakuhandler.ListHandler(svcCtx))
		v1.GET("/users/:id", notImplemented)
		v1.GET("/comments/:videoId", commenthandler.ListHandler(svcCtx))
	}

	// Auth-required routes
	auth := v1.Group("")
	auth.Use(authMW)
	{
		auth.GET("/auth/me", authhandler.MeHandler(svcCtx))
		auth.POST("/videos/upload", videohandler.UploadHandler(svcCtx))
		auth.PUT("/videos/:id", notImplemented)
		auth.POST("/videos/:id/like", videohandler.LikeHandler(svcCtx))
		auth.POST("/videos/:id/collect", videohandler.CollectHandler(svcCtx))
		auth.POST("/danmaku", danmakuhandler.SendHandler(svcCtx))
		auth.POST("/comments", commenthandler.CreateHandler(svcCtx))
		auth.POST("/users/:id/follow", notImplemented)
		auth.GET("/live", notImplemented)
		auth.POST("/live", notImplemented)
		auth.PUT("/live/:id/end", notImplemented)
	}

	// Admin routes
	admin := v1.Group("")
	admin.Use(authMW, middleware.AdminMiddleware)
	{
		admin.GET("/admin/videos", notImplemented)
		admin.PUT("/admin/videos/:id/status", notImplemented)
		admin.PUT("/admin/danmaku/:id/block", danmakuhandler.BlockHandler(svcCtx))
	}

	// WebSocket
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
