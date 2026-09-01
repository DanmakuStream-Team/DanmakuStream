package handler

import (
	"net/http"
	"strconv"
	"strings"

	"danmakustream/content-service/internal/model"
	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type internalVideoSummary struct {
	ID        uint   `json:"id"`
	CreatorID uint   `json:"creatorId"`
	Title     string `json:"title"`
	CoverURL  string `json:"coverUrl"`
	Duration  int    `json:"duration"`
	Status    string `json:"status"`
	Playable  bool   `json:"playable"`
}

func toInternalVideoSummary(video model.Video) internalVideoSummary {
	return internalVideoSummary{
		ID: video.ID, CreatorID: video.AuthorID, Title: video.Title,
		CoverURL: video.CoverURL, Duration: video.Duration, Status: video.Status,
		Playable: video.Status == "approved" && video.TranscodeStatus == "ready",
	}
}

func (h *Handler) InternalVideo(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	var video model.Video
	if err := h.DB.First(&video, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, http.StatusNotFound, 40401, "video not found")
			return
		}
		writeLogicError(c, err)
		return
	}
	response.OK(c, toInternalVideoSummary(video))
}

func (h *Handler) InternalVideos(c *gin.Context) {
	raw := strings.Split(c.Query("ids"), ",")
	if len(raw) == 0 || len(raw) > 100 {
		response.Error(c, http.StatusBadRequest, 40006, "invalid ids")
		return
	}
	ids := make([]uint, 0, len(raw))
	seen := make(map[uint]struct{}, len(raw))
	for _, value := range raw {
		parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
		id := uint(parsed)
		if err != nil || id == 0 {
			response.Error(c, http.StatusBadRequest, 40006, "invalid ids")
			return
		}
		if _, exists := seen[id]; !exists {
			seen[id] = struct{}{}
			ids = append(ids, id)
		}
	}
	var videos []model.Video
	if err := h.DB.Where("id IN ?", ids).Find(&videos).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	byID := make(map[uint]model.Video, len(videos))
	for _, video := range videos {
		byID[video.ID] = video
	}
	items := make([]internalVideoSummary, 0, len(videos))
	for _, id := range ids {
		if video, exists := byID[id]; exists {
			items = append(items, toInternalVideoSummary(video))
		}
	}
	response.OK(c, gin.H{"items": items})
}
