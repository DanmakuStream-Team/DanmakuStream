package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const middlewareTestSecret = "middleware-test-secret"

func signedToken(t *testing.T, userID uint, role string, expiresAt time.Time) string {
	t.Helper()
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(middlewareTestSecret))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func authRouter(role string) *gin.Engine {
	r := gin.New()
	r.Use(Auth(middlewareTestSecret))
	if role == "staff" {
		r.Use(Staff)
	}
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"userId": UserID(c)})
	})
	return r
}

func TestAuthAcceptsValidTokenAndPublishesClaims(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+signedToken(t, 42, "user", time.Now().Add(time.Hour)))
	w := httptest.NewRecorder()
	authRouter("").ServeHTTP(w, req)
	if w.Code != http.StatusOK || w.Body.String() != `{"userId":42}` {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestAuthRejectsInvalidTokens(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name  string
		token string
	}{
		{name: "missing"},
		{name: "malformed", token: "not-a-jwt"},
		{name: "expired", token: signedToken(t, 42, "user", time.Now().Add(-time.Hour))},
		{name: "zero user", token: signedToken(t, 0, "user", time.Now().Add(time.Hour))},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tc.token != "" {
				req.Header.Set("Authorization", "Bearer "+tc.token)
			}
			w := httptest.NewRecorder()
			authRouter("").ServeHTTP(w, req)
			if w.Code != http.StatusUnauthorized {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestStaffAllowsStaffRolesOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		role string
		want int
	}{{"admin", http.StatusOK}, {"moderator", http.StatusOK}, {"user", http.StatusForbidden}} {
		t.Run(tc.role, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+signedToken(t, 7, tc.role, time.Now().Add(time.Hour)))
			w := httptest.NewRecorder()
			authRouter("staff").ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}

func TestInternalRequiresExactConfiguredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		name     string
		expected string
		provided string
		want     int
	}{{"valid", "internal", "internal", http.StatusOK}, {"wrong", "internal", "other", http.StatusForbidden}, {"empty config", "", "", http.StatusForbidden}} {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(Internal(tc.expected))
			r.GET("/internal", func(c *gin.Context) { c.Status(http.StatusOK) })
			req := httptest.NewRequest(http.MethodGet, "/internal", nil)
			if tc.provided != "" {
				req.Header.Set("X-Internal-Token", tc.provided)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}
