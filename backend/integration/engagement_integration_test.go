//go:build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"danmakustream/backend/internal/config"
	commenthandler "danmakustream/backend/internal/handler/v1/comment"
	danmakuhandler "danmakustream/backend/internal/handler/v1/danmaku"
	livehandler "danmakustream/backend/internal/handler/v1/live"
	videohandler "danmakustream/backend/internal/handler/v1/video"
	"danmakustream/backend/internal/middleware"
	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"

	"github.com/gin-gonic/gin"
	mysqldriver "github.com/go-sql-driver/mysql"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type responseEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func TestEngagementUseCasesWithMySQL(t *testing.T) {
	db := openTemporaryMySQL(t)
	testConfig := config.Config{}
	testConfig.Auth.AccessSecret = "integration-test-secret"
	testConfig.Live.RTMPHost = "localhost:1935"
	testConfig.Live.HTTPHost = "localhost:8081"
	svcCtx := &svc.ServiceContext{
		DB:     db,
		Config: testConfig,
	}

	creator := model.User{Username: "d_creator", Nickname: "D Creator", Password: "test", Role: "creator"}
	viewer := model.User{Username: "d_viewer", Nickname: "D Viewer", Password: "test", Role: "user"}
	if err := db.Create(&creator).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&viewer).Error; err != nil {
		t.Fatal(err)
	}

	router := newEngagementRouter(svcCtx)

	t.Run("UC05 video interaction persists and updates counts", func(t *testing.T) {
		video := model.Video{Title: "UC05", Status: "approved", AuthorID: creator.ID}
		if err := db.Create(&video).Error; err != nil {
			t.Fatal(err)
		}

		assertHTTPCode(t, router, http.MethodPost, "/danmaku", viewer.ID,
			map[string]any{"videoId": video.ID, "content": "hello", "time": 3}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/comments", viewer.ID,
			map[string]any{"videoId": video.ID, "content": "useful"}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/like", video.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/collect", video.ID), viewer.ID, nil, http.StatusOK)

		var refreshed model.Video
		if err := db.First(&refreshed, video.ID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.DanmakuCount != 1 || refreshed.LikeCount != 1 || refreshed.CollectCount != 1 {
			t.Fatalf("video counts = danmaku:%d like:%d collect:%d, want 1/1/1",
				refreshed.DanmakuCount, refreshed.LikeCount, refreshed.CollectCount)
		}
		assertModelCount(t, db, &model.Danmaku{}, "video_id = ? AND scene = ?", 1, video.ID, "video")
		assertModelCount(t, db, &model.Comment{}, "video_id = ?", 1, video.ID)

		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/like", video.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/collect", video.ID), viewer.ID, nil, http.StatusOK)
		if err := db.First(&refreshed, video.ID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.LikeCount != 0 || refreshed.CollectCount != 0 {
			t.Fatalf("counts after cancel = like:%d collect:%d, want 0/0", refreshed.LikeCount, refreshed.CollectCount)
		}
		assertHTTPCode(t, router, http.MethodPost, "/danmaku", viewer.ID,
			map[string]any{"videoId": video.ID, "content": ""}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/comments", viewer.ID,
			map[string]any{"videoId": video.ID, "content": "   "}, http.StatusBadRequest)
	})

	t.Run("UC09 rejects conflicting schedule and toggles reservation", func(t *testing.T) {
		scheduledAt := time.Now().Add(24 * time.Hour).Truncate(time.Second).Format(time.RFC3339)
		first := requestJSON(t, router, http.MethodPost, "/live-schedules", creator.ID,
			map[string]any{"title": "UC09", "scheduledAt": scheduledAt})
		if first.Code != 0 {
			t.Fatalf("first schedule failed: code=%d message=%q", first.Code, first.Message)
		}
		var created struct {
			ID uint `json:"id"`
		}
		if err := json.Unmarshal(first.Data, &created); err != nil || created.ID == 0 {
			t.Fatalf("decode created schedule: id=%d err=%v", created.ID, err)
		}

		assertHTTPCode(t, router, http.MethodPost, "/live-schedules", creator.ID,
			map[string]any{"title": "conflict", "scheduledAt": scheduledAt}, http.StatusConflict)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live-schedules/%d/reserve", created.ID), viewer.ID, nil, http.StatusOK)

		var schedule model.LiveSchedule
		if err := db.First(&schedule, created.ID).Error; err != nil {
			t.Fatal(err)
		}
		if schedule.ReminderCount != 1 {
			t.Fatalf("reminder_count = %d, want 1", schedule.ReminderCount)
		}
		assertModelCount(t, db, &model.LiveReservation{}, "schedule_id = ? AND user_id = ?", 1, created.ID, viewer.ID)

		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live-schedules/%d/reserve", created.ID), viewer.ID, nil, http.StatusOK)
		if err := db.First(&schedule, created.ID).Error; err != nil {
			t.Fatal(err)
		}
		if schedule.ReminderCount != 0 {
			t.Fatalf("reminder_count after cancel = %d, want 0", schedule.ReminderCount)
		}
		assertModelCount(t, db, &model.LiveReservation{}, "schedule_id = ? AND user_id = ?", 0, created.ID, viewer.ID)
	})

	t.Run("UC10 like and Super Chat update live interaction", func(t *testing.T) {
		now := time.Now()
		room := model.LiveRoom{Title: "UC10", StreamKey: "integration-room", Status: "live", OwnerID: creator.ID, StartedAt: &now}
		if err := db.Create(&room).Error; err != nil {
			t.Fatal(err)
		}

		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/like", room.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/gifts", room.ID), viewer.ID,
			map[string]any{"giftKey": "star", "count": 2, "message": "支持主播"}, http.StatusOK)

		var refreshed model.LiveRoom
		if err := db.First(&refreshed, room.ID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.LikeCount != 1 || refreshed.GiftValue != 100 {
			t.Fatalf("live counts = like:%d gift:%d, want 1/100", refreshed.LikeCount, refreshed.GiftValue)
		}
		var gift model.LiveGift
		if err := db.Where("room_id = ? AND user_id = ?", room.ID, viewer.ID).First(&gift).Error; err != nil {
			t.Fatal(err)
		}
		if gift.DisplaySeconds != 30 || gift.Message != "支持主播" {
			t.Fatalf("Super Chat = display:%d message:%q, want 30/支持主播", gift.DisplaySeconds, gift.Message)
		}

		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/like", room.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/gifts", room.ID), viewer.ID,
			map[string]any{"giftKey": "star", "count": 100}, http.StatusBadRequest)
		if err := db.First(&refreshed, room.ID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.LikeCount != 0 || refreshed.GiftValue != 100 {
			t.Fatalf("live counts after unlike/invalid gift = like:%d gift:%d, want 0/100", refreshed.LikeCount, refreshed.GiftValue)
		}
	})
}

