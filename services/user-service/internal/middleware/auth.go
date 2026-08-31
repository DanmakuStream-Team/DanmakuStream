package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

const (
	CtxKeyUserID   = "userId"
	CtxKeyUsername = "username"
	CtxKeyRole     = "role"
)

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":      40100,
				"message":   "未授权，请先登录",
				"requestId": c.GetString(CtxKeyRequestID),
			})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid || claims.UserID == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":      40100,
				"message":   "未授权，请先登录",
				"requestId": c.GetString(CtxKeyRequestID),
			})
			return
		}

		c.Set(CtxKeyUserID, claims.UserID)
		c.Set(CtxKeyUsername, claims.Username)
		c.Set(CtxKeyRole, claims.Role)
		c.Next()
	}
}

func AdminMiddleware(c *gin.Context) {
	role, _ := c.Get(CtxKeyRole)
	if role != "admin" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":      40300,
			"message":   "权限不足",
			"requestId": c.GetString(CtxKeyRequestID),
		})
		return
	}
	c.Next()
}

func StaffMiddleware(c *gin.Context) {
	role, _ := c.Get(CtxKeyRole)
	if role != "admin" && role != "moderator" {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"code":      40300,
			"message":   "权限不足",
			"requestId": c.GetString(CtxKeyRequestID),
		})
		return
	}
	c.Next()
}
