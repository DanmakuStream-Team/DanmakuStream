package danmakulogic

import (
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

// Hub maintains the set of active clients and broadcasts messages.
type Hub struct {
	rooms      map[uint]map[*Client]bool
	mu         sync.RWMutex
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *RoomMessage
	svcCtx     *svc.ServiceContext
	chatMu     sync.Mutex
	lastChatAt map[uint]map[uint]time.Time
	// persistDanmaku is injectable in tests. Production uses svcCtx.DB.
	persistDanmaku func(*model.Danmaku) error
}

type RoomMessage struct {
	RoomID  uint
	Payload []byte
}

// BroadcastEvent publishes a typed live-room event to every connected viewer.
func (h *Hub) BroadcastEvent(roomID uint, eventType string, payload any) {
	data, err := json.Marshal(map[string]any{"type": eventType, "payload": payload})
	if err != nil {
		return
	}
	h.Broadcast <- &RoomMessage{RoomID: roomID, Payload: data}
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
			lastChatAt: make(map[uint]map[uint]time.Time),
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
	count := countRoomViewers(h.rooms[roomID])
	h.mu.RUnlock()

	// Sync viewer count to MySQL
	go func() {
		h.svcCtx.DB.Model(&model.LiveRoom{}).Where("id = ? AND status = ?", roomID, "live").
			Updates(map[string]any{
				"viewer_count": count,
				"viewer_peak":  gorm.Expr("GREATEST(viewer_peak, ?)", count),
			})
	}()

	payload, _ := json.Marshal(map[string]any{
		"type":    "viewer_count",
		"payload": count,
	})
	h.Broadcast <- &RoomMessage{RoomID: roomID, Payload: payload}
}

// Client represents a connected WebSocket client.
type Client struct {
	Hub     *Hub
	Conn    *websocket.Conn
	RoomID  uint
	UserID  uint
	User    *model.UserInfo
	Monitor bool
	Send    chan []byte
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
		incoming.Content = strings.TrimSpace(incoming.Content)
		if incoming.Type != "danmaku" || incoming.Content == "" || c.UserID == 0 {
			continue
		}
		if allowed, message, retryAfter := c.Hub.canSendChat(c.RoomID, c.UserID); !allowed {
			c.sendEvent("chat_error", map[string]any{"message": message, "retryAfter": retryAfter})
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
			Scene:    "live",
			UserID:   c.UserID,
			Content:  incoming.Content,
			Color:    incoming.Color,
			Time:     incoming.Time,
			FontSize: fontSize,
			Type:     danmakuType,
		}
		if err := c.Hub.saveDanmaku(&danmaku); err != nil {
			log.Println("[WS] persist live danmaku error:", err)
			c.sendEvent("chat_error", map[string]any{"message": "弹幕发送失败，请稍后重试", "retryAfter": 0})
			continue
		}

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
				"author":      c.User,
			},
		})
		c.Hub.Broadcast <- &RoomMessage{RoomID: c.RoomID, Payload: outgoing}
	}
}

// countRoomViewers counts each authenticated viewer once per room. Anonymous
// viewers have no stable identity, so each anonymous connection counts once.
// Broadcaster monitor connections never count as viewers.
func countRoomViewers(clients map[*Client]bool) int {
	uniqueUsers := make(map[uint]struct{})
	anonymousConnections := 0
	for client := range clients {
		if client.Monitor {
			continue
		}
		if client.UserID == 0 {
			anonymousConnections++
			continue
		}
		uniqueUsers[client.UserID] = struct{}{}
	}
	return len(uniqueUsers) + anonymousConnections
}

func (h *Hub) saveDanmaku(danmaku *model.Danmaku) error {
	if h.persistDanmaku != nil {
		return h.persistDanmaku(danmaku)
	}
	return h.svcCtx.DB.Create(danmaku).Error
}

func (h *Hub) canSendChat(roomID, userID uint) (bool, string, int) {
	var room model.LiveRoom
	if err := h.svcCtx.DB.Select("id", "owner_id", "chat_mode", "slow_mode_seconds").First(&room, roomID).Error; err != nil {
		return false, "直播间不存在", 0
	}
	if userID != room.OwnerID {
		switch room.ChatMode {
		case "followers":
			var count int64
			h.svcCtx.DB.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ?", userID, room.OwnerID).Count(&count)
			if count == 0 {
				return false, "当前直播间仅关注者可以发言", 0
			}
		case "members":
			var count int64
			h.svcCtx.DB.Model(&model.CreatorSubscription{}).
				Where("subscriber_id = ? AND creator_id = ? AND status = ? AND expires_at > ?", userID, room.OwnerID, "active", time.Now()).Count(&count)
			if count == 0 {
				return false, "当前直播间仅付费订阅者可以发言", 0
			}
		}
	}
	if room.SlowModeSeconds <= 0 || userID == room.OwnerID {
		return true, "", 0
	}
	h.chatMu.Lock()
	defer h.chatMu.Unlock()
	if h.lastChatAt[roomID] == nil {
		h.lastChatAt[roomID] = make(map[uint]time.Time)
	}
	remaining := room.SlowModeSeconds - int(time.Since(h.lastChatAt[roomID][userID]).Seconds())
	if !h.lastChatAt[roomID][userID].IsZero() && remaining > 0 {
		return false, "慢速模式已开启，请稍后再发送", remaining
	}
	h.lastChatAt[roomID][userID] = time.Now()
	return true, "", 0
}

func (c *Client) sendEvent(eventType string, payload any) {
	data, err := json.Marshal(map[string]any{"type": eventType, "payload": payload})
	if err != nil {
		return
	}
	select {
	case c.Send <- data:
	default:
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
