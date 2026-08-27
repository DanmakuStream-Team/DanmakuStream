//go:build integration

package danmakulogic

import (
	"strings"
	"testing"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"danmakustream/backend/internal/testutil"
)

func TestCanSendChatModesAndSlowMode(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t,
		&model.User{}, &model.LiveRoom{}, &model.Follow{}, &model.CreatorSubscription{},
	)
	owner := model.User{Username: "chat_owner", Nickname: "Chat Owner", Password: "test", Role: "creator"}
	viewer := model.User{Username: "chat_viewer", Nickname: "Chat Viewer", Password: "test", Role: "user"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&viewer).Error; err != nil {
		t.Fatal(err)
	}
	room := model.LiveRoom{Title: "chat", OwnerID: owner.ID, StreamKey: "chat-key", Status: "live", ChatMode: "everyone"}
	if err := db.Create(&room).Error; err != nil {
		t.Fatal(err)
	}
	hub := &Hub{
		svcCtx:     &svc.ServiceContext{DB: db},
		lastChatAt: make(map[uint]map[uint]time.Time),
	}
	if ok, reason, _ := hub.canSendChat(room.ID+9999, viewer.ID); ok || !strings.Contains(reason, "不存在") {
		t.Fatalf("missing room = ok:%v reason:%q", ok, reason)
	}

	if ok, reason, _ := hub.canSendChat(room.ID, viewer.ID); !ok || reason != "" {
		t.Fatalf("everyone mode denied viewer: ok=%v reason=%q", ok, reason)
	}

	if err := db.Model(&room).Update("chat_mode", "followers").Error; err != nil {
		t.Fatal(err)
	}
	if ok, reason, _ := hub.canSendChat(room.ID, viewer.ID); ok || !strings.Contains(reason, "关注者") {
		t.Fatalf("followers mode without follow = ok:%v reason:%q", ok, reason)
	}
	if err := db.Create(&model.Follow{FollowerID: viewer.ID, FolloweeID: owner.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if ok, reason, _ := hub.canSendChat(room.ID, viewer.ID); !ok || reason != "" {
		t.Fatalf("followers mode denied follower: ok=%v reason=%q", ok, reason)
	}

	if err := db.Model(&room).Update("chat_mode", "members").Error; err != nil {
		t.Fatal(err)
	}
	if ok, reason, _ := hub.canSendChat(room.ID, viewer.ID); ok || !strings.Contains(reason, "付费订阅者") {
		t.Fatalf("members mode without subscription = ok:%v reason:%q", ok, reason)
	}
	if err := db.Create(&model.CreatorSubscription{
		SubscriberID: viewer.ID, CreatorID: owner.ID, PriceCents: 500, Status: "active",
		StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour),
	}).Error; err != nil {
		t.Fatal(err)
	}
	if ok, reason, _ := hub.canSendChat(room.ID, viewer.ID); !ok || reason != "" {
		t.Fatalf("members mode denied active subscriber: ok=%v reason=%q", ok, reason)
	}

	if err := db.Model(&room).Updates(map[string]any{"chat_mode": "everyone", "slow_mode_seconds": 30}).Error; err != nil {
		t.Fatal(err)
	}
	if ok, _, _ := hub.canSendChat(room.ID, viewer.ID); !ok {
		t.Fatal("slow mode denied first message")
	}
	if ok, reason, remaining := hub.canSendChat(room.ID, viewer.ID); ok || remaining <= 0 || !strings.Contains(reason, "慢速模式") {
		t.Fatalf("slow mode second message = ok:%v reason:%q remaining:%d", ok, reason, remaining)
	}
	if ok, reason, _ := hub.canSendChat(room.ID, owner.ID); !ok || reason != "" {
		t.Fatalf("slow mode should not restrict owner: ok=%v reason=%q", ok, reason)
	}
}
