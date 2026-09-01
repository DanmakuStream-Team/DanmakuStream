package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestUserAdaptersMatchUserServiceContract(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "token" {
			t.Fatal("missing internal token")
		}
		if r.Header.Get("X-Request-ID") != "req-user" {
			t.Fatal("missing request id")
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/internal/v1/users":
			if got := r.URL.Query()["id"]; len(got) != 1 || got[0] != "7" {
				t.Fatalf("ids=%v", got)
			}
			_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"items":[{"id":7,"username":"viewer","nickname":"Viewer"}]}}`))
		case "/internal/v1/relationships/blocked":
			if r.URL.Query().Get("firstId") != "7" || r.URL.Query().Get("secondId") != "9" {
				t.Fatalf("blocked query=%s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"blocked":true}}`))
		case "/internal/v1/memberships/status":
			if r.URL.Query().Get("subscriberId") != "7" || r.URL.Query().Get("creatorId") != "9" {
				t.Fatalf("membership query=%s", r.URL.RawQuery)
			}
			_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"active":true}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	c := New(server.URL, "token", time.Second)
	ctx := WithRequestID(context.Background(), "req-user")
	user, err := c.User(ctx, 7)
	if err != nil || user.ID != 7 || user.Nickname != "Viewer" {
		t.Fatalf("user=%+v err=%v", user, err)
	}
	blocked, err := c.IsBlocked(ctx, 7, 9)
	if err != nil || !blocked {
		t.Fatalf("blocked=%v err=%v", blocked, err)
	}
	member, err := c.IsMember(ctx, 7, 9)
	if err != nil || !member {
		t.Fatalf("member=%v err=%v", member, err)
	}
}

func TestUserMissingFromBatchIsNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"items":[]}}`))
	}))
	defer server.Close()
	_, err := New(server.URL, "token", time.Second).User(context.Background(), 99)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
}

func TestVideoUsesInternalRouteTokenAndCompatibilityFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/videos/7" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if r.Header.Get("X-Internal-Token") != "token" || r.Header.Get("X-Request-ID") != "req-7" {
			t.Fatalf("headers=%v", r.Header)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"id":7,"authorId":11,"status":"approved","transcodeStatus":"ready"}}`))
	}))
	defer server.Close()
	c := New(server.URL, "token", time.Second)
	video, err := c.Video(WithRequestID(context.Background(), "req-7"), 7)
	if err != nil || video.ID != 7 || video.CreatorID != 11 || !video.Playable {
		t.Fatalf("video=%+v err=%v", video, err)
	}
}

func TestVideosUsesBatchRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/v1/videos/batch" || r.URL.Query().Get("ids") != "7,8" {
			t.Fatalf("url=%s", r.URL.String())
		}
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"items":[{"id":7,"creatorId":3,"playable":true},{"id":8,"authorId":4,"status":"approved"}]}}`))
	}))
	defer server.Close()
	videos, err := New(server.URL, "token", time.Second).Videos(context.Background(), []uint{7, 8})
	if err != nil || len(videos) != 2 || videos[1].CreatorID != 4 || !videos[1].Playable {
		t.Fatalf("videos=%+v err=%v", videos, err)
	}
}

func TestDependencyStatusClassification(t *testing.T) {
	cases := []struct {
		status int
		want   error
	}{
		{http.StatusNotFound, ErrNotFound},
		{http.StatusBadGateway, ErrBadGateway},
		{http.StatusServiceUnavailable, ErrUnavailable},
		{http.StatusGatewayTimeout, ErrTimeout},
	}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				http.Error(w, "downstream failed", tc.status)
			}))
			defer server.Close()
			_, err := New(server.URL, "token", time.Second).Video(context.Background(), 9)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want=%v", err, tc.want)
			}
		})
	}
}

func TestEnvelopeFailureAndMalformedResponseAreTyped(t *testing.T) {
	tests := []struct {
		name string
		body string
		want error
	}{
		{"business timeout", `{"code":50400,"message":"timeout","data":null}`, ErrTimeout},
		{"business not found", `{"code":40401,"message":"missing","data":null}`, ErrNotFound},
		{"malformed", `{`, ErrBadGateway},
		{"missing data", `{"code":0,"message":"ok"}`, ErrBadGateway},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(tc.body))
			}))
			defer server.Close()
			_, err := New(server.URL, "token", time.Second).Video(context.Background(), 9)
			if !errors.Is(err, tc.want) {
				t.Fatalf("err=%v want=%v", err, tc.want)
			}
		})
	}
}

func TestMissingInternalRouteIsBadGateway(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"code":40400,"message":"route not found","requestId":"req-route"}`))
	}))
	defer server.Close()
	_, err := New(server.URL, "token", time.Second).Video(context.Background(), 7)
	if !errors.Is(err, ErrBadGateway) {
		t.Fatalf("err=%v", err)
	}
}

func TestTransportTimeoutIsTyped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(100 * time.Millisecond)
		_, _ = w.Write([]byte(`{"code":0,"message":"ok","data":{"id":7}}`))
	}))
	defer server.Close()
	_, err := New(server.URL, "token", 20*time.Millisecond).Video(context.Background(), 7)
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("err=%v", err)
	}
}
