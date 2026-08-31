package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestVideoUsesInternalRouteAndToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/videos/7" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("X-Internal-Token") != "token" {
			t.Fatal("missing token")
		}
		if r.Header.Get("X-Request-ID") != "req-7" {
			t.Fatal("missing request id")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"id":7,"playable":true}}`))
	}))
	defer server.Close()
	c := New(server.URL, "token", time.Second)
	video, err := c.Video(WithRequestID(context.Background(), "req-7"), 7)
	if err != nil || video.ID != 7 || !video.Playable {
		t.Fatalf("video=%+v err=%v", video, err)
	}
}

func TestNotFoundIsTyped(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	defer server.Close()
	c := New(server.URL, "token", time.Second)
	_, err := c.Video(context.Background(), 9)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
}
