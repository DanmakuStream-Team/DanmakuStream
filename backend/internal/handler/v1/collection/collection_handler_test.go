package collection

import (
	"net/http/httptest"
	"testing"
	"time"

	model "danmakustream/backend/internal/model/mysql"

	"github.com/gin-gonic/gin"
)

func TestParseID(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		value string
		want  uint
		ok    bool
	}{
		{name: "valid", value: "8", want: 8, ok: true},
		{name: "zero", value: "0"},
		{name: "invalid", value: "collection"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			recorder := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(recorder)
			ctx.Params = gin.Params{{Key: "id", Value: test.value}}
			got, ok := parseID(ctx, "id")
			if got != test.want || ok != test.ok {
				t.Fatalf("parseID(%q) = (%d,%v), want (%d,%v)", test.value, got, ok, test.want, test.ok)
			}
			if !test.ok && recorder.Code != 400 {
				t.Fatalf("invalid id status = %d, want 400", recorder.Code)
			}
		})
	}
}

func TestToCollectionInfoMapsOwnerAndVideos(t *testing.T) {
	t.Parallel()
	owner := model.User{Username: "11", Nickname: "合集作者", Avatar: "/avatar.png", Role: "creator"}
	owner.ID = 11
	collection := model.VideoCollection{Title: "课程合集", Description: "测试", CoverURL: "/cover.png", Owner: owner}
	collection.ID = 12
	collection.CreatedAt = time.Date(2026, time.August, 29, 10, 0, 0, 0, time.Local)

	info := toCollectionInfo(collection, nil)
	if info.ID != 12 || info.Title != "课程合集" || info.Owner == nil || info.Owner.ID != 11 {
		t.Fatalf("collection mapping incomplete: %+v", info)
	}
	if info.CreatedAt != "2026-08-29 10:00:00" {
		t.Fatalf("createdAt = %q", info.CreatedAt)
	}
}
