package videologic

import (
	model "danmakustream/backend/internal/model/mysql"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestMediaPathToLocalPath(t *testing.T) {
	root := t.TempDir()
	want := filepath.Join(root, "videos", "12", "playlist.m3u8")
	for _, input := range []string{"/media/videos/12/playlist.m3u8", "media/videos/12/playlist.m3u8"} {
		got, err := MediaPathToLocalPath(root, input)
		if err != nil || got != want {
			t.Errorf("MediaPathToLocalPath(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	for _, input := range []string{
		"https://example.com/a.mp4",
		"/data/videos/a.mp4",
		"/media/../etc/passwd",
		"/media/videos/../../outside.mp4",
	} {
		if got, err := MediaPathToLocalPath(root, input); !errors.Is(err, ErrMediaUnavailable) {
			t.Errorf("MediaPathToLocalPath(%q) = %q, %v; want ErrMediaUnavailable", input, got, err)
		}
	}
}

func TestEnsureMediaAvailable(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "videos", "12", "playlist.m3u8")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureMediaAvailable(root, "/media/videos/12/playlist.m3u8"); !errors.Is(err, ErrMediaUnavailable) {
		t.Fatalf("missing media error = %v, want ErrMediaUnavailable", err)
	}
	if err := os.WriteFile(path, []byte("#EXTM3U\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureMediaAvailable(root, "/media/videos/12/playlist.m3u8"); err != nil {
		t.Fatalf("existing media rejected: %v", err)
	}
}

func TestEffectiveTranscodeStatusKeepsLegacyRowsCompatible(t *testing.T) {
	if got := EffectiveTranscodeStatus(model.Video{VideoURL: "/media/videos/1/playlist.m3u8"}); got != "ready" {
		t.Fatalf("legacy media status = %q, want ready", got)
	}
	if got := EffectiveTranscodeStatus(model.Video{}); got != "processing" {
		t.Fatalf("legacy empty status = %q, want processing", got)
	}
	if got := EffectiveTranscodeStatus(model.Video{VideoURL: "/media/a", TranscodeStatus: "failed"}); got != "failed" {
		t.Fatalf("explicit failed status = %q, want failed", got)
	}
}
