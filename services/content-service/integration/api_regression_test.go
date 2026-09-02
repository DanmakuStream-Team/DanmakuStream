//go:build integration

package integration

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"danmakustream/content-service/internal/config"
	"danmakustream/content-service/internal/logic"
	"danmakustream/content-service/internal/middleware"
	"danmakustream/content-service/internal/model"
	"danmakustream/content-service/internal/server"
	"danmakustream/content-service/internal/svc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	regressionJWTSecret = "content-regression-jwt-secret"
	regressionInternal  = "content-regression-internal-token"
)

var expectedRoutes = map[string]struct{}{
	"GET /api/v1/livez": {}, "GET /api/v1/health": {}, "GET /api/v1/version": {},
	"GET /api/v1/videos": {}, "GET /api/v1/videos/:id": {}, "GET /api/v1/users/:id/videos": {},
	"GET /api/v1/dynamics": {}, "GET /api/v1/banners": {}, "GET /api/v1/announcements": {},
	"GET /api/v1/users/me/videos": {}, "GET /api/v1/creator/analytics": {},
	"POST /api/v1/images/upload": {}, "POST /api/v1/videos/upload": {},
	"PUT /api/v1/videos/:id": {}, "POST /api/v1/videos/:id/cover": {},
	"GET /api/v1/videos/:id/download": {}, "DELETE /api/v1/videos/:id": {},
	"POST /api/v1/videos/:id/collaborators": {}, "DELETE /api/v1/videos/:id/collaborators/:userId": {},
	"POST /api/v1/dynamics": {}, "DELETE /api/v1/dynamics/:id": {},
	"GET /api/v1/admin/videos": {}, "PUT /api/v1/admin/videos/:id/status": {},
	"GET /internal/v1/videos/batch": {}, "GET /internal/v1/videos/:id": {},
	"GET /api/v1/admin/banners": {}, "POST /api/v1/admin/banners": {},
	"PUT /api/v1/admin/banners/:id": {}, "DELETE /api/v1/admin/banners/:id": {},
	"GET /api/v1/admin/announcements": {}, "POST /api/v1/admin/announcements": {},
	"PUT /api/v1/admin/announcements/:id": {}, "DELETE /api/v1/admin/announcements/:id": {},
}

type fixture struct {
	ownerID        uint
	videoID        uint
	pendingID      uint
	dynamicID      uint
	bannerID       uint
	announcementID uint
}

func TestAllPublicAndInternalAPIRoutes(t *testing.T) {
	dsn := os.Getenv("CONTENT_SERVICE_TEST_DSN")
	if dsn == "" {
		t.Skip("CONTENT_SERVICE_TEST_DSN is not configured")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(model.ContentModels()...); err != nil {
		t.Fatal(err)
	}
	tx := db.Begin()
	if tx.Error != nil {
		t.Fatal(tx.Error)
	}
	defer tx.Rollback()

	data := seedFixture(t, tx)
	cfg := config.Config{
		ServiceName: "content-service", ServiceVersion: "day8-regression", CommitSHA: "integration",
		BuildTime: time.Now().UTC().Format(time.RFC3339), Port: "8080", JWTSecret: regressionJWTSecret,
		InternalAPIToken: regressionInternal, RequestTimeout: time.Second, StorageDir: t.TempDir(),
		MaxVideoBytes: 1 << 20, MaxImageBytes: 1 << 20,
	}
	ctx := &svc.Context{Config: cfg, DB: tx, Logic: &logic.Service{DB: tx}}
	router := server.Router(ctx, slog.New(slog.NewTextHandler(io.Discard, nil)))
	assertRouteInventory(t, router)
	assertContentContracts(t, router, tx, data)

	bearer := regressionToken(t, data.ownerID, "admin")
	tested := 0
	for _, route := range router.Routes() {
		if !strings.HasPrefix(route.Path, "/api/v1/") && !strings.HasPrefix(route.Path, "/internal/v1/") {
			continue
		}
		tested++
		t.Run(route.Method+" "+route.Path, func(t *testing.T) {
			path := regressionPath(route.Path, data)
			body := regressionBody(route.Method, route.Path)
			req := httptest.NewRequest(route.Method, path, body)
			req.Header.Set("X-Request-ID", "content-day8-route-regression")
			if body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			if strings.HasPrefix(route.Path, "/internal/v1/") {
				req.Header.Set("X-Internal-Token", regressionInternal)
			} else {
				req.Header.Set("Authorization", "Bearer "+bearer)
			}
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)
			if recorder.Code >= http.StatusInternalServerError {
				t.Fatalf("%s %s returned %d: %s", route.Method, path, recorder.Code, recorder.Body.String())
			}
			if recorder.Header().Get("X-Request-ID") != "content-day8-route-regression" {
				t.Fatalf("request ID not preserved: %q", recorder.Header().Get("X-Request-ID"))
			}
		})
	}
	if tested != len(expectedRoutes) {
		t.Fatalf("tested %d routes, want %d", tested, len(expectedRoutes))
	}
}

