package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": data})
}

func Error(c *gin.Context, status, code int, message string) {
	c.JSON(status, gin.H{"code": code, "message": message, "requestId": c.GetString("requestId")})
}
