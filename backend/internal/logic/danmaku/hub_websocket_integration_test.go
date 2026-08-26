//go:build integration

package danmakulogic

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"danmakustream/backend/internal/testutil"

	"github.com/gorilla/websocket"
)

func TestHubWebSocketRoundTripAndPersistenceFailure(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t, &model.User{}, &model.LiveRoom{}, &model.Danmaku{})
	owner := model.User{Username: "ws_owner", Nickname: "WS Owner", Password: "test", Role: "creator"}
	viewer := model.User{Username: "ws_viewer", Nickname: "WS Viewer", Password: "test", Role: "user"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&viewer).Error; err != nil {
		t.Fatal(err)
	}
	room := model.LiveRoom{Title: "ws", OwnerID: owner.ID, StreamKey: "ws-key", Status: "live", ChatMode: "everyone"}
	if err := db.Create(&room).Error; err != nil {
		t.Fatal(err)
	}

	hub := &Hub{
		rooms: make(map[uint]map[*Client]bool), Register: make(chan *Client, 8),
		Unregister: make(chan *Client, 8), Broadcast: make(chan *RoomMessage, 32),
		svcCtx: &svc.ServiceContext{DB: db}, lastChatAt: make(map[uint]map[uint]time.Time),
	}
	go hub.Run()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &Client{
			Hub: hub, Conn: conn, RoomID: room.ID, UserID: viewer.ID,
			User: &model.UserInfo{ID: viewer.ID, Username: viewer.Username}, Send: make(chan []byte, 16),
		}
		hub.Register <- client
		go client.WritePump()
		client.ReadPump()
	}))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	connect := func() *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatal(err)
		}
		return conn
	}
	readType := func(conn *websocket.Conn, want string) map[string]any {
		t.Helper()
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		for {
			_, payload, err := conn.ReadMessage()
			if err != nil {
				t.Fatalf("read %s event: %v", want, err)
			}
			var event map[string]any
			if json.Unmarshal(payload, &event) == nil && event["type"] == want {
				return event
			}
		}
	}

	conn := connect()
	readType(conn, "viewer_count")
	if err := conn.WriteMessage(websocket.TextMessage, []byte("not-json")); err != nil {
		t.Fatal(err)
	}
	if err := conn.WriteJSON(IncomingMessage{Type: "noop", Content: "ignored"}); err != nil {
		t.Fatal(err)
	}
	if err := conn.WriteJSON(IncomingMessage{Type: "danmaku", Content: "  websocket message  ", Color: "#fff", Time: 7}); err != nil {
		t.Fatal(err)
	}
	event := readType(conn, "danmaku")
	payload, _ := event["payload"].(map[string]any)
	if payload["content"] != "websocket message" || payload["fontSize"] != "medium" || payload["danmakuType"] != "scroll" {
		t.Fatalf("unexpected danmaku payload: %#v", payload)
	}
	eventually(t, func() bool {
		var count int64
		return db.Model(&model.Danmaku{}).Where("video_id = ? AND scene = ?", room.ID, "live").Count(&count).Error == nil && count == 1
	})
	_ = conn.Close()
	eventually(t, func() bool {
		hub.mu.RLock()
		defer hub.mu.RUnlock()
		return len(hub.rooms[room.ID]) == 0
	})

	hub.persistDanmaku = func(*model.Danmaku) error { return errors.New("forced persistence error") }
	failedConn := connect()
	readType(failedConn, "viewer_count")
	if err := failedConn.WriteJSON(IncomingMessage{Type: "danmaku", Content: "must fail"}); err != nil {
		t.Fatal(err)
	}
	errorEvent := readType(failedConn, "chat_error")
	errorPayload, _ := errorEvent["payload"].(map[string]any)
	if !strings.Contains(errorPayload["message"].(string), "失败") {
		t.Fatalf("unexpected chat error: %#v", errorPayload)
	}
	_ = failedConn.Close()
}

func eventually(t *testing.T, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("condition was not met before timeout")
}
