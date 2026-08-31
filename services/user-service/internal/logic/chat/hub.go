package chat

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"

	model "danmakustream/user-service/internal/model/mysql"
	"danmakustream/user-service/internal/svc"

	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var (
	ErrEmptyContent           = errors.New("message content is empty")
	ErrTooLong                = errors.New("message content is too long")
	ErrSelfMessage            = errors.New("cannot message yourself")
	ErrUserNotFound           = errors.New("receiver not found")
	ErrBlocked                = errors.New("blocked relationship")
	ErrInvalidType            = errors.New("invalid message type")
	ErrInvalidMedia           = errors.New("invalid message media")
	ErrVideoMissing           = errors.New("shared video not found")
	ErrInvalidClientMessageID = errors.New("invalid client message id")
)

const (
	MessageTypeText       = "text"
	MessageTypeImage      = "image"
	MessageTypeVideo      = "video"
	MessageTypeVideoShare = "video_share"
)

type CreateMessageInput struct {
	ReceiverID      uint   `json:"receiverId"`
	ClientMessageID string `json:"clientMessageId"`
	Type            string `json:"type"`
	Content         string `json:"content"`
	MediaURL        string `json:"mediaUrl"`
	MediaName       string `json:"mediaName"`
	VideoID         uint   `json:"videoId"`
}

type SharedVideoInfo struct {
	ID       uint           `json:"id"`
	Title    string         `json:"title"`
	CoverURL string         `json:"coverUrl"`
	Duration int            `json:"duration"`
	Author   model.UserInfo `json:"author"`
}

type MessageInfo struct {
	ID              uint             `json:"id"`
	SenderID        uint             `json:"senderId"`
	ReceiverID      uint             `json:"receiverId"`
	ClientMessageID string           `json:"clientMessageId,omitempty"`
	Content         string           `json:"content"`
	Type            string           `json:"type"`
	MediaURL        string           `json:"mediaUrl"`
	MediaName       string           `json:"mediaName"`
	Video           *SharedVideoInfo `json:"video,omitempty"`
	Read            bool             `json:"read"`
	Sender          model.UserInfo   `json:"sender"`
	Receiver        model.UserInfo   `json:"receiver"`
	CreatedAt       string           `json:"createdAt"`
}

type envelope struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type incomingMessage struct {
	Type       string             `json:"type"`
	Message    CreateMessageInput `json:"message"`
	ReceiverID uint               `json:"receiverId"`
	Content    string             `json:"content"`
}

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	UserID uint
	Send   chan []byte
}

type Hub struct {
	svc        *svc.ServiceContext
	clients    map[uint]map[*Client]struct{}
	register   chan *Client
	unregister chan *Client
	broadcast  chan broadcastRequest
}

type broadcastRequest struct {
	userIDs []uint
	data    []byte
}

var (
	hubOnce sync.Once
	hubInst *Hub
)

func GetHub(svcCtx *svc.ServiceContext) *Hub {
	hubOnce.Do(func() {
		hubInst = &Hub{
			svc:        svcCtx,
			clients:    make(map[uint]map[*Client]struct{}),
			register:   make(chan *Client),
			unregister: make(chan *Client),
			broadcast:  make(chan broadcastRequest, 256),
		}
		go hubInst.run()
	})
	return hubInst
}

func (h *Hub) Register(client *Client) {
	h.register <- client
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			if h.clients[client.UserID] == nil {
				h.clients[client.UserID] = make(map[*Client]struct{})
			}
			h.clients[client.UserID][client] = struct{}{}
		case client := <-h.unregister:
			h.removeClient(client)
		case request := <-h.broadcast:
			seen := make(map[uint]struct{}, len(request.userIDs))
			for _, userID := range request.userIDs {
				if _, exists := seen[userID]; exists {
					continue
				}
				seen[userID] = struct{}{}
				for client := range h.clients[userID] {
					select {
					case client.Send <- request.data:
					default:
						h.removeClient(client)
					}
				}
			}
		}
	}
}

func (h *Hub) removeClient(client *Client) {
	connections := h.clients[client.UserID]
	if _, ok := connections[client]; !ok {
		return
	}
	delete(connections, client)
	close(client.Send)
	if len(connections) == 0 {
		delete(h.clients, client.UserID)
	}
}

