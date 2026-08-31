package handler

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"danmakustream/content-service/internal/logic"
	"danmakustream/content-service/internal/middleware"
	"danmakustream/content-service/internal/model"
	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListVideos(c *gin.Context) {
	page, pageSize := pageParams(c)
	result, err := h.Logic.ListVideos(logic.VideoListOptions{
		Page: page, PageSize: pageSize, Keyword: c.Query("keyword"), Tag: c.Query("tag"),
		Category: c.Query("category"), Sort: c.Query("sort"),
	})
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) UserVideos(c *gin.Context) {
	userID, ok := uintParam(c, "id")
	if !ok {
		return
	}
	page, pageSize := pageParams(c)
	result, err := h.Logic.ListVideos(logic.VideoListOptions{Page: page, PageSize: pageSize, AuthorID: &userID, Sort: c.Query("sort")})
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) MyVideos(c *gin.Context) {
	userID := middleware.UserID(c)
	page, pageSize := pageParams(c)
	result, err := h.Logic.ListVideos(logic.VideoListOptions{Page: page, PageSize: pageSize, AuthorID: &userID, IncludeUnapproved: true, Status: c.Query("status"), Sort: c.Query("sort")})
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, result)
}

func (h *Handler) VideoDetail(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	video, err := h.Logic.GetVideo(id, false)
	if err != nil {
		writeLogicError(c, err)
		return
	}
	if err := h.Logic.RecordView(&video); err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, logic.VideoView(video))
}

type updateVideoInput struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Tags        *string `json:"tags"`
	Category    *string `json:"category"`
}

func (h *Handler) UpdateVideo(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	var input updateVideoInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	fields := map[string]any{}
	if input.Title != nil {
		title := strings.TrimSpace(*input.Title)
		if title == "" || len(title) > 200 {
			response.Error(c, http.StatusBadRequest, 40002, "invalid title")
			return
		}
		fields["title"] = title
	}
	if input.Description != nil {
		fields["description"] = *input.Description
	}
	if input.Tags != nil {
		fields["tags"] = *input.Tags
	}
	if input.Category != nil {
		fields["category"] = *input.Category
	}
	if len(fields) == 0 {
		response.Error(c, http.StatusBadRequest, 40003, "no fields to update")
		return
	}
	video, err := h.Logic.UpdateVideo(id, middleware.UserID(c), fields)
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, logic.VideoView(video))
}

func (h *Handler) DeleteVideo(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	var assets []model.MediaAsset
	if err := h.DB.Where("video_id = ?", id).Find(&assets).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	if err := h.Logic.DeleteVideo(id, middleware.UserID(c)); err != nil {
		writeLogicError(c, err)
		return
	}
	for _, asset := range assets {
		removeMedia(h.Config.StorageDir, asset.Path)
	}
	response.OK(c, gin.H{"deleted": true})
}

type collaboratorInput struct {
	UserID uint `json:"userId" binding:"required"`
}

func (h *Handler) AddCollaborator(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	var input collaboratorInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	item, err := h.Logic.AddCollaborator(id, middleware.UserID(c), input.UserID)
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *Handler) RemoveCollaborator(c *gin.Context) {
	videoID, ok := uintParam(c, "id")
	if !ok {
		return
	}
	userID, ok := uintParam(c, "userId")
	if !ok {
		return
	}
	if err := h.Logic.RemoveCollaborator(videoID, middleware.UserID(c), userID); err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) DownloadVideo(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	video, err := h.Logic.GetVideo(id, true)
	if err != nil {
		writeLogicError(c, err)
		return
	}
	allowed, err := h.Logic.CanEditVideo(id, middleware.UserID(c))
	if err != nil {
		writeLogicError(c, err)
		return
	}
	role := c.GetString(middleware.ContextRole)
	if !allowed && role != "admin" && role != "moderator" && video.Status != "approved" {
		writeLogicError(c, logic.ErrForbidden)
		return
	}
	path, err := localMediaPath(h.Config.StorageDir, video.VideoURL)
	if err != nil {
		writeLogicError(c, logic.ErrNotFound)
		return
	}
	if _, err := os.Stat(path); err != nil {
		writeLogicError(c, logic.ErrNotFound)
		return
	}
	c.FileAttachment(path, filepath.Base(path))
}

func localMediaPath(root, mediaURL string) (string, error) {
	relative := strings.TrimPrefix(mediaURL, "/media/")
	clean := filepath.Clean(relative)
	if clean == "." || strings.HasPrefix(clean, "..") || filepath.IsAbs(clean) {
		return "", errors.New("invalid media path")
	}
	return filepath.Join(root, clean), nil
}

func (h *Handler) AdminListVideos(c *gin.Context) {
	page, pageSize := pageParams(c)
	result, err := h.Logic.ListVideos(logic.VideoListOptions{
		Page: page, PageSize: pageSize, IncludeUnapproved: true, Status: c.Query("status"),
		Keyword: c.Query("keyword"), Sort: c.Query("sort"),
	})
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, result)
}

type reviewInput struct {
	Status string `json:"status" binding:"required"`
}

func (h *Handler) ReviewVideo(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	var input reviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	video, err := h.Logic.ReviewVideo(id, input.Status)
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, logic.VideoView(video))
}

func (h *Handler) CreatorAnalytics(c *gin.Context) {
	result, err := h.Logic.CreatorAnalytics(middleware.UserID(c))
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, result)
}

func parseUint(value string) uint {
	parsed, _ := strconv.ParseUint(value, 10, 64)
	return uint(parsed)
}
