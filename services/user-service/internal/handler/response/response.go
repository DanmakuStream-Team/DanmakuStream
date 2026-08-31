package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data":    data,
	})
}

func Fail(c *gin.Context, httpCode int, msg string) {
	code := httpCode * 100
	c.JSON(httpCode, gin.H{
		"code":      code,
		"message":   msg,
		"requestId": c.GetString("requestId"),
	})
}
