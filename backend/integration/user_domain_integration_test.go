//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"danmakustream/backend/internal/config"
	authhandler "danmakustream/backend/internal/handler/v1/auth"
	membershiphandler "danmakustream/backend/internal/handler/v1/membership"
	messagehandler "danmakustream/backend/internal/handler/v1/message"
	notificationhandler "danmakustream/backend/internal/handler/v1/notification"
	userhandler "danmakustream/backend/internal/handler/v1/user"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestUserDomainUseCasesWithMySQL(t *testing.T) {
	db := openTemporaryMySQL(t)
	testConfig := config.Config{}
	testConfig.Auth.AccessSecret = engagementTestSecret
	testConfig.Auth.AccessExpire = 3600
	svcCtx := &svc.ServiceContext{DB: db, Config: testConfig, VideoDir: t.TempDir()}
	router := newUserDomainRouter(svcCtx)

	viewer := registerUser(t, router, "UC User", "Test1234!")
	creator := registerUser(t, router, "UC Creator", "Creator1234!")
	other := registerUser(t, router, "UC Other", "Other1234!")
	if err := db.Model(&model.User{}).Where("id = ?", creator.ID).Update("role", "creator").Error; err != nil {
		t.Fatal(err)
	}
	creator.Role = "creator"

	t.Run("UC01 register login authorization and profile persistence", func(t *testing.T) {
		assertHTTPCode(t, router, http.MethodPost, "/auth/register", 0, "malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/auth/register", 0,
			map[string]any{"nickname": " ", "password": ""}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/auth/register", 0,
			map[string]any{"nickname": viewer.Nickname, "password": "Test1234!"}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/auth/login", 0,
			map[string]any{"nickname": viewer.Nickname, "password": "wrong"}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/auth/login", 0,
			map[string]any{"nickname": "missing", "password": "Test1234!"}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/auth/login", 0,
			map[string]any{"nickname": " ", "password": "Test1234!"}, http.StatusBadRequest)

		login := requestJSON(t, router, http.MethodPost, "/auth/login", 0,
			map[string]any{"username": viewer.Nickname, "password": "Test1234!"})
		if login.Code != 0 || !strings.Contains(string(login.Data), `"token"`) {
			t.Fatalf("legacy username login failed: %+v", login)
		}
		assertHTTPCode(t, router, http.MethodGet, "/auth/me", 0, nil, http.StatusUnauthorized)
		assertHTTPCode(t, router, http.MethodGet, "/auth/me", viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/auth/me", 999999, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPut, "/users/me", viewer.ID, "malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, "/users/me", viewer.ID, map[string]any{}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, "/users/me", viewer.ID,
			map[string]any{"nickname": "UC User Updated", "bio": "profile persisted"}, http.StatusOK)

		var refreshed model.User
		if err := db.First(&refreshed, viewer.ID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.Nickname != "UC User Updated" || refreshed.Bio != "profile persisted" {
			t.Fatalf("profile = %q/%q", refreshed.Nickname, refreshed.Bio)
		}
		if refreshed.Username != fmt.Sprint(viewer.ID) || bcrypt.CompareHashAndPassword([]byte(refreshed.Password), []byte("Test1234!")) != nil {
			t.Fatalf("registered credentials were not persisted safely")
		}
		viewer.Nickname = refreshed.Nickname

		assertHTTPCode(t, router, http.MethodGet, "/users/not-a-number", 0, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, "/users/999999", 0, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/users/%d", viewer.ID), 0, nil, http.StatusOK)
		assertHTTPCodeWithToken(t, router, http.MethodGet, fmt.Sprintf("/users/%d", creator.ID), "invalid-token", nil, http.StatusOK)
		assertAvatarUpload(t, router, viewer.ID, "avatar.png", []byte("fake image payload"), http.StatusOK)
		assertAvatarUpload(t, router, viewer.ID, "avatar", []byte("no extension"), http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/users/me/avatar", viewer.ID, nil, http.StatusBadRequest)
	})

	t.Run("UC07 follow groups notifications and blocking", func(t *testing.T) {
		assertHTTPCode(t, router, http.MethodPost, "/users/not-a-number/follow", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", viewer.ID), viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/users/999999/follow", viewer.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), viewer.ID, nil, http.StatusOK)
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ?", 1, viewer.ID, creator.ID)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", viewer.ID), creator.ID, nil, http.StatusOK)

		assertHTTPCode(t, router, http.MethodPost, "/users/follow-groups", viewer.ID, map[string]any{"name": " "}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/users/follow-groups", viewer.ID, "malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/users/follow-groups", viewer.ID, map[string]any{"name": strings.Repeat("长", 21)}, http.StatusBadRequest)
		firstGroup := requestJSON(t, router, http.MethodPost, "/users/follow-groups", viewer.ID, map[string]any{"name": " Favorites "})
		groupID := decodeID(t, firstGroup)
		secondGroup := requestJSON(t, router, http.MethodPost, "/users/follow-groups", viewer.ID, map[string]any{"name": "Creators"})
		secondGroupID := decodeID(t, secondGroup)
		assertHTTPCode(t, router, http.MethodPost, "/users/follow-groups", viewer.ID, map[string]any{"name": "Favorites"}, http.StatusConflict)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), viewer.ID,
			map[string]any{"special": true}, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), viewer.ID,
			map[string]any{}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), viewer.ID,
			map[string]any{"groupId": 999999}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), viewer.ID,
			map[string]any{"groupId": groupID}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/users/%d", creator.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), viewer.ID,
			map[string]any{"groupId": 0}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), viewer.ID,
			map[string]any{"groupId": groupID}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/users/following", viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/users/follow-groups", viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/follow-groups/%d", groupID), viewer.ID,
			map[string]any{"name": "Creators"}, http.StatusConflict)
		assertHTTPCode(t, router, http.MethodPut, "/users/follow-groups/not-a-number", viewer.ID,
			map[string]any{"name": "bad"}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/follow-groups/%d", groupID), viewer.ID,
			"malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/follow-groups/%d", groupID), viewer.ID,
			map[string]any{"name": " "}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/follow-groups/%d", groupID), viewer.ID,
			map[string]any{"name": strings.Repeat("长", 21)}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, "/users/follow-groups/999999", viewer.ID,
			map[string]any{"name": "Missing"}, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/follow-groups/%d", groupID), viewer.ID,
			map[string]any{"name": "Priority"}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodDelete, fmt.Sprintf("/users/follow-groups/%d", secondGroupID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodDelete, "/users/follow-groups/999999", viewer.ID, nil, http.StatusNotFound)

		actorID := creator.ID
		notification := model.Notification{UserID: viewer.ID, ActorID: &actorID, Type: "dynamic", Title: "new content", Link: "/dynamics"}
		if err := db.Create(&notification).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodGet, "/notifications?read=bad", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, "/notifications?read=false&page=0&pageSize=500", viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, "/notifications/not-a-number/read", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, "/notifications/999999/read", viewer.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/notifications/%d/read", notification.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, "/notifications", viewer.ID, nil, http.StatusOK)

		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", viewer.ID), viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/users/999999/block", viewer.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", creator.ID), viewer.ID, nil, http.StatusOK)
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ?", 0, viewer.ID, creator.ID)
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ?", 0, creator.ID, viewer.ID)
		assertHTTPCode(t, router, http.MethodGet, "/users/blocked", viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/users/%d", creator.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), viewer.ID, nil, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", creator.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), viewer.ID,
			map[string]any{"groupId": 0}, http.StatusBadRequest)
	})

	t.Run("UC08 membership order payment idempotency and expiration", func(t *testing.T) {
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/creators/%d/membership-plan", other.ID), 0, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/creator/membership-plan", creator.ID, nil, http.StatusOK)
		assertHTTPCodeAs(t, router, http.MethodPut, "/creator/membership-plan", viewer.ID, "user",
			map[string]any{"priceCents": 500, "enabled": true}, http.StatusForbidden)
		assertHTTPCodeAs(t, router, http.MethodPut, "/creator/membership-plan", creator.ID, creator.Role,
			"malformed", http.StatusBadRequest)
		assertHTTPCodeAs(t, router, http.MethodPut, "/creator/membership-plan", creator.ID, creator.Role,
			map[string]any{"priceCents": 99, "enabled": true}, http.StatusBadRequest)
		assertHTTPCodeAs(t, router, http.MethodPut, "/creator/membership-plan", creator.ID, creator.Role,
			map[string]any{"priceCents": 800, "benefits": strings.Repeat("权", 201), "enabled": true}, http.StatusBadRequest)
		assertHTTPCodeAs(t, router, http.MethodPut, "/creator/membership-plan", creator.ID, creator.Role,
			map[string]any{"priceCents": 800, "benefits": "member priority", "enabled": true}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/creators/%d/membership-plan", creator.ID), 0, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/creators/not-a-number/membership-plan", 0, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, "/creators/999999/membership-plan", 0, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodGet, "/creator/membership-plan", creator.ID, nil, http.StatusOK)
		assertHTTPCodeAs(t, router, http.MethodPut, "/creator/membership-plan", creator.ID, creator.Role,
			map[string]any{"priceCents": 900, "benefits": "updated member priority", "enabled": true}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/subscriptions/creators/%d/status", creator.ID), viewer.ID, nil, http.StatusOK)

		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders", viewer.ID, "malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders", viewer.ID,
			map[string]any{"creatorId": viewer.ID, "months": 1}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders", viewer.ID,
			map[string]any{"creatorId": creator.ID, "months": 2}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders", viewer.ID,
			map[string]any{"creatorId": other.ID, "months": 1}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", creator.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders", viewer.ID,
			map[string]any{"creatorId": creator.ID, "months": 1}, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", creator.ID), viewer.ID, nil, http.StatusOK)
		order := requestJSON(t, router, http.MethodPost, "/subscriptions/orders", viewer.ID,
			map[string]any{"creatorId": creator.ID, "months": 3})
		orderNo := decodeString(t, order, "orderNo")
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders/missing/demo-pay", viewer.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/subscriptions/orders/%s/demo-pay", orderNo), other.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/subscriptions/orders/%s/demo-pay", orderNo), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/subscriptions/orders/%s/demo-pay", orderNo), viewer.ID, nil, http.StatusOK)
		assertModelCount(t, db, &model.CreatorSubscription{}, "subscriber_id = ? AND creator_id = ?", 1, viewer.ID, creator.ID)
		assertModelCount(t, db, &model.Notification{}, "user_id = ? AND type = ?", 1, creator.ID, "membership")
		secondOrder := requestJSON(t, router, http.MethodPost, "/subscriptions/orders", viewer.ID,
			map[string]any{"creatorId": creator.ID, "months": 1})
		secondOrderNo := decodeString(t, secondOrder, "orderNo")
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/subscriptions/orders/%s/demo-pay", secondOrderNo), viewer.ID, nil, http.StatusOK)
		assertModelCount(t, db, &model.CreatorSubscription{}, "subscriber_id = ? AND creator_id = ?", 1, viewer.ID, creator.ID)
		canceledOrder := requestJSON(t, router, http.MethodPost, "/subscriptions/orders", viewer.ID,
			map[string]any{"creatorId": creator.ID, "months": 1})
		canceledOrderNo := decodeString(t, canceledOrder, "orderNo")
		if err := db.Model(&model.SubscriptionOrder{}).Where("order_no = ?", canceledOrderNo).Update("status", "canceled").Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/subscriptions/orders/%s/demo-pay", canceledOrderNo), viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/subscriptions/creators/%d/status", creator.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/subscriptions", viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/subscriptions/orders", viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), viewer.ID, nil, http.StatusConflict)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/subscriptions/%d/auto-renew", creator.ID), viewer.ID,
			map[string]any{"enabled": true}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/subscriptions/%d/auto-renew", creator.ID), viewer.ID,
			"malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, "/subscriptions/not-a-number/auto-renew", viewer.ID,
			map[string]any{"enabled": false}, http.StatusBadRequest)
		if err := db.Model(&model.CreatorSubscription{}).Where("subscriber_id = ? AND creator_id = ?", viewer.ID, creator.ID).Update("auto_renew", true).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/subscriptions/%d/auto-renew", creator.ID), viewer.ID,
			map[string]any{"enabled": false}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, "/subscriptions/999999/auto-renew", viewer.ID,
			map[string]any{"enabled": false}, http.StatusNotFound)

		if err := db.Model(&model.CreatorSubscription{}).Where("subscriber_id = ? AND creator_id = ?", viewer.ID, creator.ID).
			Updates(map[string]any{"status": "active", "expires_at": time.Now().Add(-time.Hour), "auto_renew": true}).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodGet, "/subscriptions", viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/subscriptions/creators/%d/status", creator.ID), viewer.ID, nil, http.StatusOK)
		var expired model.CreatorSubscription
		if err := db.Where("subscriber_id = ? AND creator_id = ?", viewer.ID, creator.ID).First(&expired).Error; err != nil {
			t.Fatal(err)
		}
		if expired.Status != "expired" || expired.AutoRenew {
			t.Fatalf("expired subscription = status:%s autoRenew:%v", expired.Status, expired.AutoRenew)
		}
	})

	t.Run("UC11 messages media sharing unread history and block rules", func(t *testing.T) {
		textMessage := requestJSON(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "text", "content": " hello "})
		if decodeID(t, textMessage) == 0 {
			t.Fatal("text message was not created")
		}
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "text", "content": " "}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID, "malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "text", "content": strings.Repeat("a", 2001)}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": viewer.ID, "type": "text", "content": "self"}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": 999999, "type": "text", "content": "missing"}, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "file", "content": "bad"}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "image", "mediaUrl": "/media/messages/999/a.png"}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "image", "mediaUrl": fmt.Sprintf("/media/messages/%d/a.png", viewer.ID)}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "video", "mediaUrl": fmt.Sprintf("/media/messages/%d/clip.webm", viewer.ID)}, http.StatusOK)

		approved := model.Video{Title: "shared", Status: "approved", AuthorID: creator.ID}
		pending := model.Video{Title: "pending share", Status: "pending", AuthorID: creator.ID}
		if err := db.Create(&approved).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&pending).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "video_share", "videoId": approved.ID}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "video_share", "videoId": pending.ID}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/messages", other.ID,
			map[string]any{"receiverId": viewer.ID, "type": "text", "content": "reply"}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/messages/unread", other.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/messages/conversations", other.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/messages/not-a-number", other.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/messages/%d?page=0&pageSize=500", viewer.ID), other.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/messages/unread", other.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, "/messages/not-a-number/read", other.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/messages/%d/read", viewer.ID), other.ID, nil, http.StatusOK)

		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", viewer.ID), other.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": other.ID, "type": "text", "content": "blocked"}, http.StatusForbidden)
	})
}

