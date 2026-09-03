package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"danmakustream/content-service/internal/config"
	"danmakustream/content-service/internal/logic"
	"github.com/gin-gonic/gin"
)

func TestEnrichVideosUsesInternalUserSummary(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Internal-Token"); got != "test-token" {
			t.Fatalf("internal token = %q", got)
		}
		if got := r.URL.Query()["id"]; len(got) != 1 || got[0] != "4" {
			t.Fatalf("ids = %#v", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"items":[{"id":4,"username":"4","nickname":"sxh","avatar":"/media/avatars/4.png","role":"creator"}]},"message":"ok"}`))
	}))
	defer server.Close()

	h := Handler{Config: config.Config{UserServiceURL: server.URL, InternalAPIToken: "test-token", RequestTimeout: time.Second}}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/videos/1", nil)
	videos := []logic.VideoDTO{{Author: logic.AuthorDTO{ID: 4}}}
	h.enrichVideos(c, videos)
	if got := videos[0].Author.Nickname; got != "sxh" {
		t.Fatalf("nickname = %q", got)
	}
	if got := videos[0].Author.Role; got != "creator" {
		t.Fatalf("role = %q", got)
	}
}

func TestEnrichVideosFallsBackWhenUserServiceFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	h := Handler{Config: config.Config{UserServiceURL: server.URL, InternalAPIToken: "test-token", RequestTimeout: time.Second}}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/videos", nil)
	videos := []logic.VideoDTO{{Author: logic.AuthorDTO{ID: 9}}}
	h.enrichVideos(c, videos)
	if got := videos[0].Author.Nickname; got != "用户 #9（资料暂不可用）" {
		t.Fatalf("fallback nickname = %q", got)
	}
}

func TestEnrichVideosFallsBackOnTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(30 * time.Millisecond)
		_, _ = w.Write([]byte(`{"data":{"items":[]}}`))
	}))
	defer server.Close()

	h := Handler{Config: config.Config{UserServiceURL: server.URL, InternalAPIToken: "test-token", RequestTimeout: time.Millisecond}}
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/videos", nil)
	videos := []logic.VideoDTO{{Author: logic.AuthorDTO{ID: 7}}}
	h.enrichVideos(c, videos)
	if got := videos[0].Author.Username; got != "user-7" {
		t.Fatalf("fallback username = %q", got)
	}
}
