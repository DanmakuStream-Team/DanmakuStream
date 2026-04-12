package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT and injects user claims into context.
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				httpx.Error(w, ErrUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims := &Claims{}
			token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !token.Valid {
				httpx.Error(w, ErrUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), CtxKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, CtxKeyUsername, claims.Username)
			ctx = context.WithValue(ctx, CtxKeyRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminMiddleware ensures the user has admin role.
func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role, _ := r.Context().Value(CtxKeyRole).(string)
		if role != "admin" {
			httpx.Error(w, ErrForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Context keys
type ctxKey string

const (
	CtxKeyUserID   ctxKey = "userId"
	CtxKeyUsername ctxKey = "username"
	CtxKeyRole     ctxKey = "role"
)

// Sentinel errors
var (
	ErrUnauthorized = httpError{code: 401, message: "未授权，请先登录"}
	ErrForbidden    = httpError{code: 403, message: "权限不足"}
)

type httpError struct {
	code    int
	message string
}

func (e httpError) Error() string { return e.message }
