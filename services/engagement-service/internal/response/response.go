package response

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, data any) { c.JSON(200, gin.H{"code": 0, "message": "ok", "data": data}) }
func Error(c *gin.Context, status int, message string) {
	code := status * 100
	requestID := c.Writer.Header().Get("X-Request-ID")
	c.JSON(status, gin.H{"code": code, "message": message, "requestId": requestID})
}
