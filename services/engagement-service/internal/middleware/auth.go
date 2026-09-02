package middleware

import (
	"danmakustream/engagement-service/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strings"
)

const UserIDKey = "userId"

type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if raw == "" {
			response.Error(c, 401, "未授权，请先登录")
			c.Abort()
			return
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid || claims.UserID == 0 {
			response.Error(c, 401, "登录状态无效或已过期")
			c.Abort()
			return
		}
		c.Set(UserIDKey, claims.UserID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// OptionalAuth enriches public list endpoints when a valid bearer token is
// present, while retaining anonymous access and never turning a public read
// into an authentication failure.
func OptionalAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		if raw != "" {
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
				if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(secret), nil
			})
			if err == nil && token.Valid && claims.UserID > 0 {
				c.Set(UserIDKey, claims.UserID)
				c.Set("role", claims.Role)
			}
		}
		c.Next()
	}
}
func UserID(c *gin.Context) uint { value, _ := c.Get(UserIDKey); id, _ := value.(uint); return id }
func Staff(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" && role != "moderator" {
		response.Error(c, 403, "权限不足")
		c.Abort()
		return
	}
	c.Next()
}

func Internal(expected string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if expected == "" || c.GetHeader("X-Internal-Token") != expected {
			response.Error(c, 403, "内部调用凭证无效")
			c.Abort()
			return
		}
		c.Next()
	}
}
