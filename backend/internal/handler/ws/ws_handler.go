package handler

import (
	"net/http"
	"strconv"

	"danmakustream/backend/internal/logic/danmaku"
	"danmakustream/backend/internal/middleware"
	"danmakustream/backend/internal/svc"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:    func(r *http.Request) bool { return true }, // TODO: restrict in prod
}

// LiveWebSocketHandler handles real-time danmaku in a live room.
func LiveWebSocketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	hub := danmakulogic.GetHub(svcCtx)
	return func(w http.ResponseWriter, r *http.Request) {
		roomIDStr := r.PathValue("id") // Go 1.22 pattern routing
		roomID, err := strconv.ParseUint(roomIDStr, 10, 64)
		if err != nil {
			http.Error(w, "invalid room id", http.StatusBadRequest)
			return
		}

		userID, _ := r.Context().Value(middleware.CtxKeyUserID).(uint)

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &danmakulogic.Client{
			Hub:    hub,
			Conn:   conn,
			RoomID: uint(roomID),
			UserID: userID,
			Send:   make(chan []byte, 256),
		}
		hub.Register <- client
		go client.WritePump()
		go client.ReadPump()
	}
}
