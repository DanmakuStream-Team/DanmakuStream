package message

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"danmakustream/user-service/internal/client"
	chatlogic "danmakustream/user-service/internal/logic/chat"
	"danmakustream/user-service/internal/middleware"
	"danmakustream/user-service/internal/svc"

	"github.com/gin-gonic/gin"
)

func TestSendHandlerValidation(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "malformed json", body: "{"},
		{name: "missing receiver", body: `{"type":"text","content":"hello"}`},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			status := serveMessageHandler(t, http.MethodPost, "/messages", test.body, "/messages", SendHandler(&svc.ServiceContext{}))
			if status != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
			}
		})
	}
}

func TestHistoryHandlerRejectsInvalidUserID(t *testing.T) {
	status := serveMessageHandler(t, http.MethodGet, "/messages/nope", "", "/messages/:userId", HistoryHandler(&svc.ServiceContext{}))
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", status, http.StatusBadRequest)
	}
}

func TestGetPageNormalizesBounds(t *testing.T) {
	tests := []struct {
		query         string
		page, perPage int
	}{
		{query: "", page: 1, perPage: 50},
		{query: "?page=0&pageSize=0", page: 1, perPage: 50},
		{query: "?page=3&pageSize=100", page: 3, perPage: 100},
		{query: "?page=2&pageSize=101", page: 2, perPage: 50},
	}
	for _, test := range tests {
		context, _ := gin.CreateTestContext(httptest.NewRecorder())
		context.Request = httptest.NewRequest(http.MethodGet, "/messages"+test.query, nil)
		page, pageSize := getPage(context)
		if page != test.page || pageSize != test.perPage {
			t.Fatalf("query %q => (%d,%d), want (%d,%d)", test.query, page, pageSize, test.page, test.perPage)
		}
	}
}

func TestWriteSendErrorStatus(t *testing.T) {
	tests := []struct {
		err    error
		status int
	}{
		{err: chatlogic.ErrBlocked, status: http.StatusForbidden},
		{err: chatlogic.ErrUserNotFound, status: http.StatusNotFound},
		{err: chatlogic.ErrVideoMissing, status: http.StatusNotFound},
		{err: client.ErrBadGateway, status: http.StatusBadGateway},
		{err: client.ErrUnavailable, status: http.StatusServiceUnavailable},
		{err: client.ErrTimeout, status: http.StatusGatewayTimeout},
		{err: chatlogic.ErrEmptyContent, status: http.StatusBadRequest},
		{err: errors.New("database failed"), status: http.StatusBadRequest},
	}
	for _, test := range tests {
		recorder := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(recorder)
		writeSendError(context, test.err)
		if recorder.Code != test.status {
			t.Fatalf("error %v status = %d, want %d", test.err, recorder.Code, test.status)
		}
	}
}

func serveMessageHandler(t *testing.T, method, path, body, route string, handler gin.HandlerFunc) int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(middleware.CtxKeyUserID, uint(7))
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
