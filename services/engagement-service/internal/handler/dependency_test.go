package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"danmakustream/engagement-service/internal/client"
	"github.com/gin-gonic/gin"
)

func TestDependencyErrorMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		name    string
		err     error
		status  int
		message string
	}{
		{"not found", client.ErrNotFound, http.StatusNotFound, "视频不存在"},
		{"timeout", client.ErrTimeout, http.StatusGatewayTimeout, "依赖服务调用超时"},
		{"invalid response", client.ErrBadGateway, http.StatusBadGateway, "依赖服务响应无效"},
		{"unavailable", client.ErrUnavailable, http.StatusServiceUnavailable, "依赖服务暂不可用"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Header("X-Request-ID", "req-dependency")
			dependencyError(c, errors.Join(errors.New("context"), tc.err), "视频")
			if w.Code != tc.status || !strings.Contains(w.Body.String(), tc.message) || !strings.Contains(w.Body.String(), `"requestId":"req-dependency"`) {
				t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
			}
		})
	}
}
