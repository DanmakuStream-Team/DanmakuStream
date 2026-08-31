//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"danmakustream/backend/internal/config"
	authhandler "danmakustream/backend/internal/handler/v1/auth"
	dynamichandler "danmakustream/backend/internal/handler/v1/dynamic"
	mediahandler "danmakustream/backend/internal/handler/v1/media"
	membershiphandler "danmakustream/backend/internal/handler/v1/membership"
	messagehandler "danmakustream/backend/internal/handler/v1/message"
	notificationhandler "danmakustream/backend/internal/handler/v1/notification"
	userhandler "danmakustream/backend/internal/handler/v1/user"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"danmakustream/backend/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestMemberBUseCasesWithMySQL(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t,
		&model.User{}, &model.Video{}, &model.Follow{}, &model.FollowGroup{}, &model.UserBlock{},
		&model.CreatorMembershipPlan{}, &model.CreatorSubscription{}, &model.SubscriptionOrder{},
		&model.Notification{}, &model.DynamicPost{}, &model.ChatMessage{},
	)
	testConfig := config.Config{}
	testConfig.Auth.AccessSecret = engagementTestSecret
	testConfig.Auth.AccessExpire = 3600
	svcCtx := &svc.ServiceContext{DB: db, Config: testConfig, VideoDir: t.TempDir()}
	router := newMemberBRouter(svcCtx)

	var memberID uint
	t.Run("UC01 register login and profile", func(t *testing.T) {
		register := requestJSON(t, router, http.MethodPost, "/auth/register", 0,
			map[string]any{"nickname": "b-member", "password": "password-123"})
		if register.Code != 0 {
			t.Fatalf("register failed: code=%d message=%q", register.Code, register.Message)
		}
		var session struct {
			Token    string `json:"token"`
			UserInfo struct {
				ID uint `json:"id"`
			} `json:"userInfo"`
		}
		if err := json.Unmarshal(register.Data, &session); err != nil || session.Token == "" || session.UserInfo.ID == 0 {
			t.Fatalf("invalid register session: %+v err=%v", session, err)
		}
		memberID = session.UserInfo.ID
		assertHTTPCode(t, router, http.MethodPost, "/auth/register", 0,
			map[string]any{"nickname": "b-member", "password": "password-123"}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/auth/login", 0,
			map[string]any{"nickname": "b-member", "password": "wrong"}, http.StatusBadRequest)
		login := requestJSON(t, router, http.MethodPost, "/auth/login", 0,
			map[string]any{"nickname": "b-member", "password": "password-123"})
		if login.Code != 0 {
			t.Fatalf("login failed: code=%d message=%q", login.Code, login.Message)
		}
		assertTokenStatus(t, router, "/auth/me", session.Token, http.StatusOK)
		assertTokenStatus(t, router, "/auth/me", "invalid-token", http.StatusUnauthorized)
		assertHTTPCode(t, router, http.MethodPut, "/users/me", memberID,
			map[string]any{"nickname": "b-member-updated", "bio": "member B"}, http.StatusOK)
		var user model.User
		if err := db.First(&user, memberID).Error; err != nil {
			t.Fatal(err)
		}
		if user.Nickname != "b-member-updated" || user.Password == "password-123" {
			t.Fatalf("profile/hash = nickname:%q plaintext:%v", user.Nickname, user.Password == "password-123")
		}
	})

	creator := model.User{Username: "b-creator", Nickname: "B Creator", Password: "hash", Role: "creator"}
	if err := db.Create(&creator).Error; err != nil {
		t.Fatal(err)
	}

	t.Run("UC07 follow group and block", func(t *testing.T) {
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", memberID), memberID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/users/999999/follow", memberID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), memberID, nil, http.StatusOK)
		groupResp := requestJSON(t, router, http.MethodPost, "/users/follow-groups", memberID, map[string]any{"name": "重点"})
		var group struct {
			ID uint `json:"id"`
		}
		if groupResp.Code != 0 || json.Unmarshal(groupResp.Data, &group) != nil || group.ID == 0 {
			t.Fatalf("create group failed: %+v", groupResp)
		}
		assertHTTPCode(t, router, http.MethodPost, "/users/follow-groups", memberID, map[string]any{"name": "重点"}, http.StatusConflict)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/follow-groups/%d", group.ID), memberID,
			map[string]any{"name": "核心关注"}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/users/follow-groups", memberID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), memberID,
			map[string]any{"groupId": group.ID}, http.StatusOK)
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ? AND group_id = ?", 1, memberID, creator.ID, group.ID)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), memberID,
			map[string]any{"groupId": group.ID + 9999}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, "/users/following", memberID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", creator.ID), memberID, nil, http.StatusOK)
		assertModelCount(t, db, &model.UserBlock{}, "blocker_id = ? AND blocked_id = ?", 1, memberID, creator.ID)
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ?", 0, memberID, creator.ID)
		assertHTTPCode(t, router, http.MethodGet, "/users/blocked", memberID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), memberID, nil, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", creator.ID), memberID, nil, http.StatusOK)
		assertModelCount(t, db, &model.UserBlock{}, "blocker_id = ? AND blocked_id = ?", 0, memberID, creator.ID)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), memberID, nil, http.StatusOK)
		dynamic := requestJSON(t, router, http.MethodPost, "/dynamics", creator.ID, map[string]any{"content": "UC07 follower notice"})
		if dynamic.Code != 0 {
			t.Fatalf("create dynamic failed: %+v", dynamic)
		}
		notices := requestJSON(t, router, http.MethodGet, "/notifications", memberID, nil)
		var noticeData struct {
			UnreadCount int64 `json:"unreadCount"`
			List        []struct {
				Type    string `json:"type"`
				Content string `json:"content"`
			} `json:"list"`
		}
		if notices.Code != 0 || json.Unmarshal(notices.Data, &noticeData) != nil || noticeData.UnreadCount != 1 || len(noticeData.List) != 1 || noticeData.List[0].Type != "dynamic" || noticeData.List[0].Content != "UC07 follower notice" {
			t.Fatalf("notification response = %+v data=%+v", notices, noticeData)
		}
		assertHTTPCode(t, router, http.MethodDelete, fmt.Sprintf("/users/follow-groups/%d", group.ID), memberID, nil, http.StatusOK)
		assertModelCount(t, db, &model.FollowGroup{}, "id = ?", 0, group.ID)
	})

	t.Run("UC08 membership payment is idempotent", func(t *testing.T) {
		assertHTTPCode(t, router, http.MethodPut, "/creator/membership-plan", memberID,
			map[string]any{"priceCents": 600, "benefits": "no role", "enabled": true}, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPut, "/creator/membership-plan", creator.ID,
			map[string]any{"priceCents": 99, "benefits": "invalid", "enabled": true}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, "/creator/membership-plan", creator.ID,
			map[string]any{"priceCents": 600, "benefits": "priority", "enabled": true}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/creators/%d/membership-plan", creator.ID), memberID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/creator/membership-plan", creator.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders", memberID,
			map[string]any{"creatorId": creator.ID, "months": 2}, http.StatusBadRequest)
		orderResp := requestJSON(t, router, http.MethodPost, "/subscriptions/orders", memberID,
			map[string]any{"creatorId": creator.ID, "months": 3})
		var order struct {
			OrderNo     string `json:"orderNo"`
			AmountCents int64  `json:"amountCents"`
		}
		if orderResp.Code != 0 || json.Unmarshal(orderResp.Data, &order) != nil || order.OrderNo == "" || order.AmountCents != 1800 {
			t.Fatalf("create order failed: %+v", orderResp)
		}
		payPath := "/subscriptions/orders/" + order.OrderNo + "/demo-pay"
		assertHTTPCode(t, router, http.MethodPost, payPath, memberID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, payPath, memberID, nil, http.StatusOK)
		assertModelCount(t, db, &model.CreatorSubscription{}, "subscriber_id = ? AND creator_id = ? AND status = ?", 1, memberID, creator.ID, "active")
		assertModelCount(t, db, &model.SubscriptionOrder{}, "order_no = ? AND status = ?", 1, order.OrderNo, "paid")
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ? AND special = ?", 1, memberID, creator.ID, true)
		assertHTTPCode(t, router, http.MethodGet, "/subscriptions", memberID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/subscriptions/orders", memberID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/subscriptions/creators/%d/status", creator.ID), memberID, nil, http.StatusOK)
		var firstSubscription model.CreatorSubscription
		if err := db.Where("subscriber_id = ? AND creator_id = ?", memberID, creator.ID).First(&firstSubscription).Error; err != nil {
			t.Fatal(err)
		}
		secondOrderResp := requestJSON(t, router, http.MethodPost, "/subscriptions/orders", memberID,
			map[string]any{"creatorId": creator.ID, "months": 1})
		var secondOrder struct {
			OrderNo string `json:"orderNo"`
		}
		if secondOrderResp.Code != 0 || json.Unmarshal(secondOrderResp.Data, &secondOrder) != nil || secondOrder.OrderNo == "" {
			t.Fatalf("second order failed: %+v", secondOrderResp)
		}
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders/"+secondOrder.OrderNo+"/demo-pay", memberID, nil, http.StatusOK)
		var renewed model.CreatorSubscription
		if err := db.Where("subscriber_id = ? AND creator_id = ?", memberID, creator.ID).First(&renewed).Error; err != nil || !renewed.ExpiresAt.After(firstSubscription.ExpiresAt) {
			t.Fatalf("renewed subscription = %+v err=%v", renewed, err)
		}
		if err := db.Model(&renewed).Update("auto_renew", true).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/subscriptions/%d/auto-renew", creator.ID), memberID,
			map[string]any{"enabled": false}, http.StatusOK)
		if err := db.Model(&renewed).Updates(map[string]any{"expires_at": time.Now().Add(-time.Hour), "status": "active", "auto_renew": true}).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ?", memberID, creator.ID).Update("special", true).Error; err != nil {
			t.Fatal(err)
		}
		statusResp := requestJSON(t, router, http.MethodGet, fmt.Sprintf("/subscriptions/creators/%d/status", creator.ID), memberID, nil)
		var statusData struct {
			Active bool `json:"active"`
		}
		if statusResp.Code != 0 || json.Unmarshal(statusResp.Data, &statusData) != nil || statusData.Active {
			t.Fatalf("expired status response = %+v data=%+v", statusResp, statusData)
		}
		assertModelCount(t, db, &model.CreatorSubscription{}, "subscriber_id = ? AND creator_id = ? AND status = ? AND auto_renew = ?", 1, memberID, creator.ID, "expired", false)
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ? AND special = ?", 1, memberID, creator.ID, false)
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders/missing/demo-pay", memberID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders", creator.ID,
			map[string]any{"creatorId": creator.ID, "months": 1}, http.StatusBadRequest)

		blockedCreator := model.User{Username: "b-blocked-creator", Nickname: "B Blocked Creator", Password: "hash", Role: "creator"}
		if err := db.Create(&blockedCreator).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&model.CreatorMembershipPlan{CreatorID: blockedCreator.ID, PriceCents: 500, Enabled: true}).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", blockedCreator.ID), memberID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders", memberID,
			map[string]any{"creatorId": blockedCreator.ID, "months": 1}, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", blockedCreator.ID), memberID, nil, http.StatusOK)
	})

	t.Run("UC11 message history unread and read", func(t *testing.T) {
		video := model.Video{Title: "B shared video", Status: "approved", AuthorID: creator.ID, VideoURL: "/media/videos/b-shared.mp4"}
		if err := db.Create(&video).Error; err != nil {
			t.Fatal(err)
		}
		send := requestJSON(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": creator.ID, "type": "text", "content": "hello"})
		if send.Code != 0 {
			t.Fatalf("send failed: code=%d message=%q", send.Code, send.Message)
		}
		assertModelCount(t, db, &model.ChatMessage{}, "sender_id = ? AND receiver_id = ? AND `read` = ?", 1, memberID, creator.ID, false)
		unread := requestJSON(t, router, http.MethodGet, "/messages/unread", creator.ID, nil)
		var unreadData struct {
			Count int64 `json:"count"`
		}
		if unread.Code != 0 || json.Unmarshal(unread.Data, &unreadData) != nil || unreadData.Count != 1 {
			t.Fatalf("unread response = %+v count=%d", unread, unreadData.Count)
		}
		conversations := requestJSON(t, router, http.MethodGet, "/messages/conversations", creator.ID, nil)
		var conversationData struct {
			List []struct {
				UnreadCount int64 `json:"unreadCount"`
			} `json:"list"`
		}
		if conversations.Code != 0 || json.Unmarshal(conversations.Data, &conversationData) != nil || len(conversationData.List) != 1 || conversationData.List[0].UnreadCount != 1 {
			t.Fatalf("conversation response = %+v data=%+v", conversations, conversationData)
		}
		share := requestJSON(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": creator.ID, "type": "video_share", "videoId": video.ID})
		if share.Code != 0 {
			t.Fatalf("video share failed: %+v", share)
		}
		assertHTTPCode(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": creator.ID, "type": "video_share", "videoId": video.ID + 9999}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": 999999, "type": "text", "content": "missing"}, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/messages/%d", memberID), creator.ID, nil, http.StatusOK)
		assertModelCount(t, db, &model.ChatMessage{}, "sender_id = ? AND receiver_id = ? AND `read` = ?", 2, memberID, creator.ID, true)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/messages/%d/read", memberID), creator.ID, nil, http.StatusOK)
		clientMessageID := "e2e-retry-safe-message"
		firstRetry := requestJSON(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": creator.ID, "clientMessageId": clientMessageID, "type": "text", "content": "retry once"})
		secondRetry := requestJSON(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": creator.ID, "clientMessageId": clientMessageID, "type": "text", "content": "retry once"})
		var firstRetryInfo, secondRetryInfo struct {
			ID uint `json:"id"`
		}
		if firstRetry.Code != 0 || secondRetry.Code != 0 || json.Unmarshal(firstRetry.Data, &firstRetryInfo) != nil || json.Unmarshal(secondRetry.Data, &secondRetryInfo) != nil || firstRetryInfo.ID == 0 || firstRetryInfo.ID != secondRetryInfo.ID {
			t.Fatalf("idempotent sends = first:%+v second:%+v", firstRetry, secondRetry)
		}
		assertModelCount(t, db, &model.ChatMessage{}, "sender_id = ? AND client_message_id = ?", 1, memberID, clientMessageID)
		assertHTTPCode(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": creator.ID, "type": "image", "mediaUrl": fmt.Sprintf("/media/messages/%d/20260829/test.png", memberID), "mediaName": "test.png"}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": creator.ID, "type": "video", "mediaUrl": fmt.Sprintf("/media/messages/%d/20260829/test.mp4", memberID), "mediaName": "test.mp4"}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": memberID, "type": "text", "content": "self"}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", memberID), creator.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": creator.ID, "type": "text", "content": "blocked"}, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", memberID), creator.ID, nil, http.StatusOK)
	})
}

