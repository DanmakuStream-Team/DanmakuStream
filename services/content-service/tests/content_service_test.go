package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"danmakustream/content-service/internal/config"
	"danmakustream/content-service/internal/logic"
	"danmakustream/content-service/internal/model"
	"danmakustream/content-service/internal/svc"
	"github.com/glebarez/sqlite"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

const testSecret = "content-test-secret"

func testService(t *testing.T) (*svc.Context, http.Handler) {
	t.Helper()
	dsn := fmt.Sprintf("file:content-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(model.ContentModels()...); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{
		ServiceName: "content-service", ServiceVersion: "microservice-test", CommitSHA: "abcdef0",
		BuildTime: "2026-08-31T10:00:00+08:00", Port: "8080", JWTSecret: testSecret,
		InternalAPIToken: "internal-test-token", RequestTimeout: 100 * time.Millisecond,
		StorageDir: t.TempDir(), MaxVideoBytes: 1 << 20, MaxImageBytes: 1 << 20,
	}
	ctx := &svc.Context{Config: cfg, DB: db, Logic: &logic.Service{DB: db}}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return ctx, svc.Router(ctx, logger)
}

func token(t *testing.T, userID uint, role string) string {
	t.Helper()
	claims := struct {
		UserID uint   `json:"userId"`
		Role   string `json:"role"`
		jwt.RegisteredClaims
	}{UserID: userID, Role: role, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour))}}
	value, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(testSecret))
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func request(t *testing.T, router http.Handler, method, path, bearer string, body io.Reader, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, body)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	return recorder
}

func responseData(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response %q: %v", recorder.Body.String(), err)
	}
	return envelope.Data
}

func TestHealthAndVersionContracts(t *testing.T) {
	ctx, router := testService(t)
	for _, path := range []string{"/api/v1/livez", "/api/v1/health", "/api/v1/version"} {
		res := request(t, router, http.MethodGet, path, "", nil, "")
		if res.Code != http.StatusOK {
			t.Fatalf("%s status = %d, body=%s", path, res.Code, res.Body.String())
		}
		if !strings.Contains(res.Body.String(), `"code":0`) || !strings.Contains(res.Body.String(), `"message":"ok"`) {
			t.Fatalf("%s violates success envelope: %s", path, res.Body.String())
		}
	}
	version := request(t, router, http.MethodGet, "/api/v1/version", "", nil, "")
	data := responseData(t, version)
	for _, field := range []string{"service", "version", "commit", "buildTime"} {
		if data[field] == "" || data[field] == nil {
			t.Fatalf("version field %s is empty", field)
		}
	}
	sqlDB, err := ctx.DB.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatal(err)
	}
	down := request(t, router, http.MethodGet, "/api/v1/health", "", nil, "")
	if down.Code != http.StatusServiceUnavailable {
		t.Fatalf("closed DB health = %d, body=%s", down.Code, down.Body.String())
	}
	if !strings.Contains(down.Body.String(), `"requestId"`) {
		t.Fatalf("health error has no requestId: %s", down.Body.String())
	}
}

func TestAuthenticationAndRoles(t *testing.T) {
	_, router := testService(t)
	unauthorized := request(t, router, http.MethodPost, "/api/v1/dynamics", "", strings.NewReader(`{"content":"hello"}`), "application/json")
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorized.Code)
	}
	if !strings.Contains(unauthorized.Body.String(), `"code":40100`) || !strings.Contains(unauthorized.Body.String(), `"requestId"`) {
		t.Fatalf("invalid auth error: %s", unauthorized.Body.String())
	}
	moderator := request(t, router, http.MethodPost, "/api/v1/admin/banners", token(t, 2, "moderator"), strings.NewReader(`{"title":"x"}`), "application/json")
	if moderator.Code != http.StatusForbidden {
		t.Fatalf("moderator admin status = %d, body=%s", moderator.Code, moderator.Body.String())
	}
	admin := request(t, router, http.MethodPost, "/api/v1/admin/banners", token(t, 3, "admin"), strings.NewReader(`{"title":"notice","enabled":true}`), "application/json")
	if admin.Code != http.StatusCreated {
		t.Fatalf("admin create status = %d, body=%s", admin.Code, admin.Body.String())
	}
}

