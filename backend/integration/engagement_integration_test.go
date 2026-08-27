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
	"path/filepath"
	"strconv"
	"strings"
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
	"github.com/golang-jwt/jwt/v5"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type responseEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

const engagementTestSecret = "integration-test-secret"

func TestEngagementUseCasesWithMySQL(t *testing.T) {
	db := openTemporaryMySQL(t)
	testConfig := config.Config{}
	testConfig.Auth.AccessSecret = engagementTestSecret
	testConfig.Live.RTMPHost = "localhost:1935"
	testConfig.Live.HTTPHost = "localhost:8081"
	svcCtx := &svc.ServiceContext{
		DB:       db,
		Config:   testConfig,
		VideoDir: t.TempDir(),
	}

	creator := model.User{Username: "d_creator", Nickname: "D Creator", Password: "test", Role: "creator"}
	viewer := model.User{Username: "d_viewer", Nickname: "D Viewer", Password: "test", Role: "user"}
	admin := model.User{Username: "d_admin", Nickname: "D Admin", Password: "test", Role: "admin"}
	if err := db.Create(&creator).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&viewer).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}

	router := newEngagementRouter(svcCtx)
	var pendingScheduleForLive uint

	t.Run("UC05 video interaction persists and updates counts", func(t *testing.T) {
		video := model.Video{Title: "UC05", Status: "approved", AuthorID: creator.ID}
		if err := db.Create(&video).Error; err != nil {
			t.Fatal(err)
		}

		assertHTTPCode(t, router, http.MethodPost, "/danmaku", viewer.ID,
			map[string]any{"videoId": video.ID, "content": "hello", "time": 3}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/videos/%d/danmaku", video.ID), 0, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/videos/not-a-number/danmaku", 0, nil, http.StatusBadRequest)
		commentResponse := requestJSON(t, router, http.MethodPost, "/comments", viewer.ID,
			map[string]any{"videoId": video.ID, "content": "useful"})
		if commentResponse.Code != 0 {
			t.Fatalf("create comment failed: code=%d message=%q", commentResponse.Code, commentResponse.Message)
		}
		var createdComment struct {
			ID uint `json:"id"`
		}
		if err := json.Unmarshal(commentResponse.Data, &createdComment); err != nil || createdComment.ID == 0 {
			t.Fatalf("decode created comment: id=%d err=%v", createdComment.ID, err)
		}
		replyResponse := requestJSON(t, router, http.MethodPost, "/comments", creator.ID,
			map[string]any{"videoId": video.ID, "content": "reply", "parentId": createdComment.ID})
		if replyResponse.Code != 0 {
			t.Fatalf("create reply failed: code=%d message=%q", replyResponse.Code, replyResponse.Message)
		}
		var reply struct {
			ID uint `json:"id"`
		}
		if err := json.Unmarshal(replyResponse.Data, &reply); err != nil || reply.ID == 0 {
			t.Fatalf("decode reply: id=%d err=%v", reply.ID, err)
		}
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/comments/%d/like", createdComment.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/videos/%d/comments?sort=like", video.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/videos/%d/comments?sort=date", video.ID), 0, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/videos/%d/comments", video.ID), 0, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/videos/not-a-number/comments", 0, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/videos/%d/comments?sort=bad", video.ID), 0, nil, http.StatusBadRequest)
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
		assertModelCount(t, db, &model.Comment{}, "video_id = ?", 2, video.ID)
		assertModelCount(t, db, &model.CommentLike{}, "comment_id = ? AND user_id = ?", 1, createdComment.ID, viewer.ID)

		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/like", video.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/collect", video.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/comments/%d/like", createdComment.ID), viewer.ID, nil, http.StatusOK)
		if err := db.First(&refreshed, video.ID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.LikeCount != 0 || refreshed.CollectCount != 0 {
			t.Fatalf("counts after cancel = like:%d collect:%d, want 0/0", refreshed.LikeCount, refreshed.CollectCount)
		}
		assertModelCount(t, db, &model.CommentLike{}, "comment_id = ? AND user_id = ?", 0, createdComment.ID, viewer.ID)
		orphanParent := model.Comment{VideoID: video.ID, UserID: viewer.ID, Content: "temporary parent"}
		if err := db.Create(&orphanParent).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&model.Comment{VideoID: video.ID, UserID: viewer.ID, ParentID: &orphanParent.ID, Content: "orphan reply"}).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("SET FOREIGN_KEY_CHECKS = 0").Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Unscoped().Delete(&orphanParent).Error; err != nil {
			t.Fatal(err)
		}
		if err := db.Exec("SET FOREIGN_KEY_CHECKS = 1").Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/videos/%d/comments", video.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/danmaku", viewer.ID,
			map[string]any{"videoId": video.ID, "content": ""}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/comments", viewer.ID,
			map[string]any{"videoId": video.ID, "content": "   "}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/danmaku", viewer.ID,
			map[string]any{"videoId": video.ID + 9999, "content": "missing", "time": 1}, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, "/comments/999999/like", viewer.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, "/comments/not-a-number/like", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodDelete, fmt.Sprintf("/comments/%d", reply.ID), viewer.ID, nil, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodDelete, fmt.Sprintf("/comments/%d", reply.ID), admin.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodDelete, "/comments/not-a-number", admin.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodDelete, "/comments/999999", admin.ID, nil, http.StatusNotFound)
		deleteOwn := requestJSON(t, router, http.MethodPost, "/comments", viewer.ID,
			map[string]any{"videoId": video.ID, "content": "delete by owner"})
		var deleteOwnComment struct {
			ID uint `json:"id"`
		}
		if deleteOwn.Code != 0 || json.Unmarshal(deleteOwn.Data, &deleteOwnComment) != nil || deleteOwnComment.ID == 0 {
			t.Fatalf("create owner-delete comment failed: %+v", deleteOwn)
		}
		assertHTTPCode(t, router, http.MethodDelete, fmt.Sprintf("/comments/%d", deleteOwnComment.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/comments", viewer.ID, "malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/danmaku", viewer.ID, "malformed", http.StatusBadRequest)

		otherVideo := model.Video{Title: "other", Status: "approved", AuthorID: creator.ID}
		if err := db.Create(&otherVideo).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodPost, "/comments", viewer.ID,
			map[string]any{"videoId": otherVideo.ID, "content": "wrong parent", "parentId": createdComment.ID}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/comments", viewer.ID,
			map[string]any{"videoId": video.ID, "content": "missing parent", "parentId": 999999}, http.StatusNotFound)

		pendingVideo := model.Video{Title: "pending", Status: "pending", AuthorID: creator.ID}
		if err := db.Create(&pendingVideo).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodPost, "/danmaku", viewer.ID,
			map[string]any{"videoId": pendingVideo.ID, "content": "blocked"}, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, "/comments", viewer.ID,
			map[string]any{"videoId": pendingVideo.ID, "content": "blocked"}, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/like", pendingVideo.ID), viewer.ID, nil, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/collect", pendingVideo.ID), viewer.ID, nil, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, "/danmaku", 0,
			map[string]any{"videoId": video.ID, "content": "anonymous"}, http.StatusUnauthorized)
		assertHTTPCode(t, router, http.MethodPost, "/comments", 0,
			map[string]any{"videoId": video.ID, "content": "anonymous"}, http.StatusUnauthorized)
		assertHTTPCode(t, router, http.MethodPost, "/videos/not-a-number/like", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/videos/999999/like", viewer.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, "/videos/not-a-number/collect", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/videos/999999/collect", viewer.ID, nil, http.StatusNotFound)
	})

	t.Run("UC09 rejects conflicting schedule and toggles reservation", func(t *testing.T) {
		if err := db.Create(&model.Follow{FollowerID: viewer.ID, FolloweeID: creator.ID}).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodGet, "/live-schedules?status=invalid", 0, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/live-schedules", creator.ID, "malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/live-schedules", creator.ID,
			map[string]any{"title": "invalid time", "scheduledAt": "tomorrow"}, http.StatusBadRequest)
		scheduledAt := time.Now().Add(24 * time.Hour).Truncate(time.Second).Format(time.RFC3339)
		assertHTTPCode(t, router, http.MethodPost, "/live-schedules", creator.ID,
			map[string]any{"title": "past", "scheduledAt": time.Now().Add(-time.Hour).Format(time.RFC3339)}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/live-schedules", creator.ID,
			map[string]any{"title": "   ", "scheduledAt": scheduledAt}, http.StatusBadRequest)
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
		assertModelCount(t, db, &model.Notification{}, "user_id = ? AND type = ?", 1, viewer.ID, "live_schedule")

		assertHTTPCode(t, router, http.MethodPost, "/live-schedules", creator.ID,
			map[string]any{"title": "conflict", "scheduledAt": scheduledAt}, http.StatusConflict)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live-schedules/%d/reserve", created.ID), creator.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodDelete, fmt.Sprintf("/live-schedules/%d", created.ID), viewer.ID, nil, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live-schedules/%d/reserve", created.ID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/live-schedules?status=pending&page=0&pageSize=500", viewer.ID, nil, http.StatusOK)

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
		assertHTTPCode(t, router, http.MethodDelete, fmt.Sprintf("/live-schedules/%d", created.ID), creator.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live-schedules/%d/reserve", created.ID), viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodDelete, "/live-schedules/not-a-number", creator.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodDelete, "/live-schedules/999999", creator.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, "/live-schedules/not-a-number/reserve", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/live-schedules/999999/reserve", viewer.ID, nil, http.StatusNotFound)

		adminScheduleAt := time.Now().Add(48 * time.Hour).Truncate(time.Second).Format(time.RFC3339)
		adminCancelable := requestJSON(t, router, http.MethodPost, "/live-schedules", creator.ID,
			map[string]any{"title": "admin cancel", "scheduledAt": adminScheduleAt})
		var adminSchedule struct {
			ID uint `json:"id"`
		}
		if adminCancelable.Code != 0 || json.Unmarshal(adminCancelable.Data, &adminSchedule) != nil || adminSchedule.ID == 0 {
			t.Fatalf("create admin-cancel schedule failed: %+v", adminCancelable)
		}
		assertHTTPCode(t, router, http.MethodDelete, fmt.Sprintf("/live-schedules/%d", adminSchedule.ID), admin.ID, nil, http.StatusOK)

		pendingAt := time.Now().Add(72 * time.Hour).Truncate(time.Second).Format(time.RFC3339)
		pendingResponse := requestJSON(t, router, http.MethodPost, "/live-schedules", creator.ID,
			map[string]any{"title": "notify on manual live", "scheduledAt": pendingAt})
		var pending struct {
			ID uint `json:"id"`
		}
		if pendingResponse.Code != 0 || json.Unmarshal(pendingResponse.Data, &pending) != nil || pending.ID == 0 {
			t.Fatalf("create pending schedule failed: %+v", pendingResponse)
		}
		pendingScheduleForLive = pending.ID
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live-schedules/%d/reserve", pending.ID), viewer.ID, nil, http.StatusOK)
	})

	t.Run("UC10 create manage interact and end live room", func(t *testing.T) {
		assertHTTPCode(t, router, http.MethodPost, "/live", 0, map[string]any{"title": "anonymous"}, http.StatusUnauthorized)
		assertHTTPCode(t, router, http.MethodPost, "/live", creator.ID, "malformed", http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/live", creator.ID, map[string]any{"title": "   "}, http.StatusBadRequest)
		createdResponse := requestJSON(t, router, http.MethodPost, "/live", creator.ID, map[string]any{"title": "UC10"})
		if createdResponse.Code != 0 {
			t.Fatalf("create live failed: code=%d message=%q", createdResponse.Code, createdResponse.Message)
		}
		var createdRoom struct {
			ID        uint   `json:"id"`
			StreamKey string `json:"streamKey"`
		}
		if err := json.Unmarshal(createdResponse.Data, &createdRoom); err != nil || createdRoom.ID == 0 || createdRoom.StreamKey == "" {
			t.Fatalf("decode created live room: room=%+v err=%v", createdRoom, err)
		}
		roomID := createdRoom.ID
		var notifiedSchedule model.LiveSchedule
		if err := db.First(&notifiedSchedule, pendingScheduleForLive).Error; err != nil || notifiedSchedule.Status != "live" {
			t.Fatalf("manual live should activate reserved schedule: status=%q err=%v", notifiedSchedule.Status, err)
		}
		// The viewer receives one notification for the reservation and one for following the creator.
		assertModelCount(t, db, &model.Notification{}, "user_id = ? AND type = ?", 2, viewer.ID, "live_start")
		assertHTTPCode(t, router, http.MethodGet, "/live?page=0&pageSize=999", 0, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d", roomID), 0, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/live/not-a-number", 0, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, "/live/999999", 0, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/manage", roomID), viewer.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/manage", roomID), creator.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/live/not-a-number/manage", creator.ID, nil, http.StatusBadRequest)

		repeated := requestJSON(t, router, http.MethodPost, "/live", creator.ID, map[string]any{"title": "UC10 repeated"})
		var repeatedRoom struct {
			ID uint `json:"id"`
		}
		if repeated.Code != 0 || json.Unmarshal(repeated.Data, &repeatedRoom) != nil || repeatedRoom.ID != roomID {
			t.Fatalf("repeated create should reuse live room: code=%d room=%+v", repeated.Code, repeatedRoom)
		}

		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/chat-settings", roomID), creator.ID,
			map[string]any{"chatMode": "invalid", "slowModeSeconds": 0}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, "/live/not-a-number/chat-settings", creator.ID,
			map[string]any{"chatMode": "everyone", "slowModeSeconds": 0}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/chat-settings", roomID), creator.ID,
			map[string]any{"chatMode": "everyone", "slowModeSeconds": -1}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/chat-settings", roomID), creator.ID,
			map[string]any{"chatMode": "everyone", "slowModeSeconds": 121}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/chat-settings", roomID), creator.ID,
			map[string]any{"chatMode": "everyone", "slowModeSeconds": 0, "pinnedMessage": strings.Repeat("公", 201)}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, "/live/999999/chat-settings", creator.ID,
			map[string]any{"chatMode": "everyone", "slowModeSeconds": 0}, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/chat-settings", roomID), viewer.ID,
			map[string]any{"chatMode": "followers", "slowModeSeconds": 5, "pinnedMessage": "notice"}, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/chat-settings", roomID), creator.ID,
			map[string]any{"chatMode": "followers", "slowModeSeconds": 5, "pinnedMessage": "notice"}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/chat-settings", roomID), admin.ID,
			map[string]any{"chatMode": "everyone", "slowModeSeconds": 0, "pinnedMessage": "admin notice"}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/chat-settings", roomID), creator.ID, "malformed", http.StatusBadRequest)

		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/like-status", roomID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/live/not-a-number/like-status", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/like", roomID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/like", roomID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/like", roomID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/like-status", roomID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/gifts", roomID), viewer.ID,
			map[string]any{"giftKey": "star", "count": 2, "message": "支持主播"}, http.StatusOK)
		if err := db.Create(&model.Danmaku{VideoID: roomID, Scene: "live", UserID: viewer.ID, Content: "直播弹幕", Color: "#fff", FontSize: "medium", Type: "scroll"}).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/interaction", roomID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/live/not-a-number/interaction", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, "/live/999999/interaction", viewer.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/danmaku", roomID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/live/not-a-number/danmaku", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodGet, "/live/999999/danmaku", viewer.ID, nil, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/monitor", roomID), viewer.ID, nil, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/monitor", roomID), creator.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodGet, "/live/not-a-number/monitor", creator.ID, nil, http.StatusBadRequest)
		otherCreator := model.User{Username: "rank_creator", Nickname: "Rank Creator", Password: "test", Role: "creator"}
		if err := db.Create(&otherCreator).Error; err != nil {
			t.Fatal(err)
		}
		otherStarted := time.Now()
		if err := db.Create(&model.LiveRoom{Title: "rank", OwnerID: otherCreator.ID, StreamKey: "rank-key", Status: "live", StartedAt: &otherStarted, ViewerCount: 100}).Error; err != nil {
			t.Fatal(err)
		}
		assertHTTPCode(t, router, http.MethodGet, "/live/heat-ranking", 0, nil, http.StatusOK)

		var refreshed model.LiveRoom
		refreshed = model.LiveRoom{}
		if err := db.First(&refreshed, roomID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.LikeCount != 1 || refreshed.GiftValue != 100 {
			t.Fatalf("live counts = like:%d gift:%d, want 1/100", refreshed.LikeCount, refreshed.GiftValue)
		}
		var gift model.LiveGift
		if err := db.Where("room_id = ? AND user_id = ?", roomID, viewer.ID).First(&gift).Error; err != nil {
			t.Fatal(err)
		}
		if gift.DisplaySeconds != 30 || gift.Message != "支持主播" {
			t.Fatalf("Super Chat = display:%d message:%q, want 30/支持主播", gift.DisplaySeconds, gift.Message)
		}

		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/like", roomID), viewer.ID, nil, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/gifts", roomID), viewer.ID,
			map[string]any{"giftKey": "star", "count": 100}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/live/not-a-number/gifts", viewer.ID,
			map[string]any{"giftKey": "flower", "count": 1}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/gifts", roomID), viewer.ID,
			map[string]any{"giftKey": "missing", "count": 1}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/gifts", roomID), viewer.ID,
			map[string]any{"giftKey": "flower", "count": 0}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/gifts", roomID), viewer.ID,
			map[string]any{"giftKey": "flower", "count": 1, "message": strings.Repeat("长", 201)}, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/gifts", roomID), viewer.ID,
			map[string]any{"giftKey": "flower", "count": 1}, http.StatusOK)
		assertHTTPCode(t, router, http.MethodPost, "/live/999999/gifts", viewer.ID,
			map[string]any{"giftKey": "flower", "count": 1}, http.StatusNotFound)
		assertHTTPCode(t, router, http.MethodPost, "/live/not-a-number/like", viewer.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPost, "/live/999999/like", viewer.ID, nil, http.StatusNotFound)
		refreshed = model.LiveRoom{}
		if err := db.First(&refreshed, roomID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.LikeCount != 0 || refreshed.GiftValue != 110 {
			t.Fatalf("live counts after unlike/gifts = like:%d gift:%d, want 0/110", refreshed.LikeCount, refreshed.GiftValue)
		}

		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/end", roomID), viewer.ID, nil, http.StatusForbidden)
		assertHTTPCode(t, router, http.MethodPut, "/live/not-a-number/end", creator.ID, nil, http.StatusBadRequest)
		assertHTTPCode(t, router, http.MethodPut, "/live/999999/end", creator.ID, nil, http.StatusNotFound)
		prepareReplayPlaylist(t, svcCtx.VideoDir, createdRoom.StreamKey)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/end", roomID), creator.ID, nil, http.StatusOK)
		refreshed = model.LiveRoom{}
		if err := db.First(&refreshed, roomID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.Status != "ended" || refreshed.ViewerCount != 0 || refreshed.EndedAt == nil {
			t.Fatalf("ended room = status:%s viewers:%d endedAt:%v", refreshed.Status, refreshed.ViewerCount, refreshed.EndedAt)
		}
		assertModelCount(t, db, &model.LiveReplay{}, "room_id = ? AND stream_key = ?", 1, roomID, createdRoom.StreamKey)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d", roomID), 0, nil, http.StatusNotFound)
		restarted := requestJSON(t, router, http.MethodPost, "/live", creator.ID, map[string]any{"title": "UC10 restarted"})
		var restartedRoom struct {
			ID        uint   `json:"id"`
			StreamKey string `json:"streamKey"`
		}
		if restarted.Code != 0 || json.Unmarshal(restarted.Data, &restartedRoom) != nil || restartedRoom.ID != roomID || restartedRoom.StreamKey == createdRoom.StreamKey {
			t.Fatalf("ended room should be reset and reused with a new key: code=%d room=%+v", restarted.Code, restartedRoom)
		}
		refreshed = model.LiveRoom{}
		if err := db.First(&refreshed, roomID).Error; err != nil {
			t.Fatal(err)
		}
		if refreshed.Status != "live" || refreshed.LikeCount != 0 || refreshed.GiftValue != 0 || refreshed.EndedAt != nil {
			t.Fatalf("restarted room not reset: status=%s likes=%d gifts=%d endedAt=%v", refreshed.Status, refreshed.LikeCount, refreshed.GiftValue, refreshed.EndedAt)
		}
		prepareReplayPlaylist(t, svcCtx.VideoDir, restartedRoom.StreamKey)
		assertHTTPCode(t, router, http.MethodPut, fmt.Sprintf("/live/%d/end", roomID), admin.ID, nil, http.StatusOK)
		time.Sleep(1200 * time.Millisecond) // allow both replay finalizers to exit before the temporary DB closes
	})
}

