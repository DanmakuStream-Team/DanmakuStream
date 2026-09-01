//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"danmakustream/user-service/internal/client"
	chatlogic "danmakustream/user-service/internal/logic/chat"
	model "danmakustream/user-service/internal/model/mysql"
	"danmakustream/user-service/internal/svc"
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

	content := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Token") != "integration-token" || r.Header.Get("X-Request-ID") != "integration-request" {
			t.Error("internal headers not propagated")
		}
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
}
