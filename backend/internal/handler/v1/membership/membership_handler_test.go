package membership

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
)

func TestUpdateMyPlanHandlerValidation(t *testing.T) {
	tests := []struct {
		name   string
		role   string
		body   string
		status int
	}{
		{name: "ordinary user forbidden", role: "user", body: `{"priceCents":500}`, status: http.StatusForbidden},
		{name: "malformed json", role: "creator", body: `{`, status: http.StatusBadRequest},
		{name: "price below minimum", role: "creator", body: `{"priceCents":99}`, status: http.StatusBadRequest},
		{name: "price above maximum", role: "creator", body: `{"priceCents":100001}`, status: http.StatusBadRequest},
		{name: "benefits too long", role: "creator", body: `{"priceCents":500,"benefits":"` + strings.Repeat("权", 201) + `"}`, status: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := serveMembershipHandler(t, http.MethodPut, "/plan", test.body, 7, test.role, "/plan", UpdateMyPlanHandler(&svc.ServiceContext{}))
			if status != test.status {
				t.Fatalf("status = %d, want %d", status, test.status)
			}
		})
	}
}

func TestCreateOrderHandlerValidation(t *testing.T) {
	tests := []struct {
		name   string
		body   string
		status int
	}{
		{name: "malformed json", body: `{`, status: http.StatusBadRequest},
		{name: "missing creator", body: `{"months":1}`, status: http.StatusBadRequest},
		{name: "subscribe self", body: `{"creatorId":7,"months":1}`, status: http.StatusBadRequest},
		{name: "unsupported months", body: `{"creatorId":9,"months":2}`, status: http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := serveMembershipHandler(t, http.MethodPost, "/orders", test.body, 7, "user", "/orders", CreateOrderHandler(&svc.ServiceContext{}))
			if status != test.status {
				t.Fatalf("status = %d, want %d", status, test.status)
			}
		})
	}
}

func TestPlanHandlerRejectsInvalidCreatorID(t *testing.T) {
	status := serveMembershipHandler(t, http.MethodGet, "/creators/nope/plan", "", 7, "user", "/creators/:id/plan", PlanHandler(&svc.ServiceContext{}))
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}
}

func TestMembershipInfoFormatting(t *testing.T) {
	now := time.Now()
	paidAt := now.Add(-time.Hour)
	creator := model.User{}
	creator.ID = 11
	creator.Nickname = "creator"
	subscription := model.CreatorSubscription{
		CreatorID: 11, Creator: creator, PriceCents: 500, Status: "active",
		StartedAt: now, ExpiresAt: now.Add(49 * time.Hour),
	}
	info := toSubscriptionInfo(subscription)
	if info.Creator.ID != 11 || info.DaysRemaining != 3 {
		t.Fatalf("subscription info = %+v", info)
	}
	order := model.SubscriptionOrder{Creator: creator, PaidAt: &paidAt}
	if got := toOrderInfo(order); got.PaidAt == nil || *got.PaidAt == "" {
		t.Fatalf("paid order should expose paidAt: %+v", got)
	}
	expired := subscription
	expired.ExpiresAt = now.Add(-time.Hour)
	if got := toSubscriptionInfo(expired).DaysRemaining; got != 0 {
		t.Fatalf("expired days = %d, want 0", got)
	}
}

func TestRenewalExpiry(t *testing.T) {
	now := time.Date(2026, 8, 29, 12, 0, 0, 0, time.UTC)
	future := now.AddDate(0, 1, 0)

	active := model.CreatorSubscription{Status: "active", ExpiresAt: future}
	if got, want := renewalExpiry(active, now, 3), future.AddDate(0, 3, 0); !got.Equal(want) {
		t.Fatalf("active renewal = %v, want %v", got, want)
	}

	expired := model.CreatorSubscription{Status: "expired", ExpiresAt: now.Add(-time.Hour)}
	if got, want := renewalExpiry(expired, now, 1), now.AddDate(0, 1, 0); !got.Equal(want) {
		t.Fatalf("expired renewal = %v, want %v", got, want)
	}
}

func serveMembershipHandler(t *testing.T, method, path, body string, userID uint, role, route string, handler gin.HandlerFunc) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.CtxKeyUserID, userID)
		c.Set(middleware.CtxKeyRole, role)
		c.Next()
	})
	router.Handle(method, route, handler)
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder.Code
}
