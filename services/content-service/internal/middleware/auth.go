package middleware

import (
	"net/http"
	"strings"

	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextUserID = "userId"
	ContextRole   = "role"
)

type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func Authenticate(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Error(c, http.StatusUnauthorized, 40100, "authentication required")
			c.Abort()
			return
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(strings.TrimPrefix(header, "Bearer "), claims, func(token *jwt.Token) (any, error) {
			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid || claims.UserID == 0 {
			response.Error(c, http.StatusUnauthorized, 40101, "invalid or expired token")
			c.Abort()
			return
		}
		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[role] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := allowed[c.GetString(ContextRole)]; !ok {
			response.Error(c, http.StatusForbidden, 40300, "permission denied")
			c.Abort()
			return
		}
		c.Next()
	}
}

func UserID(c *gin.Context) uint {
	value, _ := c.Get(ContextUserID)
	userID, _ := value.(uint)
	return userID
}
