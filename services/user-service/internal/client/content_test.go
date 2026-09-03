package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestContentVideoContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/videos/7" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("X-Internal-Token") != "token" {
			t.Fatal("missing internal token")
		}
		if r.Header.Get("X-Request-ID") != "request-7" {
			t.Fatal("missing request id")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"id":7,"creatorId":3,"title":"demo","coverUrl":"/cover.jpg","duration":60,"status":"published","playable":true}}`))
	}))
	defer server.Close()
	video, err := NewContent(server.URL, "token", time.Second).Video(context.Background(), 7, "request-7")
	if err != nil {
		t.Fatal(err)
	}
	if video.ID != 7 || video.CreatorID != 3 || !video.Playable {
		t.Fatalf("video=%+v", video)
	}
}

func TestContentVideoErrorMapping(t *testing.T) {
	for _, test := range []struct {
		name    string
		handler http.HandlerFunc
		want    error
	}{
		{name: "not found", handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) }, want: ErrNotFound},
		{name: "invalid json", handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("{")) }, want: ErrBadGateway},
		{name: "business error", handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(`{"code":40401,"message":"not found"}`)) }, want: ErrBadGateway},
		{name: "credential rejected", handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusForbidden) }, want: ErrBadGateway},
		{name: "unavailable", handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }, want: ErrUnavailable},
		{name: "timeout", handler: func(w http.ResponseWriter, _ *http.Request) { time.Sleep(50 * time.Millisecond) }, want: ErrTimeout},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(test.handler)
			defer server.Close()
			_, err := NewContent(server.URL, "token", 10*time.Millisecond).Video(context.Background(), 1, "")
			if !errors.Is(err, test.want) {
				t.Fatalf("error=%v want=%v", err, test.want)
			}
		})
	}
}

func TestContentVideosBatchContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/videos/batch" || r.URL.Query().Get("ids") != "2,5" {
			t.Fatalf("url=%s", r.URL.String())
		}
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"items":[{"id":2,"playable":true},{"id":5,"playable":true}]}}`))
	}))
	defer server.Close()
	videos, err := NewContent(server.URL, "token", time.Second).Videos(context.Background(), []uint{2, 5}, "")
	if err != nil || len(videos) != 2 {
		t.Fatalf("videos=%+v error=%v", videos, err)
	}
}
