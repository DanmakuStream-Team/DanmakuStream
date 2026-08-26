package danmakulogic

import (
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

func TestSaveDanmakuPropagatesPersistenceFailure(t *testing.T) {
	wantErr := errors.New("database unavailable")
	hub := &Hub{persistDanmaku: func(*model.Danmaku) error { return wantErr }}

	if err := hub.saveDanmaku(&model.Danmaku{}); !errors.Is(err, wantErr) {
		t.Fatalf("saveDanmaku() error = %v, want %v", err, wantErr)
	}
}
