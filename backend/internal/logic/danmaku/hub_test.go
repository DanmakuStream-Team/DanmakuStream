package danmakulogic

import (
	"encoding/json"
	"errors"
	"testing"

	model "danmakustream/backend/internal/model/mysql"
)

func TestCountRoomViewersDeduplicatesAuthenticatedUsers(t *testing.T) {
	clients := map[*Client]bool{
		{UserID: 7}:                true,
		{UserID: 7}:                true,
		{UserID: 8}:                true,
		{UserID: 8, Monitor: true}: true,
		{UserID: 0}:                true,
		{UserID: 0}:                true,
	}

	// Two authenticated users plus two anonymous connections. The monitor
	// connection and the second connection for user 7 are not counted.
	if got, want := countRoomViewers(clients), 4; got != want {
		t.Fatalf("countRoomViewers() = %d, want %d", got, want)
	}
}

func TestBroadcastEventAndSendEvent(t *testing.T) {
	hub := &Hub{Broadcast: make(chan *RoomMessage, 1)}
	hub.BroadcastEvent(9, "live_like", map[string]any{"liked": true})
	message := <-hub.Broadcast
	if message.RoomID != 9 {
		t.Fatalf("broadcast room = %d, want 9", message.RoomID)
	}
	var event map[string]any
	if err := json.Unmarshal(message.Payload, &event); err != nil || event["type"] != "live_like" {
		t.Fatalf("broadcast payload = %s, err=%v", message.Payload, err)
	}

	// Unsupported JSON values must be discarded without blocking the caller.
	hub.BroadcastEvent(9, "invalid", make(chan int))
	select {
	case extra := <-hub.Broadcast:
		t.Fatalf("unexpected event after marshal failure: %s", extra.Payload)
	default:
	}

	client := &Client{Send: make(chan []byte, 1)}
	client.sendEvent("chat_error", map[string]any{"retryAfter": 3})
	if err := json.Unmarshal(<-client.Send, &event); err != nil || event["type"] != "chat_error" {
		t.Fatalf("client event decode failed: event=%v err=%v", event, err)
	}
	client.sendEvent("invalid", make(chan int))
	client.Send <- []byte("occupied")
	client.sendEvent("dropped", nil)
}

func TestSaveDanmakuPropagatesPersistenceFailure(t *testing.T) {
	wantErr := errors.New("database unavailable")
	hub := &Hub{persistDanmaku: func(*model.Danmaku) error { return wantErr }}

	if err := hub.saveDanmaku(&model.Danmaku{}); !errors.Is(err, wantErr) {
		t.Fatalf("saveDanmaku() error = %v, want %v", err, wantErr)
	}
}
