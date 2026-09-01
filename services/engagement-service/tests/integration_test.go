package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"danmakustream/engagement-service/internal/app"
	"danmakustream/engagement-service/internal/config"
	"danmakustream/engagement-service/internal/database"
	"danmakustream/engagement-service/internal/middleware"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

const (
	testJWTSecret     = "engagement-integration-jwt"
	testInternalToken = "engagement-integration-internal"
)

type apiEnvelope struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	RequestID string          `json:"requestId"`
	Data      json.RawMessage `json:"data"`
}

func TestEngagementAPIAndWebSocketRegression(t *testing.T) {
	dsn := os.Getenv("ENGAGEMENT_INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("set ENGAGEMENT_INTEGRATION_DSN to run the MySQL integration regression")
	}

	userDependency := fakeUserService(t)
	defer userDependency.Close()
	contentDependency := fakeContentService(t)
	defer contentDependency.Close()

	db, err := database.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()

	var cfg config.Config
	cfg.Name = "engagement-service"
	cfg.Auth.AccessSecret = testJWTSecret
	cfg.Dependencies.UserBaseURL = userDependency.URL
	cfg.Dependencies.ContentBaseURL = contentDependency.URL
	cfg.Dependencies.InternalToken = testInternalToken
	cfg.Dependencies.Timeout = 250 * time.Millisecond
	cfg.Build.Version = "integration"
	cfg.Build.GitSHA = "abcdef0"
	cfg.Build.Time = "2026-09-01T09:00:00+08:00"

	server := httptest.NewServer(app.New(cfg, db).Router())
	defer server.Close()

	seed := uint(time.Now().UnixNano()%1_000_000 + 1_000_000)
	ownerToken := signedToken(t, seed, "creator")
	viewerToken := signedToken(t, seed+1, "user")
	videoID := seed + 2

	t.Run("platform contracts", func(t *testing.T) {
		for _, path := range []string{"/api/v1/livez", "/api/v1/health", "/api/v1/version"} {
			response := callAPI(t, http.MethodGet, server.URL+path, "", nil, http.StatusOK)
			if response.Code != 0 {
				t.Fatalf("path=%s response=%+v", path, response)
			}
		}
	})

	t.Run("video interaction and library", func(t *testing.T) {
		liked := callData(t, http.MethodPost, fmt.Sprintf("%s/api/v1/videos/%d/like", server.URL, videoID), viewerToken, nil, http.StatusOK)
		assertBool(t, liked, "liked", true)
		liked = callData(t, http.MethodPost, fmt.Sprintf("%s/api/v1/videos/%d/like", server.URL, videoID), viewerToken, nil, http.StatusOK)
		assertBool(t, liked, "liked", false)

		collected := callData(t, http.MethodPost, fmt.Sprintf("%s/api/v1/videos/%d/collect", server.URL, videoID), viewerToken, nil, http.StatusOK)
		assertBool(t, collected, "collected", true)

		callAPI(t, http.MethodPost, server.URL+"/api/v1/comments", viewerToken, map[string]any{"videoId": videoID, "content": "integration comment"}, http.StatusOK)
		callAPI(t, http.MethodPost, server.URL+"/api/v1/comments", viewerToken, map[string]any{"videoId": videoID, "content": ""}, http.StatusBadRequest)
		callAPI(t, http.MethodPost, server.URL+"/api/v1/danmaku", viewerToken, map[string]any{"videoId": videoID, "content": "integration danmaku", "time": 12}, http.StatusOK)

		historyURL := fmt.Sprintf("%s/api/v1/users/me/history/%d", server.URL, videoID)
		callAPI(t, http.MethodPut, historyURL, viewerToken, map[string]any{"position": 30}, http.StatusOK)
		callAPI(t, http.MethodPut, historyURL, viewerToken, map[string]any{"position": 45}, http.StatusOK)
		history := callData(t, http.MethodGet, historyURL, viewerToken, nil, http.StatusOK)
		assertNumber(t, history, "position", 45)

		watchLaterURL := fmt.Sprintf("%s/api/v1/users/me/watch-later/%d", server.URL, videoID)
		watchLater := callData(t, http.MethodPost, watchLaterURL, viewerToken, nil, http.StatusOK)
		assertBool(t, watchLater, "saved", true)
		status := callData(t, http.MethodGet, watchLaterURL+"/status", viewerToken, nil, http.StatusOK)
		assertBool(t, status, "saved", true)
		callAPI(t, http.MethodDelete, watchLaterURL, viewerToken, nil, http.StatusOK)
	})

	var roomID uint
	t.Run("live room gifts and WebSocket reconnect", func(t *testing.T) {
		room := callData(t, http.MethodPost, server.URL+"/api/v1/live", ownerToken, map[string]any{"title": "integration live"}, http.StatusOK)
		roomID = uint(number(t, room, "ID"))
		if roomID == 0 {
			t.Fatal("room ID is zero")
		}

		settingsURL := fmt.Sprintf("%s/api/v1/live/%d/chat-settings", server.URL, roomID)
		callAPI(t, http.MethodPut, settingsURL, ownerToken, map[string]any{"chatMode": "followers", "slowModeSeconds": 0, "pinnedMessage": "welcome"}, http.StatusOK)

		likeURL := fmt.Sprintf("%s/api/v1/live/%d/like", server.URL, roomID)
		like := callData(t, http.MethodPost, likeURL, viewerToken, nil, http.StatusOK)
		assertBool(t, like, "liked", true)
		callAPI(t, http.MethodPost, fmt.Sprintf("%s/api/v1/live/%d/gifts", server.URL, roomID), viewerToken, map[string]any{"giftKey": "rocket", "count": 1, "message": "great stream"}, http.StatusOK)

		wsBase := "ws" + strings.TrimPrefix(server.URL, "http")
		wsURL := fmt.Sprintf("%s/ws/live/%d?token=%s", wsBase, roomID, url.QueryEscape(viewerToken))
		first, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer first.Close()
		second, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer second.Close()
		waitForViewerCount(t, server.URL, roomID, 1)

		if err := second.WriteJSON(map[string]any{"type": "danmaku", "content": "ws regression", "time": 3}); err != nil {
			t.Fatal(err)
		}
		if !waitForWebSocketType(second, "danmaku", 2*time.Second) {
			t.Fatal("did not receive WebSocket danmaku broadcast")
		}

		_ = first.Close()
		waitForViewerCount(t, server.URL, roomID, 1)
		_ = second.Close()
		waitForViewerCount(t, server.URL, roomID, 0)

		callAPI(t, http.MethodPut, fmt.Sprintf("%s/api/v1/live/%d/end", server.URL, roomID), ownerToken, nil, http.StatusOK)
	})

	t.Run("schedule conflicts and reservation idempotency", func(t *testing.T) {
		scheduledAt := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
		body := map[string]any{"title": "integration schedule", "scheduledAt": scheduledAt}
		schedule := callData(t, http.MethodPost, server.URL+"/api/v1/live-schedules", ownerToken, body, http.StatusOK)
		scheduleID := uint(number(t, schedule, "ID"))
		callAPI(t, http.MethodPost, server.URL+"/api/v1/live-schedules", ownerToken, body, http.StatusConflict)

		reserveURL := fmt.Sprintf("%s/api/v1/live-schedules/%d/reserve", server.URL, scheduleID)
		reserved := callData(t, http.MethodPost, reserveURL, viewerToken, nil, http.StatusOK)
		assertBool(t, reserved, "reserved", true)
		reserved = callData(t, http.MethodPost, reserveURL, viewerToken, nil, http.StatusOK)
		assertBool(t, reserved, "reserved", false)
		callAPI(t, http.MethodPost, reserveURL, ownerToken, nil, http.StatusConflict)
		callAPI(t, http.MethodDelete, fmt.Sprintf("%s/api/v1/live-schedules/%d", server.URL, scheduleID), ownerToken, nil, http.StatusOK)
	})
}

