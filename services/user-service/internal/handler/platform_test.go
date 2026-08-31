package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"danmakustream/user-service/internal/config"
	"danmakustream/user-service/internal/middleware"
	"danmakustream/user-service/internal/svc"
	"github.com/gin-gonic/gin"
)

func decodeBody(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	return body
}

func TestLivezContract(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestContext("user-service"))
	r.GET("/api/v1/livez", Livez)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/livez", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	body := decodeBody(t, w)
	if body["code"] != float64(0) || body["message"] != "ok" || body["data"] == nil {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestVersionContract(t *testing.T) {
	ctx := &svc.ServiceContext{Config: config.Config{Name: "user-service", Version: "1.2.3", Commit: "abc", BuildTime: "now"}}
	r := gin.New()
	r.Use(middleware.RequestContext("user-service"))
	r.GET("/api/v1/version", Version(ctx))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/version", nil))
	if w.Code != http.StatusOK || decodeBody(t, w)["code"] != float64(0) {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}

func TestHealthFailureHasRequestID(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequestContext("user-service"))
	r.GET("/api/v1/health", Health(&svc.ServiceContext{}))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Request-ID", "contract-test")
	r.ServeHTTP(w, req)
	body := decodeBody(t, w)
	if w.Code != http.StatusServiceUnavailable || body["requestId"] != "contract-test" || body["code"] == float64(0) {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestInternalToken(t *testing.T) {
	r := gin.New()
	r.Use(middleware.RequestContext("user-service"))
	r.GET("/internal/v1/ping", middleware.InternalAuth("secret"), func(c *gin.Context) { c.Status(http.StatusNoContent) })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/internal/v1/ping", nil))
	if w.Code != http.StatusForbidden {
		t.Fatalf("status=%d", w.Code)
	}
	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/v1/ping", nil)
	req.Header.Set("X-Internal-Token", "secret")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status=%d", w.Code)
	}
}
