//go:build integration

package membership

import (
	"testing"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"danmakustream/backend/internal/testutil"

	"gorm.io/gorm"
)

func TestSubscriptionExpirationAndFollowRepairWithMySQL(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t,
		&model.User{}, &model.Follow{}, &model.UserBlock{},
		&model.CreatorSubscription{},
	)
	viewer := model.User{Username: "expiry-viewer", Nickname: "Expiry Viewer", Password: "test", Role: "user"}
	expiredCreator := model.User{Username: "expiry-creator", Nickname: "Expired Creator", Password: "test", Role: "creator"}
	activeCreator := model.User{Username: "active-creator", Nickname: "Active Creator", Password: "test", Role: "creator"}
	orphanCreator := model.User{Username: "orphan-creator", Nickname: "Orphan Creator", Password: "test", Role: "creator"}
	for _, user := range []*model.User{&viewer, &expiredCreator, &activeCreator, &orphanCreator} {
		if err := db.Create(user).Error; err != nil {
			t.Fatal(err)
		}
	}
	now := time.Now()
	expired := model.CreatorSubscription{
		SubscriberID: viewer.ID, CreatorID: expiredCreator.ID, PriceCents: 500,
		Status: "active", AutoRenew: true, StartedAt: now.AddDate(0, -1, 0), ExpiresAt: now.Add(-time.Minute),
	}
	active := model.CreatorSubscription{
		SubscriberID: viewer.ID, CreatorID: activeCreator.ID, PriceCents: 500,
		Status: "active", StartedAt: now, ExpiresAt: now.AddDate(0, 1, 0),
	}
	if err := db.Create(&expired).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&active).Error; err != nil {
		t.Fatal(err)
	}
	for _, creatorID := range []uint{expiredCreator.ID, activeCreator.ID, orphanCreator.ID} {
		if err := db.Create(&model.Follow{FollowerID: viewer.ID, FolloweeID: creatorID, Special: true}).Error; err != nil {
			t.Fatal(err)
		}
	}

	ctx := &svc.ServiceContext{DB: db}
	if err := expireAll(ctx); err != nil {
		t.Fatal(err)
	}
	assertMembershipState(t, db, expired.ID, "expired", false)
	assertSpecialFollow(t, db, viewer.ID, expiredCreator.ID, false)
	assertSpecialFollow(t, db, viewer.ID, orphanCreator.ID, false)
	assertSpecialFollow(t, db, viewer.ID, activeCreator.ID, true)

	// Cover no-op paths for missing, already-expired and still-active subscriptions.
	if err := expireUserSubscription(ctx, viewer.ID, 999999); err != nil {
		t.Fatal(err)
	}
	if err := expireUserSubscription(ctx, viewer.ID, expiredCreator.ID); err != nil {
		t.Fatal(err)
	}
	if err := expireUserSubscription(ctx, viewer.ID, activeCreator.ID); err != nil {
		t.Fatal(err)
	}

	// Cover direct per-user expiration and its special-follow downgrade.
	directCreator := model.User{Username: "direct-creator", Nickname: "Direct Creator", Password: "test", Role: "creator"}
	if err := db.Create(&directCreator).Error; err != nil {
		t.Fatal(err)
	}
	direct := model.CreatorSubscription{
		SubscriberID: viewer.ID, CreatorID: directCreator.ID, PriceCents: 500,
		Status: "active", AutoRenew: true, StartedAt: now.AddDate(0, -1, 0), ExpiresAt: now.Add(-time.Minute),
	}
	if err := db.Create(&direct).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Follow{FollowerID: viewer.ID, FolloweeID: directCreator.ID, Special: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := expireUserSubscription(ctx, viewer.ID, directCreator.ID); err != nil {
		t.Fatal(err)
	}
	assertMembershipState(t, db, direct.ID, "expired", false)
	assertSpecialFollow(t, db, viewer.ID, directCreator.ID, false)

	// ensureFollow creates once and is idempotent on repeated calls.
	newCreator := model.User{Username: "new-creator", Nickname: "New Creator", Password: "test", Role: "creator"}
	if err := db.Create(&newCreator).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return ensureFollow(tx, viewer.ID, newCreator.ID) }); err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error { return ensureFollow(tx, viewer.ID, newCreator.ID) }); err != nil {
		t.Fatal(err)
	}
}

func TestSubscriptionExpirationStorageFailures(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t, &model.User{}, &model.Follow{}, &model.UserBlock{}, &model.CreatorSubscription{})
	ctx := &svc.ServiceContext{DB: db}
	if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("DROP TABLE creator_subscriptions").Error; err != nil {
		t.Fatal(err)
	}
	if err := expireAll(ctx); err == nil {
		t.Fatal("expireAll should fail without creator_subscriptions")
	}
	if err := expireUserSubscription(ctx, 1, 2); err == nil {
		t.Fatal("expireUserSubscription should fail without creator_subscriptions")
	}
	if _, err := hasBlockRelation(db, 1, 2); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("DROP TABLE user_blocks").Error; err != nil {
		t.Fatal(err)
	}
	if _, err := hasBlockRelation(db, 1, 2); err == nil {
		t.Fatal("hasBlockRelation should fail without user_blocks")
	}
}

func assertMembershipState(t *testing.T, db *gorm.DB, id uint, status string, autoRenew bool) {
	t.Helper()
	var item model.CreatorSubscription
	if err := db.First(&item, id).Error; err != nil {
		t.Fatal(err)
	}
	if item.Status != status || item.AutoRenew != autoRenew {
		t.Fatalf("subscription = status:%s autoRenew:%v", item.Status, item.AutoRenew)
	}
}

func assertSpecialFollow(t *testing.T, db *gorm.DB, followerID, followeeID uint, special bool) {
	t.Helper()
	var follow model.Follow
	if err := db.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).First(&follow).Error; err != nil {
		t.Fatal(err)
	}
	if follow.Special != special {
		t.Fatalf("follow %d -> %d special = %v", followerID, followeeID, follow.Special)
	}
}
