package model

import "testing"

func TestOwnedTablesAndLegacyNames(t *testing.T) {
	if got := (VideoLike{}).TableName(); got != "video_likes" {
		t.Fatalf("video likes table=%q", got)
	}
	if got := (VideoCollection{}).TableName(); got != "video_collections" {
		t.Fatalf("video collections table=%q", got)
	}
	if got := (Danmaku{}).TableName(); got != "danmaku" {
		t.Fatalf("danmaku table=%q", got)
	}
	const want = 14
	if got := len(OwnedTables()); got != want {
		t.Fatalf("owned table count=%d want=%d", got, want)
	}
}
