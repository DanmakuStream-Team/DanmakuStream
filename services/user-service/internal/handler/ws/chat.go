package ws

import (
	chatlogic "danmakustream/user-service/internal/logic/chat"
	"danmakustream/user-service/internal/middleware"
	"danmakustream/user-service/internal/svc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
)

var upgrader = websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024, CheckOrigin: func(*http.Request) bool { return true }}

func Chat(ctx *svc.ServiceContext) gin.HandlerFunc {
	hub := chatlogic.GetHub(ctx)
	return func(c *gin.Context) {
		raw := c.Query("token")
		if raw == "" && strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
			raw = strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
		}
		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(raw, claims, func(*jwt.Token) (interface{}, error) { return []byte(ctx.Config.Auth.AccessSecret), nil })
		if err != nil || !token.Valid || claims.UserID == 0 {
			c.JSON(401, gin.H{"code": 40100, "message": "unauthorized", "requestId": c.GetString(middleware.CtxKeyRequestID)})
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		client := &chatlogic.Client{Hub: hub, Conn: conn, UserID: claims.UserID, RequestID: c.GetString(middleware.CtxKeyRequestID), Send: make(chan []byte, 128)}
		hub.Register(client)
		go client.WritePump()
		go client.ReadPump()
	}
}