func newEngagementRouter(svcCtx *svc.ServiceContext) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if raw := c.GetHeader("X-Test-User-ID"); raw != "" {
			id, _ := strconv.ParseUint(raw, 10, 64)
			c.Set(middleware.CtxKeyUserID, uint(id))
		}
		c.Next()
	})
	router.POST("/danmaku", danmakuhandler.SendHandler(svcCtx))
	router.POST("/comments", commenthandler.CreateHandler(svcCtx))
	router.POST("/videos/:id/like", videohandler.LikeHandler(svcCtx))
	router.POST("/videos/:id/collect", videohandler.CollectHandler(svcCtx))
	router.POST("/live-schedules", livehandler.CreateScheduleHandler(svcCtx))
	router.POST("/live-schedules/:id/reserve", livehandler.ReserveScheduleHandler(svcCtx))
	router.POST("/live/:id/like", livehandler.LikeHandler(svcCtx))
	router.POST("/live/:id/gifts", livehandler.GiftHandler(svcCtx))
	return router
}

func requestJSON(t *testing.T, router http.Handler, method, path string, userID uint, body any) responseEnvelope {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(userID), 10))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	var envelope responseEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response for %s %s: status=%d body=%q err=%v", method, path, recorder.Code, recorder.Body.String(), err)
	}
	return envelope
}

func assertHTTPCode(t *testing.T, router http.Handler, method, path string, userID uint, body any, want int) {
	t.Helper()
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", strconv.FormatUint(uint64(userID), 10))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != want {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, path, recorder.Code, want, recorder.Body.String())
	}
}

func assertModelCount(t *testing.T, db *gorm.DB, modelValue any, query string, want int64, args ...any) {
	t.Helper()
	var count int64
	if err := db.Model(modelValue).Where(query, args...).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != want {
		t.Fatalf("model count = %d, want %d", count, want)
	}
}

func openTemporaryMySQL(t *testing.T) *gorm.DB {
	t.Helper()
	adminDSN := os.Getenv("DANMAKU_TEST_ADMIN_DSN")
	if adminDSN == "" {
		t.Skip("set DANMAKU_TEST_ADMIN_DSN to run MySQL integration tests")
	}
	adminConfig, err := mysqldriver.ParseDSN(adminDSN)
	if err != nil {
		t.Fatal(err)
	}
	adminConfig.DBName = ""
	adminDB, err := sql.Open("mysql", adminConfig.FormatDSN())
	if err != nil {
		t.Fatal(err)
	}
	databaseName := fmt.Sprintf("danmakustream_d_sxh_%d", time.Now().UnixNano())
	if _, err := adminDB.ExecContext(context.Background(), "CREATE DATABASE `"+databaseName+"` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"); err != nil {
		adminDB.Close()
		t.Fatal(err)
	}

	testConfig := *adminConfig
	testConfig.DBName = databaseName
	db, err := gorm.Open(gormmysql.Open(testConfig.FormatDSN()), &gorm.Config{})
	if err != nil {
		adminDB.ExecContext(context.Background(), "DROP DATABASE `"+databaseName+"`")
		adminDB.Close()
		t.Fatal(err)
	}
	models := []any{
		&model.User{}, &model.Video{}, &model.Danmaku{}, &model.Comment{},
		&model.Like{}, &model.Collect{}, &model.CreatorDailyStat{}, &model.VideoDailyStat{},
		&model.Follow{}, &model.LiveSchedule{}, &model.LiveReservation{}, &model.Notification{},
		&model.LiveRoom{}, &model.LiveLike{}, &model.LiveGift{},
	}
	if err := db.AutoMigrate(models...); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
		_, _ = adminDB.ExecContext(context.Background(), "DROP DATABASE `"+databaseName+"`")
		_ = adminDB.Close()
	})
	return db
}