func seedFixture(t *testing.T, db *gorm.DB) fixture {
	t.Helper()
	suffix := time.Now().UnixNano()
	ownerID := uint(suffix%1_000_000 + 100)
	approved := model.Video{Title: fmt.Sprintf("regression searchable %d", suffix), Description: "content API regression", VideoURL: "/media/videos/missing.mp4", AuthorID: ownerID, Status: "approved", TranscodeStatus: "ready"}
	pending := model.Video{Title: fmt.Sprintf("regression pending %d", suffix), VideoURL: "/media/videos/pending.mp4", AuthorID: ownerID, Status: "pending", TranscodeStatus: "ready"}
	dynamic := model.DynamicPost{UserID: ownerID, Content: "regression dynamic", Images: "[]"}
	banner := model.SiteBanner{Title: "regression banner", Enabled: true}
	announcement := model.SiteAnnouncement{Content: "regression announcement", Enabled: true}
	for _, value := range []any{&approved, &pending, &dynamic, &banner, &announcement} {
		if err := db.Create(value).Error; err != nil {
			t.Fatal(err)
		}
	}
	return fixture{ownerID: ownerID, videoID: approved.ID, pendingID: pending.ID, dynamicID: dynamic.ID, bannerID: banner.ID, announcementID: announcement.ID}
}

func assertRouteInventory(t *testing.T, router *gin.Engine) {
	t.Helper()
	actual := make(map[string]struct{})
	for _, route := range router.Routes() {
		if strings.HasPrefix(route.Path, "/api/v1/") || strings.HasPrefix(route.Path, "/internal/v1/") {
			actual[route.Method+" "+route.Path] = struct{}{}
		}
	}
	for route := range expectedRoutes {
		if _, ok := actual[route]; !ok {
			t.Errorf("registered API route is missing from regression: %s", route)
		}
	}
	for route := range actual {
		if _, ok := expectedRoutes[route]; !ok {
			t.Errorf("new API route has no regression mapping: %s", route)
		}
	}
}

