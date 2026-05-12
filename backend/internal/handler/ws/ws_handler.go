package handler

import (
	"net/http"
	"strconv"

	danmakulogic "danmakustream/backend/internal/logic/danmaku"
	"danmakustream/backend/internal/middleware"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func LiveWebSocketHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	hub := danmakulogic.GetHub(svcCtx)
	return func(c *gin.Context) {
		roomID, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
			return
		}

		userID, _ := c.Get(middleware.CtxKeyUserID)

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &danmakulogic.Client{
			Hub:    hub,
			Conn:   conn,
			RoomID: uint(roomID),
			UserID: userID.(uint),
			Send:   make(chan []byte, 256),
		}
		hub.Register <- client
		go client.WritePump()
		go client.ReadPump()
	}
}