func TestUploadReviewAndOwnershipFlow(t *testing.T) {
	ctx, router := testService(t)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("title", "D06 content video")
	_ = writer.WriteField("category", "technology")
	part, err := writer.CreateFormFile("video", "sample.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(append([]byte{0, 0, 0, 24}, []byte("ftypmp42local test video bytes")...)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	upload := request(t, router, http.MethodPost, "/api/v1/videos/upload", token(t, 1, "user"), body, writer.FormDataContentType())
	if upload.Code != http.StatusCreated {
		t.Fatalf("upload status = %d, body=%s", upload.Code, upload.Body.String())
	}
	data := responseData(t, upload)
	videoID := uint(data["id"].(float64))
	if data["status"] != "pending" {
		t.Fatalf("upload status field = %v", data["status"])
	}
	videoURL, _ := data["videoUrl"].(string)
	mediaPath := filepath.Join(ctx.Config.StorageDir, strings.TrimPrefix(videoURL, "/media/"))
	if _, err := os.Stat(mediaPath); err != nil {
		t.Fatalf("uploaded media missing: %v", err)
	}

	publicBeforeReview := request(t, router, http.MethodGet, "/api/v1/videos/"+strconv.Itoa(int(videoID)), "", nil, "")
	if publicBeforeReview.Code != http.StatusNotFound {
		t.Fatalf("pending video public status = %d", publicBeforeReview.Code)
	}
	forbiddenEdit := request(t, router, http.MethodPut, "/api/v1/videos/"+strconv.Itoa(int(videoID)), token(t, 2, "user"), strings.NewReader(`{"title":"stolen"}`), "application/json")
	if forbiddenEdit.Code != http.StatusForbidden {
		t.Fatalf("non-owner edit status = %d, body=%s", forbiddenEdit.Code, forbiddenEdit.Body.String())
	}
	review := request(t, router, http.MethodPut, "/api/v1/admin/videos/"+strconv.Itoa(int(videoID))+"/status", token(t, 9, "moderator"), strings.NewReader(`{"status":"approved"}`), "application/json")
	if review.Code != http.StatusOK {
		t.Fatalf("review status = %d, body=%s", review.Code, review.Body.String())
	}
	repeated := request(t, router, http.MethodPut, "/api/v1/admin/videos/"+strconv.Itoa(int(videoID))+"/status", token(t, 9, "moderator"), strings.NewReader(`{"status":"rejected"}`), "application/json")
	if repeated.Code != http.StatusConflict {
		t.Fatalf("repeat review status = %d, body=%s", repeated.Code, repeated.Body.String())
	}
	publicAfterReview := request(t, router, http.MethodGet, "/api/v1/videos/"+strconv.Itoa(int(videoID)), "", nil, "")
	if publicAfterReview.Code != http.StatusOK {
		t.Fatalf("approved video status = %d, body=%s", publicAfterReview.Code, publicAfterReview.Body.String())
	}
	publicData := responseData(t, publicAfterReview)
	author := publicData["author"].(map[string]any)
	if author["id"] != float64(1) || author["nickname"] != "" {
		t.Fatalf("unexpected safe author fallback: %#v", author)
	}
}

func TestUploadRejectsInvalidAndOversizedVideo(t *testing.T) {
	_, router := testService(t)
	makeUpload := func(name string, content []byte) (*bytes.Buffer, string) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("title", "invalid upload")
		part, err := writer.CreateFormFile("video", name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := part.Write(content); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		return body, writer.FormDataContentType()
	}
	invalidBody, invalidType := makeUpload("fake.mp4", []byte("this is not a video"))
	invalid := request(t, router, http.MethodPost, "/api/v1/videos/upload", token(t, 1, "user"), invalidBody, invalidType)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid media status = %d, body=%s", invalid.Code, invalid.Body.String())
	}
	large := append([]byte{0, 0, 0, 24}, []byte("ftypmp42")...)
	large = append(large, bytes.Repeat([]byte{0}, (1<<20)+1)...)
	largeBody, largeType := makeUpload("large.mp4", large)
	oversized := request(t, router, http.MethodPost, "/api/v1/videos/upload", token(t, 1, "user"), largeBody, largeType)
	if oversized.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized media status = %d, body=%s", oversized.Code, oversized.Body.String())
	}
}

func TestContentSchemaDoesNotOwnOtherDomains(t *testing.T) {
	ctx, _ := testService(t)
	for _, table := range []string{"videos", "media_assets", "video_collaborators", "dynamic_posts", "site_banners", "site_announcements", "creator_daily_stats", "video_daily_stats"} {
		if !ctx.DB.Migrator().HasTable(table) {
			t.Errorf("content table %s missing", table)
		}
	}
	for _, table := range []string{"users", "comments", "danmakus", "video_likes", "video_collections", "live_rooms"} {
		if ctx.DB.Migrator().HasTable(table) {
			t.Errorf("foreign-domain table %s must not exist", table)
		}
	}
	var foreignKeys []struct {
		Table string `gorm:"column:table"`
	}
	_ = ctx.DB.Raw("SELECT name AS `table` FROM sqlite_master WHERE type='table' AND sql LIKE '%FOREIGN KEY%'").Scan(&foreignKeys).Error
	if len(foreignKeys) != 0 {
		t.Fatalf("cross-table foreign keys found: %#v", foreignKeys)
	}
}

func TestNotFoundUsesBusinessError(t *testing.T) {
	_, router := testService(t)
	res := request(t, router, http.MethodGet, "/api/v1/videos/999", "", nil, "")
	if res.Code != http.StatusNotFound || !strings.Contains(res.Body.String(), `"code":40401`) {
		t.Fatalf("not found response = %d %s", res.Code, res.Body.String())
	}
	unknown := request(t, router, http.MethodGet, "/api/v1/unknown", "", nil, "")
	if unknown.Code != http.StatusNotFound || !strings.Contains(unknown.Body.String(), `"code":40400`) {
		t.Fatalf("unknown route response = %d %s", unknown.Code, unknown.Body.String())
	}
}

func TestPublicPaginationShape(t *testing.T) {
	ctx, router := testService(t)
	for i := 0; i < 3; i++ {
		if err := ctx.DB.Create(&model.Video{Title: fmt.Sprintf("v%d", i), VideoURL: fmt.Sprintf("/media/videos/%d.mp4", i), AuthorID: 1, Status: "approved", TranscodeStatus: "ready"}).Error; err != nil {
			t.Fatal(err)
		}
	}
	res := request(t, router, http.MethodGet, "/api/v1/videos?"+url.Values{"page": {"1"}, "pageSize": {"2"}}.Encode(), "", nil, "")
	if res.Code != http.StatusOK {
		t.Fatalf("list status = %d, body=%s", res.Code, res.Body.String())
	}
	data := responseData(t, res)
	if data["total"] != float64(3) || len(data["items"].([]any)) != 2 {
		t.Fatalf("invalid page: %#v", data)
	}
}