func assertContentContracts(t *testing.T, router http.Handler, db *gorm.DB, data fixture) {
	t.Helper()
	list := serve(t, router, http.MethodGet, "/api/v1/videos?keyword=searchable", "", nil)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), strconv.FormatUint(uint64(data.videoID), 10)) {
		t.Fatalf("search contract failed: %d %s", list.Code, list.Body.String())
	}
	detail := serve(t, router, http.MethodGet, fmt.Sprintf("/api/v1/videos/%d", data.videoID), "", nil)
	if detail.Code != http.StatusOK {
		t.Fatalf("playable detail failed: %d %s", detail.Code, detail.Body.String())
	}
	var stored model.Video
	if err := db.First(&stored, data.videoID).Error; err != nil || stored.ViewCount != 1 {
		t.Fatalf("view statistic was not persisted: count=%d err=%v", stored.ViewCount, err)
	}
	internal := serve(t, router, http.MethodGet, fmt.Sprintf("/internal/v1/videos/%d", data.videoID), regressionInternal, nil)
	if internal.Code != http.StatusOK || !strings.Contains(internal.Body.String(), `"playable":true`) {
		t.Fatalf("internal playable contract failed: %d %s", internal.Code, internal.Body.String())
	}
	review := serve(t, router, http.MethodPut, fmt.Sprintf("/api/v1/admin/videos/%d/status", data.pendingID), regressionToken(t, data.ownerID, "moderator"), bytes.NewBufferString(`{"status":"approved"}`))
	if review.Code != http.StatusOK {
		t.Fatalf("review contract failed: %d %s", review.Code, review.Body.String())
	}
	analytics := serve(t, router, http.MethodGet, "/api/v1/creator/analytics", regressionToken(t, data.ownerID, "user"), nil)
	if analytics.Code != http.StatusOK || !strings.Contains(analytics.Body.String(), `"videoCount":2`) {
		t.Fatalf("analytics contract failed: %d %s", analytics.Code, analytics.Body.String())
	}
}

func serve(t *testing.T, router http.Handler, method, path, credential string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if strings.HasPrefix(path, "/internal/") {
		req.Header.Set("X-Internal-Token", credential)
	} else if credential != "" {
		req.Header.Set("Authorization", "Bearer "+credential)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func regressionPath(pattern string, data fixture) string {
	path := pattern
	switch {
	case strings.Contains(path, "/admin/banners/:id"):
		path = strings.ReplaceAll(path, ":id", strconv.FormatUint(uint64(data.bannerID), 10))
	case strings.Contains(path, "/admin/announcements/:id"):
		path = strings.ReplaceAll(path, ":id", strconv.FormatUint(uint64(data.announcementID), 10))
	case strings.Contains(path, "/dynamics/:id"):
		path = strings.ReplaceAll(path, ":id", strconv.FormatUint(uint64(data.dynamicID), 10))
	case strings.Contains(path, "/users/:id/videos"):
		path = strings.ReplaceAll(path, ":id", strconv.FormatUint(uint64(data.ownerID), 10))
	default:
		path = strings.ReplaceAll(path, ":id", strconv.FormatUint(uint64(data.videoID), 10))
	}
	path = strings.ReplaceAll(path, ":userId", strconv.FormatUint(uint64(data.ownerID+1), 10))
	if pattern == "/internal/v1/videos/batch" {
		path += fmt.Sprintf("?ids=%d,%d", data.videoID, data.pendingID)
	}
	return path
}

func regressionBody(method, path string) io.Reader {
	if method != http.MethodPost && method != http.MethodPut {
		return nil
	}
	switch {
	case path == "/api/v1/dynamics":
		return bytes.NewBufferString(`{"content":"route regression","images":[]}`)
	case strings.HasSuffix(path, "/collaborators"):
		return bytes.NewBufferString(`{"userId":999999}`)
	case strings.HasSuffix(path, "/status"):
		return bytes.NewBufferString(`{"status":"rejected"}`)
	case strings.Contains(path, "/admin/banners"):
		return bytes.NewBufferString(`{"title":"route regression","enabled":true}`)
	case strings.Contains(path, "/admin/announcements"):
		return bytes.NewBufferString(`{"content":"route regression","enabled":true}`)
	case method == http.MethodPut && strings.Contains(path, "/videos/:id"):
		return bytes.NewBufferString(`{"title":"route regression"}`)
	default:
		return bytes.NewBufferString(`{}`)
	}
}

func regressionToken(t *testing.T, userID uint, role string) string {
	t.Helper()
	claims := middleware.Claims{UserID: userID, Username: "content-regression", Role: role, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute))}}
	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(regressionJWTSecret))
	if err != nil {
		t.Fatal(err)
	}
	return value
}
