package handler

import (
	"github.com/gorilla/websocket"
	"testing"
)

func TestSuperChatSeconds(t *testing.T) {
	cases := map[int64]int{10: 15, 50: 30, 200: 60, 500: 90, 1000: 120}
	for value, want := range cases {
		if got := superChatSeconds(value); got != want {
			t.Fatalf("value=%d got=%d want=%d", value, got, want)
		}
	}
}

func TestCountedConnectionsDeduplicatesAuthenticatedViewer(t *testing.T) {
	a, b, c := &websocket.Conn{}, &websocket.Conn{}, &websocket.Conn{}
	viewers := map[*websocket.Conn]viewer{a: {UserID: 7, Counted: true}, b: {UserID: 7, Counted: true}, c: {UserID: 0, Counted: true}}
	if got := countedConnections(viewers); got != 2 {
		t.Fatalf("got=%d want=2", got)
	}
}
