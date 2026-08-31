package videologic

import (
	model "danmakustream/backend/internal/model/mysql"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

var ErrMediaUnavailable = errors.New("视频媒体暂时不可用")

func EffectiveTranscodeStatus(video model.Video) string {
	switch video.TranscodeStatus {
	case "processing", "ready", "failed":
		return video.TranscodeStatus
	}
	if video.VideoURL != "" {
		return "ready"
	}
	return "processing"
}

// MediaPathToLocalPath resolves an internal /media URL below videoDir and
// rejects remote URLs and traversal outside the configured media root.
func MediaPathToLocalPath(videoDir, mediaURL string) (string, error) {
	if strings.HasPrefix(mediaURL, "http://") || strings.HasPrefix(mediaURL, "https://") {
		return "", ErrMediaUnavailable
	}
	cleanURL := strings.TrimPrefix(mediaURL, "/")
	if !strings.HasPrefix(cleanURL, "media/") {
		return "", ErrMediaUnavailable
	}
	relative := filepath.FromSlash(strings.TrimPrefix(cleanURL, "media/"))
	root, err := filepath.Abs(videoDir)
	if err != nil {
		return "", ErrMediaUnavailable
	}
	resolved, err := filepath.Abs(filepath.Join(root, relative))
	if err != nil {
		return "", ErrMediaUnavailable
	}
	rel, err := filepath.Rel(root, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", ErrMediaUnavailable
	}
	return resolved, nil
}

func EnsureMediaAvailable(videoDir, mediaURL string) error {
	path, err := MediaPathToLocalPath(videoDir, mediaURL)
	if err != nil {
		return ErrMediaUnavailable
	}
	info, err := os.Stat(path)
	if err != nil || !info.Mode().IsRegular() {
		return ErrMediaUnavailable
	}
	return nil
}
