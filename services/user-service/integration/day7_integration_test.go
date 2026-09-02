//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"danmakustream/user-service/internal/client"
	wshandler "danmakustream/user-service/internal/handler/ws"
	chatlogic "danmakustream/user-service/internal/logic/chat"
	"danmakustream/user-service/internal/middleware"
	model "danmakustream/user-service/internal/model/mysql"
	"danmakustream/user-service/internal/svc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestVideoShareUsesContentAPIAndUserDB(t *testing.T) {
	dsn := os.Getenv("USER_SERVICE_TEST_DSN")
	if dsn == "" {
		t.Skip("USER_SERVICE_TEST_DSN is not configured")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.UserBlock{}, &model.ChatMessage{}); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin()
	defer tx.Rollback()

	suffix := time.Now().UnixNano()
	sender := model.User{Username: fmt.Sprintf("sender-%d", suffix), Nickname: fmt.Sprintf("发送者-%d", suffix), Password: "hash"}
	receiver := model.User{Username: fmt.Sprintf("receiver-%d", suffix), Nickname: fmt.Sprintf("接收者-%d", suffix), Password: "hash"}
	if err := tx.Create(&sender).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&receiver).Error; err != nil {
		t.Fatal(err)
	}

	requestIDs := make(chan string, 2)
	content := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "integration-token" || r.Header.Get("X-Request-ID") == "" {
			t.Error("internal headers not propagated")
		}
		requestIDs <- r.Header.Get("X-Request-ID")
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"code":0,"message":"ok","data":{"id":77,"creatorId":%d,"title":"联调视频","coverUrl":"/media/cover.jpg","duration":88,"status":"published","playable":true}}`, receiver.ID)
	}))
	defer content.Close()

	ctx := &svc.ServiceContext{DB: tx, Content: client.NewContent(content.URL, "integration-token", time.Second)}
	message, err := chatlogic.GetHub(ctx).CreateAndBroadcast(context.Background(), "integration-request", sender.ID, chatlogic.CreateMessageInput{ReceiverID: receiver.ID, Type: chatlogic.MessageTypeVideoShare, VideoID: 77, ClientMessageID: "day7-integration"})
	if err != nil {
		t.Fatal(err)
	}
	if message.Video == nil || message.Video.Title != "联调视频" || message.Video.Author.ID != receiver.ID {
		t.Fatalf("message=%+v", message)
	}
	var stored model.ChatMessage
	if err := tx.First(&stored, message.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.SharedVideoID == nil || *stored.SharedVideoID != 77 {
		t.Fatalf("stored=%+v", stored)
	}
	if got := <-requestIDs; got != "integration-request" {
		t.Fatalf("HTTP request ID = %q", got)
	}

	ctx.Config.Auth.AccessSecret = "websocket-test-secret"
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.RequestContext("user-service"))
	router.GET("/ws/chat", wshandler.Chat(ctx))
	server := httptest.NewServer(router)
	defer server.Close()
	claims := middleware.Claims{UserID: sender.ID, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(ctx.Config.Auth.AccessSecret))
	if err != nil {
		t.Fatal(err)
	}
	header := http.Header{"X-Request-ID": []string{"ws-integration-request"}}
	conn, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http")+"/ws/chat?token="+token, header)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	if err := conn.WriteJSON(map[string]any{"type": "message", "message": map[string]any{"receiverId": receiver.ID, "type": "video_share", "videoId": 77, "clientMessageId": "day7-websocket-integration"}}); err != nil {
		t.Fatal(err)
	}
	select {
	case got := <-requestIDs:
		if got != "ws-integration-request" {
			t.Fatalf("WebSocket request ID = %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("WebSocket video share did not call content-service")
	}
	if err := conn.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
		t.Fatal(err)
	}
	var outbound struct {
		Type string `json:"type"`
	}
	if err := conn.ReadJSON(&outbound); err != nil {
		t.Fatal(err)
	}
	if outbound.Type != "message" {
		t.Fatalf("WebSocket response type = %q, want message", outbound.Type)
	}
}
