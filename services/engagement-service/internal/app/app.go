package app

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"

	"danmakustream/engagement-service/internal/client"
	"danmakustream/engagement-service/internal/config"
	"danmakustream/engagement-service/internal/database"
	"danmakustream/engagement-service/internal/handler"
	"danmakustream/engagement-service/internal/middleware"
	"danmakustream/engagement-service/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Config  config.Config
	DB      *gorm.DB
	Users   *client.Client
	Content *client.Client
	Hub     *handler.Hub
}

func New(c config.Config, db *gorm.DB) *App {
	return &App{Config: c, DB: db, Users: client.New(c.Dependencies.UserBaseURL, c.Dependencies.InternalToken, c.Dependencies.Timeout), Content: client.New(c.Dependencies.ContentBaseURL, c.Dependencies.InternalToken, c.Dependencies.Timeout), Hub: handler.NewHub()}
}

func (a *App) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), a.requestLog())
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "Authorization,Content-Type,X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	h := handler.New(a.DB, a.Users, a.Content, a.Hub, a.Config)
	v1 := r.Group("/api/v1")
	v1.GET("/livez", func(c *gin.Context) { response.OK(c, gin.H{"status": "up"}) })
	v1.GET("/health", func(c *gin.Context) {
		if a.DB == nil || database.Ping(a.DB) != nil {
			response.Error(c, http.StatusServiceUnavailable, "engagement_db unavailable")
			return
		}
		response.OK(c, gin.H{"status": "up", "database": "up", "dependencies": gin.H{"user-service": "unchecked", "content-service": "unchecked"}})
	})
	v1.GET("/version", func(c *gin.Context) {
		response.OK(c, gin.H{"service": a.Config.Name, "version": defaultValue(a.Config.Build.Version, "microservice-0.1.0"), "commit": defaultValue(a.Config.Build.GitSHA, "0000000"), "buildTime": defaultValue(a.Config.Build.Time, "1970-01-01T00:00:00Z")})
	})

	v1.GET("/danmaku/:videoId", h.ListDanmaku)
	v1.GET("/comments/:videoId", h.ListComments)
	v1.GET("/live", h.ListLiveRooms)
	v1.GET("/live/rankings/heat", h.HeatRanking)
	v1.GET("/live-schedules", middleware.OptionalAuth(a.Config.Auth.AccessSecret), h.ListSchedules)
	v1.GET("/live/:id", h.GetLiveRoom)
	v1.GET("/live/:id/interaction", h.LiveInteraction)
	v1.GET("/live/:id/danmaku", h.CurrentLiveDanmaku)
	internal := r.Group("/internal/v1")
	internal.Use(middleware.Internal(a.Config.Dependencies.InternalToken))
	internal.POST("/live/hooks/srs", h.SRSHook)

	auth := v1.Group("")
	auth.Use(middleware.Auth(a.Config.Auth.AccessSecret))
	auth.POST("/danmaku", h.SendDanmaku)
	auth.POST("/comments", h.CreateComment)
	auth.DELETE("/comments/:id", h.DeleteComment)
	auth.POST("/comments/:id/like", h.ToggleCommentLike)
	auth.POST("/videos/:id/like", h.ToggleVideoLike)
	auth.POST("/videos/:id/collect", h.ToggleVideoCollection)
	auth.GET("/users/me/history", h.ListHistory)
	auth.GET("/users/me/history/:videoId", h.GetHistory)
	auth.PUT("/users/me/history/:videoId", h.SaveHistory)
	auth.DELETE("/users/me/history/:videoId", h.DeleteHistory)
	auth.DELETE("/users/me/history", h.ClearHistory)
	auth.GET("/users/me/watch-later", h.ListWatchLater)
	auth.GET("/users/me/watch-later/:videoId/status", h.WatchLaterStatus)
	auth.POST("/users/me/watch-later/:videoId", h.ToggleWatchLater)
	auth.DELETE("/users/me/watch-later/:videoId", h.DeleteWatchLater)
	auth.DELETE("/users/me/watch-later", h.ClearWatchLater)
	auth.GET("/users/me/collections", h.ListVideoCollections)
	auth.POST("/live", h.CreateLiveRoom)
	auth.GET("/live/:id/manage", h.ManageLiveRoom)
	auth.GET("/live/:id/monitor", h.MonitorLiveRoom)
	auth.PUT("/live/:id/chat-settings", h.UpdateChatSettings)
	auth.PUT("/live/:id/end", h.EndLiveRoom)
	auth.POST("/live/:id/like", h.ToggleLiveLike)
	auth.GET("/live/:id/like/status", h.LiveLikeStatus)
	auth.POST("/live/:id/gifts", h.SendGift)
	auth.POST("/live-schedules", h.CreateSchedule)
	auth.DELETE("/live-schedules/:id", h.CancelSchedule)
	auth.POST("/live-schedules/:id/reserve", h.ToggleReservation)

	staff := v1.Group("")
	staff.Use(middleware.Auth(a.Config.Auth.AccessSecret), middleware.Staff)
	staff.GET("/admin/danmaku", h.AdminListDanmaku)
	staff.PUT("/admin/danmaku/:id/block", h.BlockDanmaku)

	r.GET("/ws/live/:id", h.LiveWebSocket)
	r.GET("/ws/live-publish/:id", h.LivePublishWebSocket)
	return r
}

func (a *App) requestLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			id = randomID()
		}
		c.Header("X-Request-ID", id)
		c.Request = c.Request.WithContext(client.WithRequestID(c.Request.Context(), id))
		started := time.Now()
		c.Next()
		log.Printf("level=info service=%s requestId=%s method=%s path=%s status=%d latencyMs=%d userId=%d", a.Config.Name, id, c.Request.Method, c.Request.URL.Path, c.Writer.Status(), time.Since(started).Milliseconds(), middleware.UserID(c))
	}
}
func randomID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprint(time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
func defaultValue(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
