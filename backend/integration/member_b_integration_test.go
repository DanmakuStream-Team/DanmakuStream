//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"danmakustream/backend/internal/config"
	authhandler "danmakustream/backend/internal/handler/v1/auth"
	membershiphandler "danmakustream/backend/internal/handler/v1/membership"
	messagehandler "danmakustream/backend/internal/handler/v1/message"
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
		&model.Notification{}, &model.ChatMessage{},
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
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), memberID, nil, http.StatusOK)
		groupResp := requestJSON(t, router, http.MethodPost, "/users/follow-groups", memberID, map[string]any{"name": "重点"})
		var group struct {
			ID uint `json:"id"`
		}
		if groupResp.Code != 0 || json.Unmarshal(groupResp.Data, &group) != nil || group.ID == 0 {
			t.Fatalf("create group failed: %+v", groupResp)
		}
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/users/%d/follow-settings", creator.ID), memberID,
			map[string]any{"groupId": group.ID}, http.StatusOK)
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ? AND group_id = ?", 1, memberID, creator.ID, group.ID)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", creator.ID), memberID, nil, http.StatusOK)
		assertModelCount(t, db, &model.UserBlock{}, "blocker_id = ? AND blocked_id = ?", 1, memberID, creator.ID)
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ?", 0, memberID, creator.ID)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/follow", creator.ID), memberID, nil, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/users/%d/block", creator.ID), memberID, nil, http.StatusOK)
	})

	t.Run("UC08 membership payment is idempotent", func(t *testing.T) {
		assertHTTPCode(t, router, http.MethodPut, "/creator/membership-plan", creator.ID,
			map[string]any{"priceCents": 600, "benefits": "priority", "enabled": true}, http.StatusOK)
		orderResp := requestJSON(t, router, http.MethodPost, "/subscriptions/orders", memberID,
			map[string]any{"creatorId": creator.ID, "months": 3})
		var order struct {
			OrderNo string `json:"orderNo"`
		}
		if orderResp.Code != 0 || json.Unmarshal(orderResp.Data, &order) != nil || order.OrderNo == "" {
			t.Fatalf("create order failed: %+v", orderResp)
		}
		payPath := "/subscriptions/orders/" + order.OrderNo + "/demo-pay"
		assertHTTPCode(t, router, http.MethodPost, payPath, memberID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, payPath, memberID, nil, http.StatusOK)
		assertModelCount(t, db, &model.CreatorSubscription{}, "subscriber_id = ? AND creator_id = ? AND status = ?", 1, memberID, creator.ID, "active")
		assertModelCount(t, db, &model.SubscriptionOrder{}, "order_no = ? AND status = ?", 1, order.OrderNo, "paid")
		assertModelCount(t, db, &model.Follow{}, "follower_id = ? AND followee_id = ? AND special = ?", 1, memberID, creator.ID, true)
		assertHTTPCode(t, router, http.MethodPost, "/subscriptions/orders", creator.ID,
			map[string]any{"creatorId": creator.ID, "months": 1}, http.StatusBadRequest)
	})

	t.Run("UC11 message history unread and read", func(t *testing.T) {
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
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/messages/%d", memberID), creator.ID, nil, http.StatusOK)
		assertModelCount(t, db, &model.ChatMessage{}, "sender_id = ? AND receiver_id = ? AND `read` = ?", 1, memberID, creator.ID, true)
		assertHTTPCode(t, router, http.MethodPost, "/messages", memberID,
			map[string]any{"receiverId": memberID, "type": "text", "content": "self"}, http.StatusBadRequest)
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
	protected.POST("/users/:id/follow", userhandler.FollowHandler(svcCtx))
	protected.POST("/users/follow-groups", userhandler.CreateFollowGroupHandler(svcCtx))
	protected.PUT("/users/:id/follow-settings", userhandler.UpdateFollowSettingsHandler(svcCtx))
	protected.POST("/users/:id/block", userhandler.BlockHandler(svcCtx))
	protected.PUT("/creator/membership-plan", membershiphandler.UpdateMyPlanHandler(svcCtx))
	protected.POST("/subscriptions/orders", membershiphandler.CreateOrderHandler(svcCtx))
	protected.POST("/subscriptions/orders/:orderNo/demo-pay", membershiphandler.DemoPayHandler(svcCtx))
	protected.GET("/messages/unread", messagehandler.UnreadHandler(svcCtx))
	protected.GET("/messages/:userId", messagehandler.HistoryHandler(svcCtx))
	protected.POST("/messages", messagehandler.SendHandler(svcCtx))
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
