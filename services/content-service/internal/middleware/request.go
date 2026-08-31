package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			var buf [12]byte
			if _, err := rand.Read(buf[:]); err == nil {
				requestID = hex.EncodeToString(buf[:])
			} else {
				requestID = time.Now().UTC().Format("20060102150405.000000000")
			}
		}
		c.Set("requestId", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func AccessLog(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		c.Next()
		logger.Info("http_request",
			"requestId", c.GetString("requestId"),
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latencyMs", time.Since(started).Milliseconds(),
		)
	}
}
