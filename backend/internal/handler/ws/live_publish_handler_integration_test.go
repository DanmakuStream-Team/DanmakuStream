//go:build integration

package handler

import (
	"testing"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"danmakustream/backend/internal/testutil"
)

func TestFinalizeAbandonedBrowserPublisher(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t, &model.User{}, &model.LiveRoom{})
	owner := model.User{Username: "publisher-owner", Nickname: "Publisher Owner", Password: "hash", Role: "creator"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	svcCtx := &svc.ServiceContext{DB: db}

	t.Run("disconnected publisher ends live room", func(t *testing.T) {
		room := model.LiveRoom{Title: "abandoned", OwnerID: owner.ID, StreamKey: "abandoned-key", Status: "live", ViewerCount: 3}
		if err := db.Create(&room).Error; err != nil {
			t.Fatal(err)
		}
		activeBrowserPublishers.Lock()
		delete(activeBrowserPublishers.rooms, room.ID)
		activeBrowserPublishers.Unlock()
		endedAt := time.Now()
		if err := finalizeAbandonedBrowserPublisher(svcCtx, room.ID, endedAt); err != nil {
			t.Fatal(err)
		}
		if err := db.First(&room, room.ID).Error; err != nil {
			t.Fatal(err)
		}
		if room.Status != "ended" || room.ViewerCount != 0 || room.EndedAt == nil {
			t.Fatalf("room after abnormal disconnect = status:%q viewers:%d endedAt:%v", room.Status, room.ViewerCount, room.EndedAt)
		}
	})

	t.Run("publisher reconnect during grace period keeps room live", func(t *testing.T) {
		secondOwner := model.User{Username: "publisher-owner-2", Nickname: "Publisher Owner 2", Password: "hash", Role: "creator"}
		if err := db.Create(&secondOwner).Error; err != nil {
			t.Fatal(err)
		}
		room := model.LiveRoom{Title: "recovered", OwnerID: secondOwner.ID, StreamKey: "recovered-key", Status: "live", ViewerCount: 1}
		if err := db.Create(&room).Error; err != nil {
			t.Fatal(err)
		}
		activeBrowserPublishers.Lock()
		activeBrowserPublishers.rooms[room.ID] = true
		activeBrowserPublishers.Unlock()
		t.Cleanup(func() {
			activeBrowserPublishers.Lock()
			delete(activeBrowserPublishers.rooms, room.ID)
			activeBrowserPublishers.Unlock()
		})
		if err := finalizeAbandonedBrowserPublisher(svcCtx, room.ID, time.Now()); err != nil {
			t.Fatal(err)
		}
		if err := db.First(&room, room.ID).Error; err != nil {
			t.Fatal(err)
		}
		if room.Status != "live" || room.EndedAt != nil {
			t.Fatalf("recovered room = status:%q endedAt:%v", room.Status, room.EndedAt)
		}
	})
}
