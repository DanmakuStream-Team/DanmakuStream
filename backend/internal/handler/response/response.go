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
	c.JSON(httpCode, gin.H{
		"code":    httpCode,
		"message": msg,
	})
}