func fakeUserService(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != testInternalToken || r.Header.Get("X-Request-ID") == "" {
			http.Error(w, "missing internal authentication or request ID", http.StatusForbidden)
			return
		}
		switch r.URL.Path {
		case "/internal/v1/users":
			items := make([]map[string]any, 0, len(r.URL.Query()["id"]))
			for _, raw := range r.URL.Query()["id"] {
				id, _ := strconv.ParseUint(raw, 10, 64)
				items = append(items, map[string]any{"id": id, "username": "integration-user", "nickname": "Integration User", "role": "user"})
			}
			writeOK(w, map[string]any{"items": items})
		case "/internal/v1/relationships/blocked":
			writeOK(w, map[string]any{"blocked": false})
		case "/internal/v1/memberships/status":
			writeOK(w, map[string]any{"active": true})
		case "/internal/v1/relationships/following":
			writeOK(w, map[string]any{"following": true})
		default:
			http.NotFound(w, r)
		}
	}))
}

func fakeContentService(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != testInternalToken || r.Header.Get("X-Request-ID") == "" {
			http.Error(w, "missing internal authentication or request ID", http.StatusForbidden)
			return
		}
		if r.URL.Path == "/internal/v1/videos/batch" {
			parts := strings.Split(r.URL.Query().Get("ids"), ",")
			items := make([]map[string]any, 0, len(parts))
			for _, part := range parts {
				id, _ := strconv.ParseUint(part, 10, 64)
				if id > 0 {
					items = append(items, videoSummary(uint(id)))
				}
			}
			writeOK(w, map[string]any{"items": items})
			return
		}
		const prefix = "/internal/v1/videos/"
		if strings.HasPrefix(r.URL.Path, prefix) {
			id, err := strconv.ParseUint(strings.TrimPrefix(r.URL.Path, prefix), 10, 64)
			if err != nil || id == 0 {
				http.NotFound(w, r)
				return
			}
			writeOK(w, videoSummary(uint(id)))
			return
		}
		http.NotFound(w, r)
	}))
}