func (h *Hub) CreateAndBroadcast(senderID uint, input CreateMessageInput) (MessageInfo, error) {
	input.Type = strings.TrimSpace(input.Type)
	if input.Type == "" {
		input.Type = MessageTypeText
	}
	input.Content = strings.TrimSpace(input.Content)
	input.MediaURL = strings.TrimSpace(input.MediaURL)
	input.MediaName = strings.TrimSpace(input.MediaName)
	input.ClientMessageID = strings.TrimSpace(input.ClientMessageID)
	if len(input.ClientMessageID) > 64 {
		return MessageInfo{}, ErrInvalidClientMessageID
	}
	if senderID == input.ReceiverID {
		return MessageInfo{}, ErrSelfMessage
	}
	if input.ReceiverID == 0 {
		return MessageInfo{}, ErrUserNotFound
	}
	if err := validateMessageInput(senderID, &input); err != nil {
		return MessageInfo{}, err
	}

	var receiverCount int64
	if err := h.svc.DB.Model(&model.User{}).Where("id = ?", input.ReceiverID).Count(&receiverCount).Error; err != nil {
		return MessageInfo{}, err
	}
	if receiverCount == 0 {
		return MessageInfo{}, ErrUserNotFound
	}

	var blockCount int64
	if err := h.svc.DB.Model(&model.UserBlock{}).
		Where("(blocker_id = ? AND blocked_id = ?) OR (blocker_id = ? AND blocked_id = ?)", senderID, input.ReceiverID, input.ReceiverID, senderID).
		Count(&blockCount).Error; err != nil {
		return MessageInfo{}, err
	}
	if blockCount > 0 {
		return MessageInfo{}, ErrBlocked
	}
	if input.ClientMessageID != "" {
		var existing model.ChatMessage
		err := h.svc.DB.Preload("Sender").Preload("Receiver").
			Where("sender_id = ? AND client_message_id = ?", senderID, input.ClientMessageID).First(&existing).Error
		if err == nil {
			return ToMessageInfo(existing), nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return MessageInfo{}, err
		}
	}

	message := model.ChatMessage{
		SenderID: senderID, ReceiverID: input.ReceiverID, MessageType: input.Type,
		Content: input.Content, MediaURL: input.MediaURL, MediaName: input.MediaName,
	}
	if input.ClientMessageID != "" {
		message.ClientMessageID = &input.ClientMessageID
	}
	if input.Type == MessageTypeVideoShare {
		// content-service owns video availability. Day 6 stores only the external
		// ID; validation moves to its authenticated internal API on Day 7.
		message.SharedVideoID = &input.VideoID
	}
	if err := h.svc.DB.Create(&message).Error; err != nil {
		// Two HTTP/WebSocket retries may race after both miss the lookup above.
		// The composite unique index is the final arbiter; return its winning row
		// so callers still observe a successful idempotent operation.
		if input.ClientMessageID != "" {
			var existing model.ChatMessage
			lookupErr := h.svc.DB.Preload("Sender").Preload("Receiver").
				Where("sender_id = ? AND client_message_id = ?", senderID, input.ClientMessageID).First(&existing).Error
			if lookupErr == nil {
				return ToMessageInfo(existing), nil
			}
		}
		return MessageInfo{}, err
	}
	if err := h.svc.DB.Preload("Sender").Preload("Receiver").First(&message, message.ID).Error; err != nil {
		return MessageInfo{}, err
	}
	info := ToMessageInfo(message)
	h.publish([]uint{senderID, input.ReceiverID}, envelope{Type: "message", Payload: info})
	return info, nil
}

