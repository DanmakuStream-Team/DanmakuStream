package handler

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"danmakustream/content-service/internal/logic"
	"danmakustream/content-service/internal/middleware"
	"danmakustream/content-service/internal/model"
	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var errUploadTooLarge = errors.New("upload too large")

func (h *Handler) UploadVideo(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.Config.MaxVideoBytes+h.Config.MaxImageBytes+(2<<20))
	file, err := c.FormFile("video")
	if err != nil {
		writeUploadError(c, err)
		return
	}
	title := strings.TrimSpace(c.PostForm("title"))
	if title == "" || len(title) > 200 {
		response.Error(c, http.StatusBadRequest, 40002, "invalid title")
		return
	}
	url, mimeType, size, err := h.saveUpload(file, "videos", h.Config.MaxVideoBytes, map[string]bool{
		".mp4": true, ".webm": true, ".mov": true, ".mkv": true,
	})
	if err != nil {
		writeUploadError(c, err)
		return
	}
	video := model.Video{
		Title: title, Description: c.PostForm("description"), Tags: c.PostForm("tags"),
		Category: c.PostForm("category"), VideoURL: url, AuthorID: middleware.UserID(c), TranscodeStatus: "ready",
	}
	coverURL, coverMime, coverSize := "", "", int64(0)
	if cover, coverErr := c.FormFile("cover"); coverErr == nil {
		coverURL, coverMime, coverSize, err = h.saveUpload(cover, "covers", h.Config.MaxImageBytes, map[string]bool{
			".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
		})
		if err != nil {
			removeMedia(h.Config.StorageDir, url)
			writeUploadError(c, err)
			return
		}
		video.CoverURL = coverURL
	} else if !errors.Is(coverErr, http.ErrMissingFile) {
		removeMedia(h.Config.StorageDir, url)
		writeUploadError(c, coverErr)
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		service := logic.Service{DB: tx}
		if err := service.CreateVideo(&video); err != nil {
			return err
		}
		assets := []model.MediaAsset{{OwnerID: video.AuthorID, VideoID: &video.ID, Kind: "video", Path: url, MimeType: mimeType, Size: size}}
		if coverURL != "" {
			assets = append(assets, model.MediaAsset{OwnerID: video.AuthorID, VideoID: &video.ID, Kind: "cover", Path: coverURL, MimeType: coverMime, Size: coverSize})
		}
		return tx.Create(&assets).Error
	}); err != nil {
		removeMedia(h.Config.StorageDir, url)
		removeMedia(h.Config.StorageDir, coverURL)
		writeLogicError(c, err)
		return
	}
	response.Created(c, logic.VideoView(video))
}

func removeMedia(root, mediaURL string) {
	if mediaURL == "" {
		return
	}
	if path, err := localMediaPath(root, mediaURL); err == nil {
		_ = os.Remove(path)
	}
}

func (h *Handler) UploadImage(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.Config.MaxImageBytes+(1<<20))
	file, err := c.FormFile("image")
	if err != nil {
		writeUploadError(c, err)
		return
	}
	url, mimeType, size, err := h.saveUpload(file, "images", h.Config.MaxImageBytes, map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true,
	})
	if err != nil {
		writeUploadError(c, err)
		return
	}
	asset := model.MediaAsset{OwnerID: middleware.UserID(c), Kind: "image", Path: url, MimeType: mimeType, Size: size}
	if err := h.DB.Create(&asset).Error; err != nil {
		if path, pathErr := localMediaPath(h.Config.StorageDir, url); pathErr == nil {
			_ = os.Remove(path)
		}
		writeLogicError(c, err)
		return
	}
	response.Created(c, asset)
}

