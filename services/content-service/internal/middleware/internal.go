package middleware

import (
	"crypto/subtle"
	"net/http"

	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
)

// InternalAuth protects service-to-service endpoints with the shared internal token.
func InternalAuth(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		provided := c.GetHeader("X-Internal-Token")
		if expected == "" || len(provided) != len(expected) || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			response.Error(c, http.StatusForbidden, 40302, "invalid internal credential")
			c.Abort()
			return
		}
		c.Next()
	}
}
