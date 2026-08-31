package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// UNIT-TC13-01 AdminMiddleware：仅 admin 放行，user/moderator/缺失角色返回 403
func TestAdminMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name     string
		role     string
		setRole  bool
		wantCode int
		wantNext bool
	}{
		{name: "role_user_returns_403", role: "user", setRole: true, wantCode: http.StatusForbidden, wantNext: false},
		{name: "role_moderator_returns_403", role: "moderator", setRole: true, wantCode: http.StatusForbidden, wantNext: false},
		{name: "role_admin_passes", role: "admin", setRole: true, wantCode: http.StatusOK, wantNext: true},
		{name: "missing_role_returns_403", role: "", setRole: false, wantCode: http.StatusForbidden, wantNext: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			r := gin.New()
			r.Use(func(c *gin.Context) {
				if tc.setRole {
					c.Set(CtxKeyRole, tc.role)
				}
			})
			r.Use(AdminMiddleware)
			r.GET("/admin/whatever", func(c *gin.Context) {
				nextCalled = true
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/admin/whatever", nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.wantCode {
				t.Errorf("status = %d, want %d, body = %s", w.Code, tc.wantCode, w.Body.String())
			}
			if nextCalled != tc.wantNext {
				t.Errorf("handler reached = %v, want %v", nextCalled, tc.wantNext)
			}
		})
	}
}

// UNIT-TC13-02 StaffMiddleware：user 返回 403，moderator/admin 放行
func TestStaffMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cases := []struct {
		name     string
		role     string
		wantCode int
		wantNext bool
	}{
		{name: "role_user_returns_403", role: "user", wantCode: http.StatusForbidden, wantNext: false},
		{name: "role_moderator_passes", role: "moderator", wantCode: http.StatusOK, wantNext: true},
		{name: "role_admin_passes", role: "admin", wantCode: http.StatusOK, wantNext: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			nextCalled := false
			r := gin.New()
			r.Use(func(c *gin.Context) { c.Set(CtxKeyRole, tc.role) })
			r.Use(StaffMiddleware)
			r.GET("/admin/videos", func(c *gin.Context) {
				nextCalled = true
				c.Status(http.StatusOK)
			})

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/admin/videos", nil)
			r.ServeHTTP(w, req)

			if w.Code != tc.wantCode {
				t.Errorf("status = %d, want %d, body = %s", w.Code, tc.wantCode, w.Body.String())
			}
			if nextCalled != tc.wantNext {
				t.Errorf("handler reached = %v, want %v", nextCalled, tc.wantNext)
			}
		})
	}
}
