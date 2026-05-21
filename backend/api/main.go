package main

import (
	"flag"
	"fmt"
	"net/http"

	"danmakustream/backend/internal/config"
	authhandler "danmakustream/backend/internal/handler/v1/auth"
	videohandler "danmakustream/backend/internal/handler/v1/video"
	wshandler "danmakustream/backend/internal/handler/ws"
	"danmakustream/backend/internal/middleware"
	"danmakustream/backend/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)	

var configFile = flag.String("f", "etc/config.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()

	ctx := svc.NewServiceContext(c)

	authMW := middleware.AuthMiddleware(c.Auth.AccessSecret)
	adminMW := middleware.AdminMiddleware

	// ─── Public routes ────────────────────────────────────────────────
	server.AddRoutes([]rest.Route{
		{Method: http.MethodPost, Path: "/api/v1/auth/login", Handler: authhandler.LoginHandler(ctx)},
		{Method: http.MethodPost, Path: "/api/v1/auth/register", Handler: authhandler.RegisterHandler(ctx)},
		{Method: http.MethodGet, Path: "/api/v1/videos", Handler: videohandler.ListHandler(ctx)},
		{Method: http.MethodGet, Path: "/api/v1/videos/:id", Handler: videohandler.DetailHandler(ctx)},
		{Method: http.MethodGet, Path: "/api/v1/danmaku/:videoId", Handler: danmakuListHandler(ctx)},
		{Method: http.MethodGet, Path: "/api/v1/users/:id", Handler: userProfileHandler(ctx)},
	})

	// ─── Auth-required routes ─────────────────────────────────────────
	server.AddRoutes(rest.WithMiddlewares([]rest.Middleware{authMW},
		[]rest.Route{
			{Method: http.MethodGet, Path: "/api/v1/auth/me", Handler: authhandler.MeHandler(ctx)},
			{Method: http.MethodPost, Path: "/api/v1/videos/upload", Handler: videoUploadHandler(ctx)},
			{Method: http.MethodPut, Path: "/api/v1/videos/:id", Handler: videoUpdateHandler(ctx)},
			{Method: http.MethodPost, Path: "/api/v1/videos/:id/like", Handler: videoLikeHandler(ctx)},
			{Method: http.MethodPost, Path: "/api/v1/videos/:id/collect", Handler: videoCollectHandler(ctx)},
			{Method: http.MethodPost, Path: "/api/v1/danmaku", Handler: danmakuSendHandler(ctx)},
			{Method: http.MethodPost, Path: "/api/v1/comments", Handler: commentCreateHandler(ctx)},
			{Method: http.MethodPost, Path: "/api/v1/users/:id/follow", Handler: followHandler(ctx)},
			{Method: http.MethodGet, Path: "/api/v1/live", Handler: liveListHandler(ctx)},
			{Method: http.MethodPost, Path: "/api/v1/live", Handler: liveStartHandler(ctx)},
			{Method: http.MethodPut, Path: "/api/v1/live/:id/end", Handler: liveEndHandler(ctx)},
		}...))

	// ─── Admin routes ─────────────────────────────────────────────────
	server.AddRoutes(rest.WithMiddlewares([]rest.Middleware{authMW, adminMW},
		[]rest.Route{
			{Method: http.MethodGet, Path: "/api/v1/admin/videos", Handler: adminVideoListHandler(ctx)},
			{Method: http.MethodPut, Path: "/api/v1/admin/videos/:id/status", Handler: adminVideoStatusHandler(ctx)},
			{Method: http.MethodPut, Path: "/api/v1/admin/danmaku/:id/block", Handler: adminDanmakuBlockHandler(ctx)},
		}...))

	// ─── WebSocket ────────────────────────────────────────────────────
	server.AddRoute(rest.Route{
		Method:  http.MethodGet,
		Path:    "/ws/live/:id",
		Handler: wshandler.LiveWebSocketHandler(ctx),
	})

	fmt.Printf("DanmakuStream API server starting on :%d\n", c.Port)
	server.Start()
}

// Placeholder handler functions (to be implemented in respective handler packages)
func videoUploadHandler(ctx *svc.ServiceContext) http.HandlerFunc       { return notImplemented }
func videoUpdateHandler(ctx *svc.ServiceContext) http.HandlerFunc       { return notImplemented }
func videoLikeHandler(ctx *svc.ServiceContext) http.HandlerFunc         { return notImplemented }
func videoCollectHandler(ctx *svc.ServiceContext) http.HandlerFunc      { return notImplemented }
func danmakuListHandler(ctx *svc.ServiceContext) http.HandlerFunc       { return notImplemented }
func danmakuSendHandler(ctx *svc.ServiceContext) http.HandlerFunc       { return notImplemented }
func commentCreateHandler(ctx *svc.ServiceContext) http.HandlerFunc     { return notImplemented }
func userProfileHandler(ctx *svc.ServiceContext) http.HandlerFunc       { return notImplemented }
func getMeHandler(ctx *svc.ServiceContext) http.HandlerFunc             { return notImplemented }
func followHandler(ctx *svc.ServiceContext) http.HandlerFunc            { return notImplemented }
func liveListHandler(ctx *svc.ServiceContext) http.HandlerFunc          { return notImplemented }
func liveStartHandler(ctx *svc.ServiceContext) http.HandlerFunc         { return notImplemented }
func liveEndHandler(ctx *svc.ServiceContext) http.HandlerFunc           { return notImplemented }
func adminVideoListHandler(ctx *svc.ServiceContext) http.HandlerFunc    { return notImplemented }
func adminVideoStatusHandler(ctx *svc.ServiceContext) http.HandlerFunc  { return notImplemented }
func adminDanmakuBlockHandler(ctx *svc.ServiceContext) http.HandlerFunc { return notImplemented }

func notImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"code":501,"message":"not implemented yet"}`))
}