func videoSummary(id uint) map[string]any {
	return map[string]any{"id": id, "authorId": 9001, "title": "Integration Video", "duration": 120, "status": "approved", "transcodeStatus": "ready"}
}

func writeOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "message": "ok", "data": data})
}

func signedToken(t *testing.T, userID uint, role string) string {
	t.Helper()
	claims := middleware.Claims{UserID: userID, Role: role, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func callAPI(t *testing.T, method, endpoint, token string, body any, wantStatus int) apiEnvelope {
	t.Helper()
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequest(method, endpoint, reader)
	if err != nil {
		t.Fatal(err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("X-Request-ID", "integration-"+strconv.FormatInt(time.Now().UnixNano(), 10))
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var envelope apiEnvelope
	if err := json.NewDecoder(response.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode %s %s: %v", method, endpoint, err)
	}
	if response.StatusCode != wantStatus {
		t.Fatalf("%s %s status=%d want=%d body=%+v", method, endpoint, response.StatusCode, wantStatus, envelope)
	}
	if envelope.Code == 0 && wantStatus >= 400 {
		t.Fatalf("%s %s returned success code for error status", method, endpoint)
	}
	return envelope
}

func callData(t *testing.T, method, endpoint, token string, body any, wantStatus int) map[string]any {
	t.Helper()
	envelope := callAPI(t, method, endpoint, token, body, wantStatus)
	var data map[string]any
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode data: %v data=%s", err, envelope.Data)
	}
	return data
}

func number(t *testing.T, data map[string]any, key string) float64 {
	t.Helper()
	value, ok := data[key].(float64)
	if !ok {
		t.Fatalf("%s is not a number in %+v", key, data)
	}
	return value
}

func assertNumber(t *testing.T, data map[string]any, key string, want float64) {
	t.Helper()
	if got := number(t, data, key); got != want {
		t.Fatalf("%s=%v want=%v data=%+v", key, got, want, data)
	}
}

func assertBool(t *testing.T, data map[string]any, key string, want bool) {
	t.Helper()
	got, ok := data[key].(bool)
	if !ok || got != want {
		t.Fatalf("%s=%v want=%v data=%+v", key, got, want, data)
	}
}

func waitForViewerCount(t *testing.T, baseURL string, roomID uint, want float64) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		data := callData(t, http.MethodGet, fmt.Sprintf("%s/api/v1/live/%d/interaction", baseURL, roomID), "", nil, http.StatusOK)
		if got, ok := data["viewerCount"].(float64); ok && got == want {
			return
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("viewer count did not become %v", want)
}

func waitForWebSocketType(conn *websocket.Conn, want string, timeout time.Duration) bool {
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	for {
		var message struct {
			Type string `json:"type"`
		}
		if err := conn.ReadJSON(&message); err != nil {
			return false
		}
		if message.Type == want {
			return true
		}
	}
}