func newMemberBRouter(svcCtx *svc.ServiceContext) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/auth/register", authhandler.RegisterHandler(svcCtx))
	router.POST("/auth/login", authhandler.LoginHandler(svcCtx))
	protected := router.Group("")
	protected.Use(middleware.AuthMiddleware(svcCtx.Config.Auth.AccessSecret))
	protected.Use(func(c *gin.Context) {
		var user model.User
		if svcCtx.DB.Select("role").First(&user, c.GetUint(middleware.CtxKeyUserID)).Error == nil {
			c.Set(middleware.CtxKeyRole, user.Role)
		}
		c.Next()
	})
	protected.GET("/auth/me", authhandler.MeHandler(svcCtx))
	protected.PUT("/users/me", userhandler.UpdateMeHandler(svcCtx))
	protected.GET("/users/following", userhandler.FollowingListHandler(svcCtx))
	protected.POST("/users/:id/follow", userhandler.FollowHandler(svcCtx))
	protected.GET("/users/follow-groups", userhandler.FollowGroupListHandler(svcCtx))
	protected.POST("/users/follow-groups", userhandler.CreateFollowGroupHandler(svcCtx))
	protected.PUT("/users/follow-groups/:id", userhandler.UpdateFollowGroupHandler(svcCtx))
	protected.DELETE("/users/follow-groups/:id", userhandler.DeleteFollowGroupHandler(svcCtx))
	protected.PUT("/users/:id/follow-settings", userhandler.UpdateFollowSettingsHandler(svcCtx))
	protected.GET("/users/blocked", userhandler.BlockedListHandler(svcCtx))
	protected.POST("/users/:id/block", userhandler.BlockHandler(svcCtx))
	protected.POST("/dynamics", dynamichandler.CreateHandler(svcCtx))
	protected.GET("/notifications", notificationhandler.ListHandler(svcCtx))
	protected.GET("/creators/:id/membership-plan", membershiphandler.PlanHandler(svcCtx))
	protected.GET("/creator/membership-plan", membershiphandler.MyPlanHandler(svcCtx))
	protected.PUT("/creator/membership-plan", membershiphandler.UpdateMyPlanHandler(svcCtx))
	protected.GET("/subscriptions", membershiphandler.MineHandler(svcCtx))
	protected.GET("/subscriptions/orders", membershiphandler.OrderListHandler(svcCtx))
	protected.GET("/subscriptions/creators/:id/status", membershiphandler.StatusHandler(svcCtx))
	protected.POST("/subscriptions/orders", membershiphandler.CreateOrderHandler(svcCtx))
	protected.POST("/subscriptions/orders/:orderNo/demo-pay", membershiphandler.DemoPayHandler(svcCtx))
	protected.PUT("/subscriptions/:creatorId/auto-renew", membershiphandler.AutoRenewHandler(svcCtx))
	protected.GET("/messages/conversations", messagehandler.ConversationListHandler(svcCtx))
	protected.GET("/messages/unread", messagehandler.UnreadHandler(svcCtx))
	protected.GET("/messages/:userId", messagehandler.HistoryHandler(svcCtx))
	protected.POST("/messages", messagehandler.SendHandler(svcCtx))
	protected.POST("/messages/media", mediahandler.UploadMessageMediaHandler(svcCtx))
	protected.PUT("/messages/:userId/read", messagehandler.ReadHandler(svcCtx))
	return router
}

func assertTokenStatus(t *testing.T, router http.Handler, path, token string, want int) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, bytes.NewReader(nil))
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != want {
		t.Fatalf("GET %s status=%d want=%d body=%s", path, recorder.Code, want, recorder.Body.String())
	}
}
