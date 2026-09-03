//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"danmakustream/user-service/internal/client"
	"danmakustream/user-service/internal/config"
	"danmakustream/user-service/internal/middleware"
	model "danmakustream/user-service/internal/model/mysql"
	"danmakustream/user-service/internal/server"
	"danmakustream/user-service/internal/svc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestAllPublicAndInternalAPIRoutes(t *testing.T) {
	dsn := os.Getenv("USER_SERVICE_TEST_DSN")
	if dsn == "" {
		t.Skip("USER_SERVICE_TEST_DSN is not configured")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(userModels()...); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()

	suffix := time.Now().UnixNano()
	admin := model.User{Username: fmt.Sprintf("reg-admin-%d", suffix), Nickname: fmt.Sprintf("回归管理员-%d", suffix), Password: "hash", Role: "admin"}
	member := model.User{Username: fmt.Sprintf("reg-member-%d", suffix), Nickname: fmt.Sprintf("回归用户-%d", suffix), Password: "hash", Role: "user"}
	if err := tx.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&member).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&model.Follow{FollowerID: admin.ID, FolloweeID: member.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&model.UserBlock{BlockerID: member.ID, BlockedID: admin.ID}).Error; err != nil {
		t.Fatal(err)
	}
	if err := tx.Create(&model.CreatorSubscription{SubscriberID: admin.ID, CreatorID: member.ID, PriceCents: 500, Status: "active", StartedAt: time.Now(), ExpiresAt: time.Now().Add(time.Hour)}).Error; err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{Name: "user-service", InternalToken: "regression-internal-token", Version: "day8-test", Commit: "test", BuildTime: "now", RequestTimeout: time.Second}
	cfg.Auth.AccessSecret = "regression-jwt-secret"
	ctx := &svc.ServiceContext{Config: cfg, DB: tx, VideoDir: t.TempDir(), Content: client.NewContent("http://127.0.0.1:1", cfg.InternalToken, 50*time.Millisecond)}
	gin.SetMode(gin.TestMode)
	router := server.Router(ctx)
	bearer := regressionToken(t, cfg.Auth.AccessSecret, admin)
	assertInternalBool(t, router, cfg.InternalToken, fmt.Sprintf("/internal/v1/users/%d/exists", admin.ID), "exists", true)
	assertInternalBool(t, router, cfg.InternalToken, fmt.Sprintf("/internal/v1/relationships/following?followerId=%d&followeeId=%d", admin.ID, member.ID), "following", true)
	assertInternalBool(t, router, cfg.InternalToken, fmt.Sprintf("/internal/v1/relationships/blocked?firstId=%d&secondId=%d", admin.ID, member.ID), "blocked", true)
	assertInternalBool(t, router, cfg.InternalToken, fmt.Sprintf("/internal/v1/memberships/status?subscriberId=%d&creatorId=%d", admin.ID, member.ID), "active", true)

	tested := 0
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/v1/") && !strings.HasPrefix(route.Path, "/internal/v1/") {
			continue
		}
		tested++
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			path := regressionPath(route.Path, admin.ID, member.ID)
			body := io.Reader(nil)
			if route.Method == http.MethodPost || route.Method == http.MethodPut {
				body = bytes.NewBufferString(`{}`)
			}
			req := httptest.NewRequest(route.Method, path, body)
			req.Header.Set("X-Request-ID", "day8-route-regression")
			if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			if strings.HasPrefix(route.Path, "/internal/v1/") {
				req.Header.Set("X-Internal-Token", cfg.InternalToken)
			} else {
				req.Header.Set("Authorization", "Bearer "+bearer)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code >= http.StatusInternalServerError {
				t.Fatalf("%s %s returned %d: %s", route.Method, path, recorder.Code, recorder.Body.String())
			}
			if recorder.Header().Get("X-Request-ID") != "day8-route-regression" {
				t.Fatalf("request ID not preserved: %q", recorder.Header().Get("X-Request-ID"))
			}
		})
	}
	if tested == 0 {
		t.Fatal("no public/internal routes were discovered")
	}

}

func regressionPath(pattern string, adminID, memberID uint) string {
	path := strings.ReplaceAll(pattern, ":orderNo", "missing-order")
	path = strings.ReplaceAll(path, ":creatorId", strconv.FormatUint(uint64(memberID), 10))
	path = strings.ReplaceAll(path, ":userId", strconv.FormatUint(uint64(memberID), 10))
	path = strings.ReplaceAll(path, ":id", strconv.FormatUint(uint64(memberID), 10))
	return path
}

func regressionToken(t *testing.T, secret string, user model.User) string {
	t.Helper()
	claims := middleware.Claims{UserID: user.ID, Username: user.Username, Role: user.Role, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func assertInternalBool(t *testing.T, router http.Handler, token, path, field string, want bool) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.Header.Set("X-Internal-Token", token)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != http.StatusOK {
		t.Fatalf("GET %s returned %d: %s", path, recorder.Code, recorder.Body.String())
	}
	var envelope struct {
		Data map[string]bool `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Data[field] != want {
		t.Fatalf("GET %s field %s = %v, want %v", path, field, envelope.Data[field], want)
	}
}

func userModels() []any {
	return []any{
		&model.User{}, &model.Follow{}, &model.FollowGroup{}, &model.UserBlock{},
		&model.CreatorMembershipPlan{}, &model.CreatorSubscription{}, &model.SubscriptionOrder{},
		&model.ChatMessage{}, &model.Notification{},
	}
}
