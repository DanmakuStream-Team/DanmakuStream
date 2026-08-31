package app

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"danmakustream/engagement-service/internal/config"
)

func testConfig() config.Config {
	var c config.Config
	c.Name = "engagement-service"
	c.Auth.AccessSecret = "secret"
	c.Dependencies.InternalToken = "token"
	c.Dependencies.Timeout = time.Second
	c.Build.Version = "microservice-0.1.0"
	c.Build.GitSHA = "abcdef0"
	c.Build.Time = "2026-08-31T10:00:00+08:00"
	return c
}

func TestLivezContract(t *testing.T) {
	r := New(testConfig(), nil).Router()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/livez", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d", w.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["code"] != float64(0) || body["message"] != "ok" {
		t.Fatalf("body=%s", w.Body.String())
	}
}

func TestHealthReportsDatabaseUnavailable(t *testing.T) {
	r := New(testConfig(), nil).Router()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("X-Request-ID", "req-1")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"code":50300`) || !strings.Contains(w.Body.String(), `"requestId":"req-1"`) {
		t.Fatalf("body=%s", w.Body.String())
	}
}

func TestVersionFieldsAreNonEmpty(t *testing.T) {
	r := New(testConfig(), nil).Router()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/v1/version", nil))
	for _, value := range []string{"engagement-service", "microservice-0.1.0", "abcdef0", "2026-08-31"} {
		if !strings.Contains(w.Body.String(), value) {
			t.Fatalf("missing %s: %s", value, w.Body.String())
		}
	}
}

func TestProtectedRouteRejectsMissingJWT(t *testing.T) {
	r := New(testConfig(), nil).Router()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/v1/videos/7/like", nil))
	if w.Code != http.StatusUnauthorized || !strings.Contains(w.Body.String(), `"code":40100`) {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
func TestInternalHookRejectsMissingToken(t *testing.T) {
	r := New(testConfig(), nil).Router()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/internal/v1/live/hooks/srs", nil))
	if w.Code != http.StatusForbidden || !strings.Contains(w.Body.String(), `"code":40300`) {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
}
