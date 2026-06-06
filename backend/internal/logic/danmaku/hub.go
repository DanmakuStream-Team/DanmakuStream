package danmakulogic

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	rooms      map[uint]map[*Client]bool
	mu         sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *RoomMessage
	svcCtx     *svc.ServiceContext
}

type RoomMessage struct {
	RoomID  uint
	Payload []byte
}

var (
	globalHub *Hub
	hubOnce   sync.Once
)

func GetHub(svcCtx *svc.ServiceContext) *Hub {
	hubOnce.Do(func() {
		globalHub = &Hub{
			rooms:      make(map[uint]map[*Client]bool),
			Register:   make(chan *Client, 256),
			Unregister: make(chan *Client, 256),
			Broadcast:  make(chan *RoomMessage, 1024),
			svcCtx:     svcCtx,
		}
		go globalHub.Run()
	})
	return globalHub
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.rooms[client.RoomID] == nil {
				h.rooms[client.RoomID] = make(map[*Client]bool)
			}
			h.rooms[client.RoomID][client] = true
			h.mu.Unlock()
			h.broadcastViewerCount(client.RoomID)

		case client := <-h.Unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[client.RoomID]; ok {
				delete(clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			h.broadcastViewerCount(client.RoomID)

		case msg := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.rooms[msg.RoomID] {
				select {
				case client.Send <- msg.Payload:
				default:
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) broadcastViewerCount(roomID uint) {
	h.mu.RLock()
	count := len(h.rooms[roomID])
	h.mu.RUnlock()

	// Sync viewer count to MySQL
	go func() {
		h.svcCtx.DB.Model(&model.LiveRoom{}).Where("id = ? AND status = ?", roomID, "live").
			Update("viewer_count", count)
	}()

	payload, _ := json.Marshal(map[string]any{
		"type":    "viewer_count",
		"payload": count,
	})
	h.Broadcast <- &RoomMessage{RoomID: roomID, Payload: payload}
}

// Client represents a connected WebSocket client.
type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	RoomID uint
	UserID uint
	Send   chan []byte
}

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type IncomingMessage struct {
	Type        string `json:"type"`
	Content     string `json:"content"`
	Color       string `json:"color"`
	Time        int    `json:"time"`
	FontSize    string `json:"fontSize"`
	DanmakuType string `json:"danmakuType"`
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var incoming IncomingMessage
		if err := json.Unmarshal(message, &incoming); err != nil {
			continue
		}
		if incoming.Type != "danmaku" || incoming.Content == "" {
			continue
		}

		// Persist danmaku directly to MySQL
		fontSize := incoming.FontSize
		if fontSize == "" {
			fontSize = "medium"
		}
		danmakuType := incoming.DanmakuType
		if danmakuType == "" {
			danmakuType = "scroll"
		}
		danmaku := model.Danmaku{
			VideoID:  c.RoomID,
			UserID:   c.UserID,
			Content:  incoming.Content,
			Color:    incoming.Color,
			Time:     incoming.Time,
			FontSize: fontSize,
			Type:     danmakuType,
		}
		c.Hub.svcCtx.DB.Create(&danmaku)

		// Broadcast to room
		outgoing, _ := json.Marshal(map[string]any{
			"type": "danmaku",
			"payload": map[string]any{
				"id":          danmaku.ID,
				"userId":      c.UserID,
				"content":     incoming.Content,
				"color":       incoming.Color,
				"time":        incoming.Time,
				"fontSize":    fontSize,
				"danmakuType": danmakuType,
			},
		})
		c.Hub.Broadcast <- &RoomMessage{RoomID: c.RoomID, Payload: outgoing}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("[WS] write error:", err)
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
