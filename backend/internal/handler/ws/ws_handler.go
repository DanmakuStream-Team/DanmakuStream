package handler

import (
	"net/http"
	"strconv"
	"strings"

	chatlogic "danmakustream/backend/internal/logic/chat"
	danmakulogic "danmakustream/backend/internal/logic/danmaku"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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

		userID, _ := getUserIDFromLiveRequest(c, svcCtx)
		monitor := c.Query("monitor") == "1"
		if monitor {
			var count int64
			if userID == 0 || svcCtx.DB.Model(&model.LiveRoom{}).Where("id = ? AND owner_id = ?", roomID, userID).Count(&count).Error != nil || count == 0 {
				c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
		}
		var userInfo *model.UserInfo
		if userID > 0 {
			var user model.User
			if err := svcCtx.DB.First(&user, userID).Error; err == nil {
				userInfo = &model.UserInfo{ID: user.ID, Username: user.Username, Nickname: user.Nickname, Avatar: user.Avatar, Role: user.Role}
			}
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		client := &danmakulogic.Client{
			Hub:     hub,
			Conn:    conn,
			RoomID:  uint(roomID),
			UserID:  userID,
			User:    userInfo,
			Monitor: monitor,
			Send:    make(chan []byte, 256),
		}
		hub.Register <- client
		go client.WritePump()
		go client.ReadPump()
	}
}

func ChatWebSocketHandler(svcCtx *svc.ServiceContext) gin.HandlerFunc {
	hub := chatlogic.GetHub(svcCtx)
	return func(c *gin.Context) {
		userID, ok := getUserIDFromLiveRequest(c, svcCtx)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		client := &chatlogic.Client{
			Hub:    hub,
			Conn:   conn,
			UserID: userID,
			Send:   make(chan []byte, 128),
		}
		hub.Register(client)
		go client.WritePump()
		go client.ReadPump()
	}
}

func getUserIDFromLiveRequest(c *gin.Context, svcCtx *svc.ServiceContext) (uint, bool) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}
	if tokenStr == "" {
		return 0, false
	}

	claims := &middleware.Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(svcCtx.Config.Auth.AccessSecret), nil
	})
	if err != nil || !token.Valid || claims.UserID == 0 {
		return 0, false
	}
	return claims.UserID, true
}
