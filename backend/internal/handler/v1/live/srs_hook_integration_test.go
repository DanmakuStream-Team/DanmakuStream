//go:build integration

package live

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	model "danmakustream/backend/internal/model/mysql"
	"danmakustream/backend/internal/svc"
	"danmakustream/backend/internal/testutil"

	"github.com/gin-gonic/gin"
)

func TestSRSStreamHookEndsOnlyAbandonedStream(t *testing.T) {
	db := testutil.OpenTemporaryMySQL(t, &model.User{}, &model.LiveRoom{})
	svcCtx := &svc.ServiceContext{DB: db}
	owner := model.User{Username: "srs_owner", Nickname: "SRS Owner", Password: "test", Role: "creator"}
	if err := db.Create(&owner).Error; err != nil {
		t.Fatal(err)
	}
	reconnectedOwner := model.User{Username: "srs_reconnected_owner", Nickname: "SRS Reconnected Owner", Password: "test", Role: "creator"}
	if err := db.Create(&reconnectedOwner).Error; err != nil {
		t.Fatal(err)
	}

	room := model.LiveRoom{Title: "abandoned", Status: "live", StreamKey: "srs-abandoned", OwnerID: owner.ID, ViewerCount: 3}
	if err := db.Create(&room).Error; err != nil {
		t.Fatal(err)
	}
	assertSRSHookResponse(t, svcCtx, `{"action":"on_publish","app":"live","stream":"srs-abandoned"}`, http.StatusOK, `"code":0`)
	generation := markSRSStream(room.StreamKey, false)
	if err := finalizeUnpublishedSRSStream(svcCtx, room.StreamKey, generation, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&room, room.ID).Error; err != nil {
		t.Fatal(err)
	}
	if room.Status != "ended" || room.ViewerCount != 0 || room.EndedAt == nil {
		t.Fatalf("abandoned room = status:%q viewers:%d endedAt:%v", room.Status, room.ViewerCount, room.EndedAt)
	}

	reconnected := model.LiveRoom{Title: "reconnected", Status: "live", StreamKey: "srs-reconnected", OwnerID: reconnectedOwner.ID, ViewerCount: 1}
	if err := db.Create(&reconnected).Error; err != nil {
		t.Fatal(err)
	}
	oldGeneration := markSRSStream(reconnected.StreamKey, false)
	markSRSStream(reconnected.StreamKey, true)
	if err := finalizeUnpublishedSRSStream(svcCtx, reconnected.StreamKey, oldGeneration, time.Now()); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&reconnected, reconnected.ID).Error; err != nil {
		t.Fatal(err)
	}
	if reconnected.Status != "live" {
		t.Fatalf("reconnected room status = %q, want live", reconnected.Status)
	}

	assertSRSHookResponse(t, svcCtx, `{"action":"on_publish","app":"live","stream":"unknown"}`, http.StatusOK, `"code":1`)
}

func assertSRSHookResponse(t *testing.T, svcCtx *svc.ServiceContext, body string, wantStatus int, wantBody string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/hook", SRSStreamHookHandler(svcCtx))
	request := httptest.NewRequest(http.MethodPost, "/hook", bytes.NewBufferString(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != wantStatus || !bytes.Contains(recorder.Body.Bytes(), []byte(wantBody)) {
		t.Fatalf("hook response = %d %s, want %d containing %s", recorder.Code, recorder.Body.String(), wantStatus, wantBody)
	}
}
