package handler

import (
	"context"
	"net/http"

	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) Livez(c *gin.Context) {
	response.OK(c, gin.H{"status": "up"})
}

func (h *Handler) Health(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	if err == nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), h.Config.RequestTimeout)
		defer cancel()
		err = sqlDB.PingContext(ctx)
	}
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code": 50300, "message": "database unavailable", "requestId": c.GetString("requestId"),
			"data": gin.H{"status": "down", "database": "down", "dependencies": gin.H{}},
		})
		return
	}
	response.OK(c, gin.H{"status": "up", "database": "up", "dependencies": gin.H{}})
}

func (h *Handler) Version(c *gin.Context) {
	response.OK(c, gin.H{
		"service": h.Config.ServiceName, "version": h.Config.ServiceVersion,
		"commit": h.Config.CommitSHA, "buildTime": h.Config.BuildTime,
	})
}