func (h *Hub) publish(userIDs []uint, payload envelope) {
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	select {
	case h.broadcast <- broadcastRequest{userIDs: userIDs, data: data}:
	default:
	}
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()
	c.Conn.SetReadLimit(16 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(70 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(70 * time.Second))
		return nil
	})
	for {
		var incoming incomingMessage
		if err := c.Conn.ReadJSON(&incoming); err != nil {
			return
		}
		if incoming.Type != "message" {
			continue
		}
		if incoming.Message.ReceiverID == 0 && incoming.ReceiverID != 0 {
			incoming.Message = CreateMessageInput{ReceiverID: incoming.ReceiverID, Type: MessageTypeText, Content: incoming.Content}
		}
		if _, err := c.Hub.CreateAndBroadcast(c.UserID, incoming.Message); err != nil {
			data, _ := json.Marshal(envelope{Type: "error", Payload: chatErrorMessage(err)})
			select {
			case c.Send <- data:
			default:
				return
			}
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func ToMessageInfo(message model.ChatMessage) MessageInfo {
	info := MessageInfo{
		ID:         message.ID,
		SenderID:   message.SenderID,
		ReceiverID: message.ReceiverID,
		Content:    message.Content,
		Type:       normalizedMessageType(message.MessageType),
		MediaURL:   message.MediaURL,
		MediaName:  message.MediaName,
		Read:       message.Read,
		Sender: model.UserInfo{
			ID:       message.Sender.ID,
			Username: message.Sender.Username,
			Nickname: message.Sender.Nickname,
			Avatar:   message.Sender.Avatar,
			Role:     message.Sender.Role,
		},
		Receiver: model.UserInfo{
			ID:       message.Receiver.ID,
			Username: message.Receiver.Username,
			Nickname: message.Receiver.Nickname,
			Avatar:   message.Receiver.Avatar,
			Role:     message.Receiver.Role,
		},
		CreatedAt: message.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if message.ClientMessageID != nil {
		info.ClientMessageID = *message.ClientMessageID
	}
	if message.SharedVideoID != nil {
		info.Video = &SharedVideoInfo{
			ID: *message.SharedVideoID,
		}
	}
	return info
}

func validateMessageInput(senderID uint, input *CreateMessageInput) error {
	if len([]rune(input.Content)) > 2000 || len([]rune(input.MediaName)) > 200 {
		return ErrTooLong
	}
	switch input.Type {
	case MessageTypeText:
		if input.Content == "" {
			return ErrEmptyContent
		}
		input.MediaURL, input.MediaName, input.VideoID = "", "", 0
	case MessageTypeImage, MessageTypeVideo:
		prefix := "/media/messages/" + strconv.FormatUint(uint64(senderID), 10) + "/"
		if !strings.HasPrefix(input.MediaURL, prefix) || strings.Contains(input.MediaURL, "..") || strings.ContainsAny(input.MediaURL, "?#") {
			return ErrInvalidMedia
		}
		lowerURL := strings.ToLower(input.MediaURL)
		validExtension := input.Type == MessageTypeImage && hasAnySuffix(lowerURL, ".jpg", ".jpeg", ".png", ".webp", ".gif")
		validExtension = validExtension || input.Type == MessageTypeVideo && hasAnySuffix(lowerURL, ".mp4", ".webm")
		if !validExtension {
			return ErrInvalidMedia
		}
		if input.Content == "" {
			if input.Type == MessageTypeImage {
				input.Content = "[图片]"
			} else {
				input.Content = "[视频]"
			}
		}
		input.VideoID = 0
	case MessageTypeVideoShare:
		if input.VideoID == 0 {
			return ErrVideoMissing
		}
		if input.Content == "" {
			input.Content = "[视频分享]"
		}
		input.MediaURL, input.MediaName = "", ""
	default:
		return ErrInvalidType
	}
	return nil
}

func normalizedMessageType(value string) string {
	if value == "" {
		return MessageTypeText
	}
	return value
}

func hasAnySuffix(value string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(value, suffix) {
			return true
		}
	}
	return false
}

func chatErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrEmptyContent):
		return "消息内容不能为空"
	case errors.Is(err, ErrTooLong):
		return "消息不能超过 2000 个字符"
	case errors.Is(err, ErrSelfMessage):
		return "不能给自己发送私信"
	case errors.Is(err, ErrUserNotFound):
		return "接收用户不存在"
	case errors.Is(err, ErrBlocked):
		return "存在拉黑关系，无法发送私信"
	case errors.Is(err, ErrInvalidType):
		return "不支持的消息类型"
	case errors.Is(err, ErrInvalidMedia):
		return "私信附件无效，请重新上传"
	case errors.Is(err, ErrVideoMissing):
		return "分享的视频不存在或尚未公开"
	case errors.Is(err, ErrInvalidClientMessageID):
		return "客户端消息编号无效"
	default:
		return "消息发送失败"
	}
}

func ErrorMessage(err error) string {
	return chatErrorMessage(err)
}