func TestEngagementDatabaseFailureResponses(t *testing.T) {
	db := openTemporaryMySQL(t)
	testConfig := config.Config{}
	testConfig.Auth.AccessSecret = engagementTestSecret
	testConfig.Live.RTMPHost = "localhost:1935"
	testConfig.Live.HTTPHost = "localhost:8081"
	svcCtx := &svc.ServiceContext{DB: db, Config: testConfig, VideoDir: t.TempDir()}
	user := model.User{Username: "failure_user", Nickname: "Failure User", Password: "test", Role: "creator"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	video := model.Video{Title: "failure video", Status: "approved", AuthorID: user.ID}
	if err := db.Create(&video).Error; err != nil {
		t.Fatal(err)
	}
	router := newEngagementRouter(svcCtx)

	drop := func(modelValue any) {
		t.Helper()
		if err := db.Migrator().DropTable(modelValue); err != nil {
			t.Fatal(err)
		}
	}
	restore := func(models ...any) {
		t.Helper()
		if err := db.AutoMigrate(models...); err != nil {
			t.Fatal(err)
		}
	}

	t.Run("UC05 persistence tables unavailable", func(t *testing.T) {
		drop(&model.Danmaku{})
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/videos/%d/danmaku", video.ID), 0, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, "/danmaku", user.ID,
			map[string]any{"videoId": video.ID, "content": "cannot persist"}, http.StatusInternalServerError)
		restore(&model.Danmaku{})

		drop(&model.Comment{})
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/videos/%d/comments", video.ID), 0, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, "/comments", user.ID,
			map[string]any{"videoId": video.ID, "content": "cannot persist"}, http.StatusInternalServerError)
		restore(&model.Comment{})
		comment := model.Comment{VideoID: video.ID, UserID: user.ID, Content: "like failure"}
		if err := db.Create(&comment).Error; err != nil {
			t.Fatal(err)
		}
		drop(&model.CommentLike{})
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/videos/%d/comments", video.ID), user.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/comments/%d/like", comment.ID), user.ID, nil, http.StatusInternalServerError)
		restore(&model.CommentLike{})

		drop(&model.Like{})
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/like", video.ID), user.ID, nil, http.StatusInternalServerError)
		restore(&model.Like{})
		drop(&model.Collect{})
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/videos/%d/collect", video.ID), user.ID, nil, http.StatusInternalServerError)
		restore(&model.Collect{})
	})

	t.Run("UC09 schedule tables unavailable", func(t *testing.T) {
		drop(&model.LiveSchedule{})
		assertHTTPCode(t, router, http.MethodGet, "/live-schedules", user.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, "/live-schedules", user.ID,
			map[string]any{"title": "db down", "scheduledAt": time.Now().Add(time.Hour).Format(time.RFC3339)}, http.StatusInternalServerError)
		restore(&model.LiveSchedule{})

		schedule := model.LiveSchedule{Title: "reservation query", ScheduledAt: time.Now().Add(time.Hour), Status: "pending", OwnerID: user.ID}
		if err := db.Create(&schedule).Error; err != nil {
			t.Fatal(err)
		}
		drop(&model.LiveReservation{})
		assertHTTPCode(t, router, http.MethodGet, "/live-schedules", user.ID, nil, http.StatusInternalServerError)
		restore(&model.LiveReservation{})
	})

	t.Run("UC10 live interaction tables unavailable", func(t *testing.T) {
		endedAt := time.Now()
		endedRoom := model.LiveRoom{Title: "ended failure room", OwnerID: user.ID, StreamKey: "ended-failure-key", Status: "ended", EndedAt: &endedAt}
		if err := db.Create(&endedRoom).Error; err != nil {
			t.Fatal(err)
		}
		drop(&model.LiveLike{})
		assertHTTPCode(t, router, http.MethodPost, "/live", user.ID, map[string]any{"title": "reset likes fails"}, http.StatusInternalServerError)
		restore(&model.LiveLike{})
		drop(&model.LiveGift{})
		assertHTTPCode(t, router, http.MethodPost, "/live", user.ID, map[string]any{"title": "reset gifts fails"}, http.StatusInternalServerError)
		restore(&model.LiveGift{})

		drop(&model.LiveRoom{})
		assertHTTPCode(t, router, http.MethodGet, "/live", 0, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, "/live/1", 0, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, "/live/1/manage", user.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, "/live", user.ID, map[string]any{"title": "db down"}, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPut, "/live/1/end", user.ID, nil, http.StatusInternalServerError)
		restore(&model.LiveRoom{})

		startedAt := time.Now()
		room := model.LiveRoom{Title: "failure room", OwnerID: user.ID, StreamKey: "failure-key", Status: "live", StartedAt: &startedAt}
		if err := db.Create(&room).Error; err != nil {
			t.Fatal(err)
		}
		drop(&model.LiveLike{})
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/like-status", room.ID), user.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/like", room.ID), user.ID, nil, http.StatusInternalServerError)
		restore(&model.LiveLike{})

		drop(&model.LiveGift{})
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/interaction", room.ID), user.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodPost, fmt.Sprintf("/live/%d/gifts", room.ID), user.ID,
			map[string]any{"giftKey": "flower", "count": 1}, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/monitor", room.ID), user.ID, nil, http.StatusInternalServerError)
		restore(&model.LiveGift{})

		drop(&model.Danmaku{})
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/danmaku", room.ID), user.ID, nil, http.StatusInternalServerError)
		assertHTTPCode(t, router, http.MethodGet, fmt.Sprintf("/live/%d/monitor", room.ID), user.ID, nil, http.StatusInternalServerError)
		restore(&model.Danmaku{})
	})
}

