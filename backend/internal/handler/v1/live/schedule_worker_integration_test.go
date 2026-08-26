//go:build integration

package live

import (
	"testing"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"danmakustream/backend/internal/testutil"
)

func TestProcessDueLiveSchedulesIsIdempotent(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t,
		&model.User{}, &model.LiveSchedule{}, &model.LiveReservation{},
		&model.LiveRoom{}, &model.Notification{}, &model.CreatorDailyStat{},
	)
	svcCtx := &svc.ServiceContext{DB: db}
	creator := model.User{Username: "worker_creator", Nickname: "Worker Creator", Password: "test", Role: "creator"}
	viewer := model.User{Username: "worker_viewer", Nickname: "Worker Viewer", Password: "test", Role: "user"}
	if err := db.Create(&creator).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&viewer).Error; err != nil {
		t.Fatal(err)
	}

	due := model.LiveSchedule{
		Title: "due", ScheduledAt: time.Now().Add(-time.Minute), Status: "pending",
		OwnerID: creator.ID, ReminderCount: 1,
	}
	canceled := model.LiveSchedule{
		Title: "canceled", ScheduledAt: time.Now().Add(-time.Minute), Status: "canceled",
		OwnerID: creator.ID,
	}
	if err := db.Create(&due).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&canceled).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.LiveReservation{ScheduleID: due.ID, UserID: viewer.ID}).Error; err != nil {
		t.Fatal(err)
	}

	processDueLiveSchedules(svcCtx)
	processDueLiveSchedules(svcCtx)

	if err := db.First(&due, due.ID).Error; err != nil {
		t.Fatal(err)
	}
	if due.Status != "live" {
		t.Fatalf("due schedule status = %q, want live", due.Status)
	}
	if err := db.First(&canceled, canceled.ID).Error; err != nil {
		t.Fatal(err)
	}
	if canceled.Status != "canceled" {
		t.Fatalf("canceled schedule status = %q, want canceled", canceled.Status)
	}

	assertCount := func(modelValue any, query string, want int64, args ...any) {
		t.Helper()
		var count int64
		if err := db.Model(modelValue).Where(query, args...).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != want {
			t.Fatalf("count for %T = %d, want %d", modelValue, count, want)
		}
	}
	assertCount(&model.LiveRoom{}, "owner_id = ? AND status = ?", 1, creator.ID, "live")
	assertCount(&model.Notification{}, "user_id = ? AND type = ?", 1, viewer.ID, "live_start")
	assertCount(&model.CreatorDailyStat{}, "creator_id = ?", 1, creator.ID)

	// A later due schedule for the same creator must reuse and update the room.
	secondDue := model.LiveSchedule{
		Title: "second due", ScheduledAt: time.Now().Add(-time.Second), Status: "pending", OwnerID: creator.ID,
	}
	if err := db.Create(&secondDue).Error; err != nil {
		t.Fatal(err)
	}
	processDueLiveSchedules(svcCtx)
	assertCount(&model.LiveRoom{}, "owner_id = ?", 1, creator.ID)
	var reused model.LiveRoom
	if err := db.Where("owner_id = ?", creator.ID).First(&reused).Error; err != nil {
		t.Fatal(err)
	}
	if reused.Title != secondDue.Title || reused.Status != "live" {
		t.Fatalf("reused room = title:%q status:%q", reused.Title, reused.Status)
	}

	// The worker must return safely when its source table is unavailable.
	if err := db.Migrator().DropTable(&model.LiveSchedule{}); err != nil {
		t.Fatal(err)
	}
	processDueLiveSchedules(svcCtx)
}

func TestScheduleNotificationEmptyAndSelfBranches(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t,
		&model.User{}, &model.LiveSchedule{}, &model.LiveReservation{}, &model.Notification{}, &model.Follow{},
	)
	owner := model.User{Username: "notice_owner", Nickname: "Notice Owner", Password: "test", Role: "creator"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	schedule := model.LiveSchedule{Title: "no viewers", ScheduledAt: time.Now().Add(time.Hour), Status: "pending", OwnerID: owner.ID}
	if err := db.Create(&schedule).Error; err != nil {
		t.Fatal(err)
	}
	if err := notifyScheduleReservationUsers(db, schedule.ID, owner.ID, schedule.Title, "/live"); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.LiveReservation{ScheduleID: schedule.ID, UserID: owner.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := notifyScheduleReservationUsers(db, schedule.ID, owner.ID, schedule.Title, "/live"); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Follow{FollowerID: owner.ID, FolloweeID: owner.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := notifyLiveFollowers(db, owner.ID, "live_start", "self", "ignored", "/live"); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&model.Notification{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("self/empty notification count = %d, want 0", count)
	}

	second := model.LiveSchedule{Title: "pending without reservations", ScheduledAt: time.Now().Add(2 * time.Hour), Status: "pending", OwnerID: owner.ID}
	if err := db.Create(&second).Error; err != nil {
		t.Fatal(err)
	}
	if err := notifyScheduleReservations(db, owner.ID, "manual live"); err != nil {
		t.Fatal(err)
	}
}