func TestUserDomainDatabaseFailureResponses(t *testing.T) {
	newEnvironment := func(t *testing.T) (*gin.Engine, *gorm.DB, model.UserInfo, model.UserInfo) {
		t.Helper()
		db := openTemporaryMySQL(t)
		testConfig := config.Config{}
		testConfig.Auth.AccessSecret = engagementTestSecret
		testConfig.Auth.AccessExpire = 3600
		router := newUserDomainRouter(&svc.ServiceContext{DB: db, Config: testConfig, VideoDir: t.TempDir()})
		viewer := registerUser(t, router, "Failure Viewer", "Test1234!")
		creator := registerUser(t, router, "Failure Creator", "Creator1234!")
		if err := db.Model(&model.User{}).Where("id = ?", creator.ID).Update("role", "creator").Error; err != nil {
			t.Fatal(err)
		}
		return router, db, viewer, creator
	}
	drop := func(t *testing.T, db *gorm.DB, table string) {
		t.Helper()
		if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("DROP TABLE `" + table + "`").Error; err != nil {
			t.Fatal(err)
		}
	}

	t.Run("UC07 relationship and notification storage failures", func(t *testing.T) {
		router, db, viewer, creator := newEnvironment(t)
		drop(t, db, "follow_groups")
		assertHTTPCode(t, router, http.MethodGet, "/users/follow-groups", viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, "/users/follow-groups", viewer.ID,
			map[string]any{"name": "broken"}, http.StatusInternalServerError)

		router, db, viewer, creator = newEnvironment(t)
		drop(t, db, "user_blocks")
		assertHTTPCode(t, router, http.MethodGet, "/users/blocked", viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/users/%d", creator.ID), viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", creator.ID), viewer.ID, nil, http.StatusInternalServerError)

		router, db, viewer, creator = newEnvironment(t)
		drop(t, db, "follows")
		assertHTTPCode(t, router, http.MethodGet, "/users/following", viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/users/%d", creator.ID), viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), viewer.ID, nil, http.StatusInternalServerError)

		router, db, viewer, _ = newEnvironment(t)
		drop(t, db, "notifications")
		assertHTTPCode(t, router, http.MethodGet, "/notifications", viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPut, "/notifications/1/read", viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPut, "/notifications", viewer.ID, nil, http.StatusInternalServerError)
	})

	t.Run("UC01 profile storage failures", func(t *testing.T) {
		router, db, viewer, _ := newEnvironment(t)
		drop(t, db, "videos")
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/users/%d", viewer.ID), viewer.ID, nil, http.StatusInternalServerError)

		router, db, viewer, _ = newEnvironment(t)
		drop(t, db, "users")
		assertHTTPCode(t, router, http.MethodPut, "/users/me", viewer.ID,
			map[string]any{"bio": "cannot persist"}, http.StatusInternalServerError)
		assertAvatarUpload(t, router, viewer.ID, "avatar.png", []byte("cannot persist"), http.StatusInternalServerError)
	})

	t.Run("UC08 membership storage failures", func(t *testing.T) {
		router, db, viewer, creator := newEnvironment(t)
		drop(t, db, "creator_membership_plans")
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/creators/%d/membership-plan", creator.ID), 0, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, "/creator/membership-plan", creator.ID, nil, http.StatusInternalServerError)
		assertHTTPCodeAs(t, router, http.MethodPut, "/creator/membership-plan", creator.ID, "creator",
			map[string]any{"priceCents": 500, "enabled": true}, http.StatusInternalServerError)

		router, db, viewer, creator = newEnvironment(t)
		drop(t, db, "creator_subscriptions")
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/subscriptions/creators/%d/status", creator.ID), viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, "/subscriptions", viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/subscriptions/%d/auto-renew", creator.ID), viewer.ID,
			map[string]any{"enabled": false}, http.StatusInternalServerError)

		router, db, viewer, _ = newEnvironment(t)
		drop(t, db, "subscription_orders")
		assertHTTPCode(t, router, http.MethodGet, "/subscriptions/orders", viewer.ID, nil, http.StatusInternalServerError)
	})

	t.Run("UC11 message storage failures", func(t *testing.T) {
		router, db, viewer, creator := newEnvironment(t)
		drop(t, db, "chat_messages")
		assertHTTPCode(t, router, http.MethodGet, "/messages/conversations", viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/messages/%d", creator.ID), viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, "/messages/unread", viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/messages/%d/read", creator.ID), viewer.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, "/messages", viewer.ID,
			map[string]any{"receiverId": creator.ID, "type": "text", "content": "storage failure"}, http.StatusBadRequest)
	})
}