func newEngagementRouter(svcCtx *svc.ServiceContext) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if raw := c.GetHeader("X-Test-User-ID"); raw != "" {
			id, _ := strconv.ParseUint(raw, 10, 64)
			c.Set(middleware.CtxKeyUserID, uint(id))
			if id > 0 {
				var user model.User
				if svcCtx.DB.Select("role").First(&user, uint(id)).Error == nil {
					c.Set(middleware.CtxKeyRole, user.Role)
				}
			}
		}
		c.Next()
	})
	router.POST("/danmaku", danmakuhandler.SendHandler(svcCtx))
	router.GET("/videos/:videoId/danmaku", danmakuhandler.ListHandler(svcCtx))
	router.POST("/comments", commenthandler.CreateHandler(svcCtx))
	router.GET("/videos/:videoId/comments", commenthandler.ListHandler(svcCtx))
	router.POST("/comments/:id/like", commenthandler.LikeHandler(svcCtx))
	router.DELETE("/comments/:id", commenthandler.DeleteHandler(svcCtx))
	router.POST("/videos/:id/like", videohandler.LikeHandler(svcCtx))
	router.POST("/videos/:id/collect", videohandler.CollectHandler(svcCtx))
	router.POST("/live-schedules", livehandler.CreateScheduleHandler(svcCtx))
	router.GET("/live-schedules", livehandler.ScheduleListHandler(svcCtx))
	router.DELETE("/live-schedules/:id", livehandler.CancelScheduleHandler(svcCtx))
	router.POST("/live-schedules/:id/reserve", livehandler.ReserveScheduleHandler(svcCtx))
	router.POST("/live", livehandler.CreateHandler(svcCtx))
	router.GET("/live", livehandler.ListHandler(svcCtx))
	router.GET("/live/heat-ranking", livehandler.HeatRankingHandler(svcCtx))
	router.GET("/live/:id", livehandler.DetailHandler(svcCtx))
	router.GET("/live/:id/manage", livehandler.ManageDetailHandler(svcCtx))
	router.PUT("/live/:id/chat-settings", livehandler.UpdateChatSettingsHandler(svcCtx))
	router.POST("/live/:id/like", livehandler.LikeHandler(svcCtx))
	router.GET("/live/:id/like-status", livehandler.LikeStatusHandler(svcCtx))
	router.POST("/live/:id/gifts", livehandler.GiftHandler(svcCtx))
	router.GET("/live/:id/interaction", livehandler.InteractionHandler(svcCtx))
	router.GET("/live/:id/danmaku", livehandler.CurrentDanmakuHandler(svcCtx))
	router.GET("/live/:id/monitor", livehandler.MonitorHandler(svcCtx))
	router.PUT("/live/:id/end", livehandler.EndHandler(svcCtx))
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
	if userID != 0 {
		req.Header.Set("Authorization", "Bearer "+testToken(t, userID))
	}
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
	if userID != 0 {
		req.Header.Set("Authorization", "Bearer "+testToken(t, userID))
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)
	if recorder.Code != want {
		t.Fatalf("%s %s status = %d, want %d; body=%s", method, path, recorder.Code, want, recorder.Body.String())
	}
}

func testToken(t *testing.T, userID uint) string {
	t.Helper()
	claims := middleware.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(engagementTestSecret))
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func prepareReplayPlaylist(t *testing.T, videoDir, streamKey string) {
	t.Helper()
	dir := filepath.Join(videoDir, "live")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, streamKey+".m3u8")
	if err := os.WriteFile(path, []byte("#EXTM3U\n#EXT-X-ENDLIST\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	old := time.Now().Add(-3 * time.Second)
	if err := os.Chtimes(path, old, old); err != nil {
		t.Fatal(err)
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
		&model.User{}, &model.Video{}, &model.Danmaku{}, &model.Comment{}, &model.CommentLike{},
		&model.Like{}, &model.Collect{}, &model.CreatorDailyStat{}, &model.VideoDailyStat{},
		&model.Follow{}, &model.FollowGroup{}, &model.UserBlock{}, &model.Notification{},
		&model.ChatMessage{}, &model.CreatorMembershipPlan{}, &model.CreatorSubscription{}, &model.SubscriptionOrder{},
		&model.LiveSchedule{}, &model.LiveReservation{},
		&model.LiveRoom{}, &model.LiveLike{}, &model.LiveGift{}, &model.LiveReplay{},
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
