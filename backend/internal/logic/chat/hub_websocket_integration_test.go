//go:build integration

package chat

import (
	"encoding/json"
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

func TestChatHubWebSocketRoundTripPersistenceAndError(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t, &model.User{}, &model.Video{}, &model.UserBlock{}, &model.ChatMessage{})
	sender := model.User{Username: "chat_sender", Nickname: "Chat Sender", Password: "test", Role: "user"}
	receiver := model.User{Username: "chat_receiver", Nickname: "Chat Receiver", Password: "test", Role: "user"}
	if err := db.Create(&sender).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&receiver).Error; err != nil {
		t.Fatal(err)
	}

	hub := &Hub{
		svc:        &svc.ServiceContext{DB: db},
		clients:    make(map[uint]map[*Client]struct{}),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan broadcastRequest, 32),
	}
	go hub.run()

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	disconnected := make(chan struct{}, 2)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := sender.ID
		if r.URL.Query().Get("user") == "receiver" {
			userID = receiver.ID
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &Client{Hub: hub, Conn: conn, UserID: userID, Send: make(chan []byte, 16)}
		hub.Register(client)
		go client.WritePump()
		client.ReadPump()
		disconnected <- struct{}{}
	}))
	defer server.Close()
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	connect := func(suffix string) *websocket.Conn {
		t.Helper()
		conn, _, err := websocket.DefaultDialer.Dial(wsURL+suffix, nil)
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

	senderConn := connect("")
	receiverConn := connect("?user=receiver")
	if err := senderConn.WriteJSON(map[string]any{
		"type": "message", "receiverId": receiver.ID, "content": " websocket hello ",
	}); err != nil {
		t.Fatal(err)
	}
	receiverEvent := readType(receiverConn, "message")
	payload, _ := receiverEvent["payload"].(map[string]any)
	if payload["content"] != "websocket hello" || uint(payload["senderId"].(float64)) != sender.ID {
		t.Fatalf("unexpected receiver event: %#v", receiverEvent)
	}
	readType(senderConn, "message")
	eventuallyChat(t, func() bool {
		var count int64
		return db.Model(&model.ChatMessage{}).
			Where("sender_id = ? AND receiver_id = ? AND content = ?", sender.ID, receiver.ID, "websocket hello").
			Count(&count).Error == nil && count == 1
	})

	if err := senderConn.WriteJSON(map[string]any{
		"type":    "message",
		"message": map[string]any{"receiverId": receiver.ID, "type": "text", "content": " "},
	}); err != nil {
		t.Fatal(err)
	}
	errorEvent := readType(senderConn, "error")
	if !strings.Contains(errorEvent["payload"].(string), "不能为空") {
		t.Fatalf("unexpected error event: %#v", errorEvent)
	}

	_ = senderConn.Close()
	_ = receiverConn.Close()
	for range 2 {
		select {
		case <-disconnected:
		case <-time.After(3 * time.Second):
			t.Fatal("websocket client did not disconnect")
		}
	}
}

func eventuallyChat(t *testing.T, condition func() bool) {
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
