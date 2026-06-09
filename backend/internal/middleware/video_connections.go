package middleware

import (
	"strings"

	"danmakustream/backend/internal/metrics"

	"github.com/gin-gonic/gin"
)

func VideoConnectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/media/videos/") {
			c.Next()
			return
		}

		metrics.BeginVideoConnection()
		defer metrics.EndVideoConnection()
		c.Next()
	}
}
