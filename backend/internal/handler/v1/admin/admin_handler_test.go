package admin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// UNIT-TC13-03 角色取值校验：user/moderator/admin 合法，其余非法
func TestIsValidRole(t *testing.T) {
	cases := []struct {
		role string
		want bool
	}{
		{role: "user", want: true},
		{role: "moderator", want: true},
		{role: "admin", want: true},
		{role: "", want: false},
		{role: "root", want: false},
		{role: "superadmin", want: false},
		{role: "ADMIN", want: false},
		{role: "admin ", want: false},
	}
	for _, tc := range cases {
		if got := isValidRole(tc.role); got != tc.want {
			t.Errorf("isValidRole(%q) = %v, want %v", tc.role, got, tc.want)
		}
	}
}

// UNIT-TC13-03 接口层：非法角色返回 400
func TestUpdateUserRoleHandlerInvalidRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := UpdateUserRoleHandler(nil)
	r := gin.New()
	r.PUT("/admin/users/:id/role", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/admin/users/1/role", bytes.NewBufferString(`{"role":"superadmin"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400, body = %s", w.Code, w.Body.String())
	}
}

// UNIT-TC13-04 横幅标题校验：空标题（含纯空白）返回 400
func TestCreateBannerHandlerEmptyTitle(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := CreateBannerHandler(nil)
	r := gin.New()
	r.POST("/admin/banners", handler)

	cases := []struct {
		name string
		body string
	}{
		{name: "missing_title", body: `{"imageUrl":"https://x/a.png"}`},
		{name: "empty_title", body: `{"title":"","imageUrl":"https://x/a.png"}`},
		{name: "whitespace_title", body: `{"title":"   ","imageUrl":"https://x/a.png"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/admin/banners", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want 400, body = %s", w.Code, w.Body.String())
			}
		})
	}
}

// UNIT-TC13-05 公告参数校验：空内容、非法时间报错；合法时间构造成功
func TestBuildAnnouncement(t *testing.T) {
	cases := []struct {
		name    string
		req     saveAnnouncementReq
		wantErr bool
	}{
		{name: "empty_content", req: saveAnnouncementReq{Content: ""}, wantErr: true},
		{name: "whitespace_content", req: saveAnnouncementReq{Content: "   "}, wantErr: true},
		{name: "invalid_started_at", req: saveAnnouncementReq{Content: "公告", StartedAt: "not-a-time"}, wantErr: true},
		{name: "invalid_ended_at", req: saveAnnouncementReq{Content: "公告", EndedAt: "2026/09/01"}, wantErr: true},
		{name: "valid_rfc3339", req: saveAnnouncementReq{Content: "公告", StartedAt: "2026-08-26T00:00:00Z", EndedAt: "2026-09-26T00:00:00Z"}, wantErr: false},
		{name: "valid_layout", req: saveAnnouncementReq{Content: "公告", StartedAt: "2026-08-26 00:00:00"}, wantErr: false},
		{name: "no_time_bounds", req: saveAnnouncementReq{Content: "公告"}, wantErr: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			item, err := buildAnnouncement(tc.req)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("buildAnnouncement(%+v) expected error, got nil (item=%+v)", tc.req, item)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildAnnouncement(%+v) unexpected error: %v", tc.req, err)
			}
			if item.Content == "" {
				t.Errorf("content should be trimmed but non-empty, got %q", item.Content)
			}
		})
	}
}
