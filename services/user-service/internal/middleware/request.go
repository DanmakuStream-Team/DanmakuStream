package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"time"
)

const CtxKeyRequestID = "requestId"

func RequestContext(service string) gin.HandlerFunc {
	return func(c *gin.Context) {
		started := time.Now()
		id := c.GetHeader("X-Request-ID")
		if id == "" {
			b := make([]byte, 12)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set(CtxKeyRequestID, id)
		c.Header("X-Request-ID", id)
		c.Next()
		gin.DefaultWriter.Write([]byte(time.Now().Format(time.RFC3339) + " service=" + service + " requestId=" + id + " method=" + c.Request.Method + " path=" + c.Request.URL.Path + " status=" + fmtInt(c.Writer.Status()) + " latencyMs=" + fmtInt(int(time.Since(started).Milliseconds())) + "\n"))
	}
}

func fmtInt(value int) string {
	if value == 0 {
		return "0"
	}
	b := make([]byte, 0, 12)
	for value > 0 {
		b = append([]byte{byte('0' + value%10)}, b...)
		value /= 10
	}
	return string(b)
}

func InternalAuth(token string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if token == "" || c.GetHeader("X-Internal-Token") != token {
			c.AbortWithStatusJSON(403, gin.H{"code": 40301, "message": "invalid internal credential", "requestId": c.GetString(CtxKeyRequestID)})
			return
		}
		c.Next()
	}
}
