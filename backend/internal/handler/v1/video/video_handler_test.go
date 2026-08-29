package video

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"danmakustream/backend/internal/middleware"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
)

// UNIT-TC13-06 审核状态校验：pending/approved/rejected 合法，其余非法
func TestIsValidVideoStatus(t *testing.T) {
	cases := []struct {
		status string
		want   bool
	}{
		{status: "pending", want: true},
		{status: "approved", want: true},
		{status: "rejected", want: true},
		{status: "", want: false},
		{status: "publish", want: false},
		{status: "Approved", want: false},
		{status: "approved ", want: false},
		{status: "deleted", want: false},
	}
	for _, tc := range cases {
		if got := isValidVideoStatus(tc.status); got != tc.want {
			t.Errorf("isValidVideoStatus(%q) = %v, want %v", tc.status, got, tc.want)
		}
	}
}

func TestIsValidReviewDecision(t *testing.T) {
	for _, status := range []string{"approved", "rejected"} {
		if !isValidReviewDecision(status) {
			t.Errorf("isValidReviewDecision(%q) = false, want true", status)
		}
	}
	for _, status := range []string{"", "pending", "published", "failed"} {
		if isValidReviewDecision(status) {
			t.Errorf("isValidReviewDecision(%q) = true, want false", status)
		}
	}
}

// UNIT-TC02-02 详情参数：非数字视频 ID 在访问数据库前返回 400。
func TestDetailHandlerRejectsInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/videos/:id", DetailHandler(nil))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/videos/not-a-number", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body = %s", w.Code, w.Body.String())
	}
}

// UNIT-TC03-01 投稿参数：缺标题或缺视频文件时不创建任何记录。
func TestUploadHandlerRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name        string
		contentType string
		body        string
	}{
		{name: "missing_title", contentType: "application/x-www-form-urlencoded", body: "description=x"},
		{name: "missing_video", contentType: "application/x-www-form-urlencoded", body: "title=demo"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.Use(func(c *gin.Context) { c.Set(middleware.CtxKeyUserID, uint(7)) })
			r.POST("/videos/upload", UploadHandler(nil))
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/videos/upload", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", tc.contentType)
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400, body = %s", w.Code, w.Body.String())
			}
		})
	}
}

// UNIT-TC03-02 下载文件名：清理路径和操作系统保留字符。
func TestSafeDownloadName(t *testing.T) {
	cases := map[string]string{
		"  示例视频  ":    "示例视频",
		"../a:b?.mp4": ".._a_b_.mp4",
		"":            "danmaku-video",
		"   ":         "danmaku-video",
	}
	for input, want := range cases {
		if got := safeDownloadName(input); got != want {
			t.Errorf("safeDownloadName(%q) = %q, want %q", input, got, want)
		}
	}
}

// UNIT-TC03-03 媒体路径：仅接受站内 /media/ 相对路径。
func TestMediaPathToLocalPath(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "videos", "12", "playlist.m3u8")
	for _, input := range []string{"/media/videos/12/playlist.m3u8", "media/videos/12/playlist.m3u8"} {
		got, err := mediaPathToLocalPath(root, input)
		if err != nil || got != want {
			t.Errorf("mediaPathToLocalPath(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{"https://example.com/a.mp4", "/data/videos/a.mp4", "videos/a.mp4"} {
		if got, err := mediaPathToLocalPath(root, input); err == nil {
			t.Errorf("mediaPathToLocalPath(%q) = %q, want error", input, got)
		}
	}
}

// UNIT-TC04-01 审核源文件判断：只有 videos/<id>/upload.* 存在时视为已上传。
func TestHasUploadedVideoSource(t *testing.T) {
	root := t.TempDir()
	ctx := &svc.ServiceContext{VideoDir: root}
	if hasUploadedVideoSource(ctx, 42) {
		t.Fatal("empty video directory must not be treated as uploaded")
	}
	dir := filepath.Join(root, "videos", "42")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "playlist.m3u8"), []byte("#EXTM3U"), 0o644); err != nil {
		t.Fatal(err)
	}
	if hasUploadedVideoSource(ctx, 42) {
		t.Fatal("transcoded playlist alone must not be treated as upload source")
	}
	if err := os.WriteFile(filepath.Join(dir, "upload.mp4"), []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !hasUploadedVideoSource(ctx, 42) {
		t.Fatal("upload.mp4 should be detected")
	}
}

// UNIT-TC04-02 审核请求：非法 ID、JSON 和状态均在数据库访问前返回 400。
func TestAdminUpdateStatusHandlerRejectsInvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name string
		path string
		body string
	}{
		{name: "invalid_id", path: "/admin/videos/nope/status", body: `{"status":"approved"}`},
		{name: "malformed_json", path: "/admin/videos/1/status", body: `{"status":`},
		{name: "invalid_status", path: "/admin/videos/1/status", body: `{"status":"published"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			r.PUT("/admin/videos/:id/status", AdminUpdateStatusHandler(nil))
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, tc.path, bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400, body = %s", w.Code, w.Body.String())
			}
		})
	}
}
