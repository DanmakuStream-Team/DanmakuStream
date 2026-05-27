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

	// 用户头像单独放一个目录，方便管理和访问
	r.Static("/media/avatars", svcCtx.VideoDir+"/avatars")

	authMW := middleware.AuthMiddleware(c.Auth.AccessSecret)

	v1 := r.Group("/api/v1")
	{
		// Public routes
		v1.POST("/auth/login", authhandler.LoginHandler(svcCtx))
		v1.POST("/auth/register", authhandler.RegisterHandler(svcCtx))
		v1.GET("/videos", videohandler.ListHandler(svcCtx))
		v1.GET("/videos/:id", videohandler.DetailHandler(svcCtx))
		v1.GET("/danmaku/:videoId", danmakuhandler.ListHandler(svcCtx))
		v1.GET("/users/:id", userhandler.ProfileHandler(svcCtx))
		// 获取用户主页信息；未登录也可访问，若携带有效 Bearer Token，额外返回当前用户是否已关注该用户 followed。
		v1.GET("/users/:id/videos", userhandler.VideosHandler(svcCtx))
		// 获取用户发布的视频
		v1.GET("/comments/:videoId", commenthandler.ListHandler(svcCtx))
		// 获取视频评论列表，支持分页；未登录也可访问，若携带有效 Bearer Token，额外返回当前用户对每条评论是否已点赞以及点赞数
		//现在是返回所有评论，后面再加分页和排序功能
	}

	// Auth-required routes
	auth := v1.Group("")
	auth.Use(authMW)
	{
		auth.GET("/auth/me", authhandler.MeHandler(svcCtx))
		auth.PUT("/users/me", userhandler.UpdateMeHandler(svcCtx))
		auth.POST("/users/me/avatar", userhandler.UploadAvatarHandler(svcCtx))
		// 上面两个是用户个人中心相关的接口，更新简介和头像
		// 暂时不允许清空简介,后面再改
		auth.GET("/users/me/videos", userhandler.MeVideosHandler(svcCtx))
		// 获取当前用户发布的视频,各种状态的都能看
		auth.POST("/videos/upload", videohandler.UploadHandler(svcCtx))
		auth.PUT("/videos/:id", videohandler.UpdateHandler(svcCtx))
		// 更新视频信息，只有标题、简介和标签可以更新，且只能更新自己发布的视频
		// 暂时不允许不允许把 description 或 tags 清空，后面再改
		auth.POST("/videos/:id/cover", videohandler.UpdateCoverHandler(svcCtx))
		// 更新视频封面，单独一个接口，方便前端上传和裁剪封面图片
		auth.DELETE("/videos/:id", videohandler.DeleteHandler(svcCtx))
		// 删除视频，只有自己发布的视频才能删除，删除后视频文件和封面文件会被删除，相关的弹幕、评论、点赞、收藏记录暂时不会删除
		// 有孤儿数据，但不影响功能，后面再改
		auth.POST("/videos/:id/like", videohandler.LikeHandler(svcCtx))
		auth.POST("/videos/:id/collect", videohandler.CollectHandler(svcCtx))
		auth.POST("/danmaku", danmakuhandler.SendHandler(svcCtx))
		auth.POST("/comments", commenthandler.CreateHandler(svcCtx))
		auth.DELETE("/comments/:id", commenthandler.DeleteHandler(svcCtx))
		// 		评论作者本人可以删除
		// 管理员可以删除
		// 普通用户不能删别人的评论
		// 删除一级评论时，回复是否一起删先不处理，第一版用软删除当前评论
		auth.POST("/comments/:id/like", commenthandler.LikeHandler(svcCtx))
		// 评论点赞接口
		auth.GET("/users/following", userhandler.FollowingListHandler(svcCtx))
				auth.POST("/users/:id/follow", userhandler.FollowHandler(svcCtx))
		auth.GET("/live", notImplemented)
		auth.POST("/live", notImplemented)
		auth.PUT("/live/:id/end", notImplemented)
	}

	// Admin routes
	admin := v1.Group("")
	admin.Use(authMW, middleware.AdminMiddleware)
	{
		admin.GET("/admin/videos", videohandler.AdminListHandler(svcCtx))
		admin.PUT("/admin/videos/:id/status", videohandler.AdminUpdateStatusHandler(svcCtx))
		admin.GET("/admin/danmaku", danmakuhandler.AdminListHandler(svcCtx))
		admin.PUT("/admin/danmaku/:id/block", danmakuhandler.BlockHandler(svcCtx))
		// 管理员可查看所有弹幕。这个功能直接分开好了
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
