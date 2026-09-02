package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"danmakustream/engagement-service/internal/config"
	"danmakustream/engagement-service/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestIDParamValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		raw    string
		wantID uint
		wantOK bool
	}{{"17", 17, true}, {"0", 0, false}, {"-1", 0, false}, {"bad", 0, false}} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: tc.raw}}
		got, ok := idParam(c, "id")
		if got != tc.wantID || ok != tc.wantOK {
			t.Fatalf("raw=%q got=(%d,%v) want=(%d,%v)", tc.raw, got, ok, tc.wantID, tc.wantOK)
		}
		if !tc.wantOK && w.Code != http.StatusBadRequest {
			t.Fatalf("raw=%q status=%d", tc.raw, w.Code)
		}
	}
}

func TestPageNormalizesInvalidValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		query    string
		wantPage int
		wantSize int
	}{{"", 1, 20}, {"?page=3&pageSize=50", 3, 50}, {"?page=0&pageSize=0", 1, 20}, {"?page=bad&pageSize=101", 1, 20}} {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/"+tc.query, nil)
		gotPage, gotSize := page(c)
		if gotPage != tc.wantPage || gotSize != tc.wantSize {
			t.Fatalf("query=%q got=(%d,%d) want=(%d,%d)", tc.query, gotPage, gotSize, tc.wantPage, tc.wantSize)
		}
	}
}

func TestWebSocketUserIDSupportsQueryAndBearerToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "ws-test-secret"
	claims := middleware.Claims{UserID: 9, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	raw, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	h := &Handler{cfg: config.Config{Auth: struct {
		AccessSecret string `yaml:"AccessSecret"`
	}{AccessSecret: secret}}}
	for _, tc := range []struct {
		name   string
		url    string
		header string
		ok     bool
	}{{"query", "/ws?token=" + raw, "", true}, {"bearer", "/ws", "Bearer " + raw, true}, {"invalid", "/ws?token=bad", "", false}} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, tc.url, nil)
			if tc.header != "" {
				c.Request.Header.Set("Authorization", tc.header)
			}
			uid, ok := h.wsUserID(c)
			if ok != tc.ok || (ok && uid != 9) {
				t.Fatalf("uid=%d ok=%v", uid, ok)
			}
		})
	}
}

func TestDefaultHostRejectsSchemes(t *testing.T) {
	if got := defaultHost("", "srs:1935"); got != "srs:1935" {
		t.Fatalf("empty host=%q", got)
	}
	if got := defaultHost("rtmp://bad", "srs:1935"); got != "srs:1935" {
		t.Fatalf("scheme host=%q", got)
	}
	if got := defaultHost("media:1935", "srs:1935"); got != "media:1935" {
		t.Fatalf("valid host=%q", got)
	}
}
