package user

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"danmakustream/user-service/internal/middleware"
	"danmakustream/user-service/internal/svc"

	"github.com/gin-gonic/gin"
)

func TestCreateFollowGroupHandlerValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "invalid json", body: "{"},
		{name: "blank name", body: `{"name":"   "}`},
		{name: "name too long", body: `{"name":"` + strings.Repeat("组", 21) + `"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := serveUserHandler(t, http.MethodPost, "/groups", test.body, 7, "", CreateFollowGroupHandler(&svc.ServiceContext{}))
			if status != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
			}
		})
	}
}

func TestUpdateFollowSettingsHandlerValidation(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		body   string
		status int
	}{
		{name: "invalid target id", path: "/users/nope/settings", body: `{}`, status: http.StatusBadRequest},
		{name: "no changes", path: "/users/9/settings", body: `{}`, status: http.StatusBadRequest},
		{name: "paid special cannot be edited", path: "/users/9/settings", body: `{"special":true}`, status: http.StatusForbidden},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := serveUserHandler(t, http.MethodPut, test.path, test.body, 7, "/users/:id/settings", UpdateFollowSettingsHandler(&svc.ServiceContext{}))
			if status != test.status {
				t.Fatalf("status = %d, want %d", status, test.status)
			}
		})
	}
}

func TestRelationshipHandlersRejectSelfOperations(t *testing.T) {
	tests := []struct {
		name    string
		handler gin.HandlerFunc
	}{
		{name: "follow self", handler: FollowHandler(&svc.ServiceContext{})},
		{name: "block self", handler: BlockHandler(&svc.ServiceContext{})},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := serveUserHandler(t, http.MethodPost, "/users/7/action", "", 7, "/users/:id/action", test.handler)
			if status != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
			}
		})
	}
}

func serveUserHandler(t *testing.T, method, path, body string, userID uint, route string, handler gin.HandlerFunc) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	if route == "" {
		route = path
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.CtxKeyUserID, userID)
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
