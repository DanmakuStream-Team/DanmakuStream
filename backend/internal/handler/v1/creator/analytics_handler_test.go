package creator

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"danmakustream/backend/internal/middleware"

	"github.com/gin-gonic/gin"
)

// UNIT-TC12-01 创作者分析鉴权与查询参数在数据库访问前完成校验。
func TestAnalyticsHandlerRejectsInvalidRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cases := []struct {
		name      string
		path      string
		creatorID uint
		want      int
	}{
		{name: "not_logged_in", path: "/creator/analytics", want: http.StatusUnauthorized},
		{name: "unsupported_days", path: "/creator/analytics?days=14", creatorID: 9, want: http.StatusBadRequest},
		{name: "non_numeric_days", path: "/creator/analytics?days=week", creatorID: 9, want: http.StatusBadRequest},
		{name: "zero_video_id", path: "/creator/analytics?videoId=0", creatorID: 9, want: http.StatusBadRequest},
		{name: "invalid_video_id", path: "/creator/analytics?videoId=abc", creatorID: 9, want: http.StatusBadRequest},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := gin.New()
			if tc.creatorID != 0 {
				r.Use(func(c *gin.Context) { c.Set(middleware.CtxKeyUserID, tc.creatorID) })
			}
			r.GET("/creator/analytics", AnalyticsHandler(nil))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if w.Code != tc.want {
				t.Fatalf("status = %d, want %d, body = %s", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

// UNIT-TC12-02 趋势日期以调用者所在时区的自然日零点为边界。
func TestBeginningOfDay(t *testing.T) {
	location := time.FixedZone("UTC+8", 8*60*60)
	value := time.Date(2026, time.August, 27, 23, 59, 58, 123, location)
	got := beginningOfDay(value)
	want := time.Date(2026, time.August, 27, 0, 0, 0, 0, location)
	if !got.Equal(want) {
		t.Fatalf("beginningOfDay(%v) = %v, want %v", value, got, want)
	}
	if got.Location() != location {
		t.Fatalf("location = %v, want %v", got.Location(), location)
	}
}