func newUserDomainRouter(svcCtx *svc.ServiceContext) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", authhandler.RegisterHandler(svcCtx))
	router.POST("/auth/login", authhandler.LoginHandler(svcCtx))
	router.GET("/users/:id", userhandler.ProfileHandler(svcCtx))
	router.GET("/creators/:id/membership-plan", membershiphandler.PlanHandler(svcCtx))
	auth := router.Group("")
	auth.Use(middleware.AuthMiddleware(svcCtx.Config.Auth.AccessSecret))
	{
		auth.GET("/auth/me", authhandler.MeHandler(svcCtx))
		auth.PUT("/users/me", userhandler.UpdateMeHandler(svcCtx))
		auth.POST("/users/me/avatar", userhandler.UploadAvatarHandler(svcCtx))
		auth.GET("/users/following", userhandler.FollowingListHandler(svcCtx))
		auth.POST("/users/:id/follow", userhandler.FollowHandler(svcCtx))
		auth.GET("/users/follow-groups", userhandler.FollowGroupListHandler(svcCtx))
		auth.POST("/users/follow-groups", userhandler.CreateFollowGroupHandler(svcCtx))
		auth.PUT("/users/follow-groups/:id", userhandler.UpdateFollowGroupHandler(svcCtx))
		auth.DELETE("/users/follow-groups/:id", userhandler.DeleteFollowGroupHandler(svcCtx))
		auth.PUT("/users/:id/follow-settings", userhandler.UpdateFollowSettingsHandler(svcCtx))
		auth.GET("/users/blocked", userhandler.BlockedListHandler(svcCtx))
		auth.POST("/users/:id/block", userhandler.BlockHandler(svcCtx))
		auth.GET("/notifications", notificationhandler.ListHandler(svcCtx))
		auth.PUT("/notifications", notificationhandler.ReadAllHandler(svcCtx))
		auth.PUT("/notifications/:id/read", notificationhandler.ReadHandler(svcCtx))
		auth.GET("/creator/membership-plan", membershiphandler.MyPlanHandler(svcCtx))
		auth.PUT("/creator/membership-plan", membershiphandler.UpdateMyPlanHandler(svcCtx))
		auth.GET("/subscriptions", membershiphandler.MineHandler(svcCtx))
		auth.GET("/subscriptions/orders", membershiphandler.OrderListHandler(svcCtx))
		auth.GET("/subscriptions/creators/:id/status", membershiphandler.StatusHandler(svcCtx))
		auth.POST("/subscriptions/orders", membershiphandler.CreateOrderHandler(svcCtx))
		auth.POST("/subscriptions/orders/:orderNo/demo-pay", membershiphandler.DemoPayHandler(svcCtx))
		auth.PUT("/subscriptions/:creatorId/auto-renew", membershiphandler.AutoRenewHandler(svcCtx))
		auth.GET("/messages/conversations", messagehandler.ConversationListHandler(svcCtx))
		auth.GET("/messages/unread", messagehandler.UnreadHandler(svcCtx))
		auth.GET("/messages/:userId", messagehandler.HistoryHandler(svcCtx))
		auth.POST("/messages", messagehandler.SendHandler(svcCtx))
		auth.PUT("/messages/:userId/read", messagehandler.ReadHandler(svcCtx))
	}
	return router
}

