package handler

import (
	"github.com/gorilla/websocket"
	"sync"
)

type Hub struct {
	mu      sync.RWMutex
	rooms   map[uint]map[*websocket.Conn]viewer
	writers map[*websocket.Conn]*sync.Mutex
}
type viewer struct {
	UserID  uint
	Counted bool
}

func NewHub() *Hub {
	return &Hub{rooms: map[uint]map[*websocket.Conn]viewer{}, writers: map[*websocket.Conn]*sync.Mutex{}}
}
func (h *Hub) Add(roomID uint, conn *websocket.Conn, userID uint, counted bool) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[roomID] == nil {
		h.rooms[roomID] = map[*websocket.Conn]viewer{}
	}
	h.rooms[roomID][conn] = viewer{UserID: userID, Counted: counted}
	h.writers[conn] = &sync.Mutex{}
	return countedConnections(h.rooms[roomID])
}
func (h *Hub) Remove(roomID uint, conn *websocket.Conn) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms[roomID], conn)
	delete(h.writers, conn)
	if len(h.rooms[roomID]) == 0 {
		delete(h.rooms, roomID)
		return 0
	}
	return countedConnections(h.rooms[roomID])
}
func countedConnections(conns map[*websocket.Conn]viewer) int {
	count := 0
	users := map[uint]struct{}{}
	for _, v := range conns {
		if !v.Counted {
			continue
		}
		if v.UserID == 0 {
			count++
			continue
		}
		users[v.UserID] = struct{}{}
	}
	return count + len(users)
}
func (h *Hub) Broadcast(roomID uint, payload any) {
	h.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.rooms[roomID]))
	for c := range h.rooms[roomID] {
		conns = append(conns, c)
	}
	h.mu.RUnlock()
	for _, c := range conns {
		h.Write(c, payload)
	}
}
func (h *Hub) Write(conn *websocket.Conn, payload any) {
	h.mu.RLock()
	mu := h.writers[conn]
	h.mu.RUnlock()
	if mu == nil {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	_ = conn.WriteJSON(payload)
}