func (h *Handler) UpdateCover(c *gin.Context) {
	videoID, ok := uintParam(c, "id")
	if !ok {
		return
	}
	allowed, err := h.Logic.CanEditVideo(videoID, middleware.UserID(c))
	if err != nil {
		writeLogicError(c, err)
		return
	}
	if !allowed {
		writeLogicError(c, logic.ErrForbidden)
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.Config.MaxImageBytes+(1<<20))
	file, err := c.FormFile("cover")
	if err != nil {
		writeUploadError(c, err)
		return
	}
	url, mimeType, size, err := h.saveUpload(file, "covers", h.Config.MaxImageBytes, map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".webp": true,
	})
	if err != nil {
		writeUploadError(c, err)
		return
	}
	if err := h.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.Video{}).Where("id = ?", videoID).Updates(map[string]any{"cover_url": url, "status": "pending"}).Error; err != nil {
			return err
		}
		asset := model.MediaAsset{OwnerID: middleware.UserID(c), VideoID: &videoID, Kind: "cover", Path: url, MimeType: mimeType, Size: size}
		return tx.Create(&asset).Error
	}); err != nil {
		if path, pathErr := localMediaPath(h.Config.StorageDir, url); pathErr == nil {
			_ = os.Remove(path)
		}
		writeLogicError(c, err)
		return
	}
	video, err := h.Logic.GetVideo(videoID, true)
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, logic.VideoView(video))
}

func (h *Handler) saveUpload(header *multipart.FileHeader, directory string, limit int64, extensions map[string]bool) (string, string, int64, error) {
	if header.Size <= 0 {
		return "", "", 0, errors.New("empty upload")
	}
	if header.Size > limit {
		return "", "", 0, errUploadTooLarge
	}
	ext := strings.ToLower(filepath.Ext(filepath.Base(header.Filename)))
	if !extensions[ext] {
		return "", "", 0, errors.New("unsupported file type")
	}
	source, err := header.Open()
	if err != nil {
		return "", "", 0, err
	}
	defer source.Close()
	var prefix [512]byte
	n, err := io.ReadFull(source, prefix[:])
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", "", 0, err
	}
	mimeType := http.DetectContentType(prefix[:n])
	if directory == "videos" && !validVideoHeader(ext, prefix[:n]) {
		return "", "", 0, errors.New("invalid video content")
	}
	if directory != "videos" && !strings.HasPrefix(mimeType, "image/") {
		return "", "", 0, errors.New("invalid image content")
	}
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		return "", "", 0, err
	}
	var random [16]byte
	if _, err := rand.Read(random[:]); err != nil {
		return "", "", 0, err
	}
	filename := hex.EncodeToString(random[:]) + ext
	targetDir := filepath.Join(h.Config.StorageDir, directory)
	if err := os.MkdirAll(targetDir, 0o750); err != nil {
		return "", "", 0, err
	}
	targetPath := filepath.Join(targetDir, filename)
	target, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return "", "", 0, err
	}
	written, copyErr := io.Copy(target, io.LimitReader(source, limit+1))
	closeErr := target.Close()
	if copyErr != nil || closeErr != nil || written > limit {
		_ = os.Remove(targetPath)
		if written > limit {
			return "", "", 0, errUploadTooLarge
		}
		if copyErr != nil {
			return "", "", 0, copyErr
		}
		return "", "", 0, closeErr
	}
	return "/media/" + directory + "/" + filename, mimeType, written, nil
}

func validVideoHeader(ext string, header []byte) bool {
	switch ext {
	case ".mp4", ".mov":
		return len(header) >= 12 && string(header[4:8]) == "ftyp"
	case ".webm", ".mkv":
		return len(header) >= 4 && header[0] == 0x1a && header[1] == 0x45 && header[2] == 0xdf && header[3] == 0xa3
	default:
		return false
	}
}

func writeUploadError(c *gin.Context, err error) {
	if errors.Is(err, errUploadTooLarge) || strings.Contains(strings.ToLower(err.Error()), "request body too large") {
		response.Error(c, http.StatusRequestEntityTooLarge, 41300, "upload exceeds size limit")
		return
	}
	response.Error(c, http.StatusBadRequest, 40010, "invalid upload")
}