func registerUser(t *testing.T, router http.Handler, nickname, password string) model.UserInfo {
	t.Helper()
	result := requestJSON(t, router, http.MethodPost, "/auth/register", 0,
		map[string]any{"nickname": nickname, "password": password})
	if result.Code != 0 {
		t.Fatalf("register %q failed: code=%d message=%q", nickname, result.Code, result.Message)
	}
	var payload struct {
		UserInfo model.UserInfo `json:"userInfo"`
	}
	if err := json.Unmarshal(result.Data, &payload); err != nil || payload.UserInfo.ID == 0 {
		t.Fatalf("decode registered user %q: %+v err=%v", nickname, payload, err)
	}
	return payload.UserInfo
}

func decodeID(t *testing.T, envelope responseEnvelope) uint {
	t.Helper()
	if envelope.Code != 0 {
		t.Fatalf("request failed: code=%d message=%q", envelope.Code, envelope.Message)
	}
	var payload struct {
		ID uint `json:"id"`
	}
	if err := json.Unmarshal(envelope.Data, &payload); err != nil || payload.ID == 0 {
		t.Fatalf("decode ID: data=%s err=%v", envelope.Data, err)
	}
	return payload.ID
}

func decodeString(t *testing.T, envelope responseEnvelope, key string) string {
	t.Helper()
	if envelope.Code != 0 {
		t.Fatalf("request failed: code=%d message=%q", envelope.Code, envelope.Message)
	}
	var payload map[string]any
	if err := json.Unmarshal(envelope.Data, &payload); err != nil {
		t.Fatal(err)
	}
	value, _ := payload[key].(string)
	if value == "" {
		t.Fatalf("missing string field %q in %s", key, envelope.Data)
	}
	return value
}

func roleToken(t *testing.T, userID uint, role string) string {
	t.Helper()
	claims := middleware.Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(engagementTestSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func assertHTTPCodeAs(t *testing.T, router http.Handler, method, path string, userID uint, role string, body any, want int) {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+roleToken(t, userID, role))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != want {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, path, recorder.Code, want, recorder.Body.String())
	}
}

func assertHTTPCodeWithToken(t *testing.T, router http.Handler, method, path, token string, body any, want int) {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != want {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, path, recorder.Code, want, recorder.Body.String())
	}
}

func assertAvatarUpload(t *testing.T, router http.Handler, userID uint, filename string, content []byte, want int) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("avatar", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/users/me/avatar", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+testToken(t, userID))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != want {
		t.Fatalf("avatar upload status = %d, want %d; body=%s", recorder.Code, want, recorder.Body.String())
	}
}
