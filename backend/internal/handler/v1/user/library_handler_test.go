package user

import (
	"net/http/httptest"
	"testing"
	"time"

	model "danmakustream/backend/internal/model/mysql"

	"github.com/gin-gonic/gin"
)

func TestLibraryPageNormalizesBounds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name               string
		query              string
		wantPage, wantSize int
	}{
		{name: "defaults", wantPage: 1, wantSize: 20},
		{name: "valid", query: "?page=3&pageSize=50", wantPage: 3, wantSize: 50},
		{name: "negative page", query: "?page=-2&pageSize=10", wantPage: 1, wantSize: 10},
		{name: "zero size", query: "?page=2&pageSize=0", wantPage: 2, wantSize: 20},
		{name: "oversized", query: "?page=2&pageSize=101", wantPage: 2, wantSize: 20},
		{name: "invalid numbers", query: "?page=x&pageSize=y", wantPage: 1, wantSize: 20},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
			ctx.Request = httptest.NewRequest("GET", "/users/me/history"+test.query, nil)
			page, size := libraryPage(ctx)
			if page != test.wantPage || size != test.wantSize {
				t.Fatalf("libraryPage() = (%d,%d), want (%d,%d)", page, size, test.wantPage, test.wantSize)
			}
		})
	}
}

func TestLibraryVideoID(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name   string
		value  string
		wantID uint64
		ok     bool
	}{
		{name: "valid", value: "42", wantID: 42, ok: true},
		{name: "zero", value: "0"},
		{name: "negative", value: "-1"},
		{name: "not number", value: "video"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Params = gin.Params{{Key: "videoId", Value: test.value}}
			id, ok := libraryVideoID(ctx)
			if id != test.wantID || ok != test.ok {
				t.Fatalf("libraryVideoID(%q) = (%d,%v), want (%d,%v)", test.value, id, ok, test.wantID, test.ok)
			}
			if !test.ok && recorder.Code != 400 {
				t.Fatalf("invalid id status = %d, want 400", recorder.Code)
			}
		})
	}
}

func TestToLibraryRecordCalculatesProgressAndMapsVideo(t *testing.T) {
	t.Parallel()
	savedAt := time.Date(2026, time.August, 29, 9, 30, 0, 0, time.Local)
	author := model.User{Username: "7", Nickname: "作者", Role: "creator"}
	author.ID = 7
	video := model.Video{
		Title: "测试视频", Description: "简介", Duration: 120, Status: "approved",
		ViewCount: 10, LikeCount: 2, CollectCount: 1, Author: author,
	}
	video.ID = 9
	video.CreatedAt = savedAt.Add(-time.Hour)

	record := toLibraryRecord(video, savedAt, 45)
	if record.Video.ID != 9 || record.Video.Title != "测试视频" || record.Video.Author == nil || record.Video.Author.ID != 7 {
		t.Fatalf("video mapping incomplete: %+v", record.Video)
	}
	if record.Position != 45 || record.Progress != 37 {
		t.Fatalf("position/progress = %d/%d, want 45/37", record.Position, record.Progress)
	}
	if record.SavedAt != "2026-08-29 09:30:00" {
		t.Fatalf("savedAt = %q", record.SavedAt)
	}

	clamped := toLibraryRecord(video, savedAt, 999)
	if clamped.Progress != 100 {
		t.Fatalf("progress = %d, want clamp to 100", clamped.Progress)
	}
	video.Duration = 0
	zeroDuration := toLibraryRecord(video, savedAt, 45)
	if zeroDuration.Progress != 0 {
		t.Fatalf("zero-duration progress = %d, want 0", zeroDuration.Progress)
	}
}
