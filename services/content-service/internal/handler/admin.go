package handler

import (
	"errors"
	"net/http"
	"time"

	"danmakustream/content-service/internal/logic"
	"danmakustream/content-service/internal/model"
	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func (h *Handler) PublicBanners(c *gin.Context) {
	var items []model.SiteBanner
	if err := h.DB.Where("enabled = ?", true).Order("sort ASC, id DESC").Find(&items).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *Handler) PublicAnnouncements(c *gin.Context) {
	var items []model.SiteAnnouncement
	if err := h.DB.Where("enabled = ? AND (started_at IS NULL OR started_at <= CURRENT_TIMESTAMP) AND (ended_at IS NULL OR ended_at >= CURRENT_TIMESTAMP)").Order("id DESC").Find(&items).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, items)
}

func (h *Handler) AdminBanners(c *gin.Context) {
	var items []model.SiteBanner
	if err := h.DB.Order("sort ASC, id DESC").Find(&items).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, items)
}

type bannerInput struct {
	Title    *string `json:"title"`
	ImageURL *string `json:"imageUrl"`
	Link     *string `json:"link"`
	Enabled  *bool   `json:"enabled"`
	Sort     *int    `json:"sort"`
}

func (h *Handler) CreateBanner(c *gin.Context) {
	var input bannerInput
	if err := c.ShouldBindJSON(&input); err != nil || input.Title == nil || *input.Title == "" {
		response.Error(c, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	item := model.SiteBanner{Title: *input.Title, Enabled: true}
	applyBannerInput(&item, input)
	if err := h.DB.Create(&item).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *Handler) UpdateBanner(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	var input bannerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	var item model.SiteBanner
	if err := h.DB.First(&item, id).Error; err != nil {
		writeDBNotFound(c, err)
		return
	}
	applyBannerInput(&item, input)
	if err := h.DB.Save(&item).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, item)
}

func applyBannerInput(item *model.SiteBanner, input bannerInput) {
	if input.Title != nil {
		item.Title = *input.Title
	}
	if input.ImageURL != nil {
		item.ImageURL = *input.ImageURL
	}
	if input.Link != nil {
		item.Link = *input.Link
	}
	if input.Enabled != nil {
		item.Enabled = *input.Enabled
	}
	if input.Sort != nil {
		item.Sort = *input.Sort
	}
}

func (h *Handler) DeleteBanner(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	result := h.DB.Delete(&model.SiteBanner{}, id)
	if result.Error != nil {
		writeLogicError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		writeLogicError(c, logic.ErrNotFound)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func (h *Handler) AdminAnnouncements(c *gin.Context) {
	var items []model.SiteAnnouncement
	if err := h.DB.Order("id DESC").Find(&items).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, items)
}

type announcementInput struct {
	Content   *string    `json:"content"`
	Enabled   *bool      `json:"enabled"`
	StartedAt *time.Time `json:"startedAt"`
	EndedAt   *time.Time `json:"endedAt"`
}

func (h *Handler) CreateAnnouncement(c *gin.Context) {
	var input announcementInput
	if err := c.ShouldBindJSON(&input); err != nil || input.Content == nil || *input.Content == "" {
		response.Error(c, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	item := model.SiteAnnouncement{Content: *input.Content, Enabled: true, StartedAt: input.StartedAt, EndedAt: input.EndedAt}
	if input.Enabled != nil {
		item.Enabled = *input.Enabled
	}
	if invalidWindow(item.StartedAt, item.EndedAt) {
		response.Error(c, http.StatusBadRequest, 40005, "invalid active window")
		return
	}
	if err := h.DB.Create(&item).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	response.Created(c, item)
}

func (h *Handler) UpdateAnnouncement(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	var input announcementInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	var item model.SiteAnnouncement
	if err := h.DB.First(&item, id).Error; err != nil {
		writeDBNotFound(c, err)
		return
	}
	if input.Content != nil {
		item.Content = *input.Content
	}
	if input.Enabled != nil {
		item.Enabled = *input.Enabled
	}
	if input.StartedAt != nil {
		item.StartedAt = input.StartedAt
	}
	if input.EndedAt != nil {
		item.EndedAt = input.EndedAt
	}
	if invalidWindow(item.StartedAt, item.EndedAt) {
		response.Error(c, http.StatusBadRequest, 40005, "invalid active window")
		return
	}
	if err := h.DB.Save(&item).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, item)
}

func (h *Handler) DeleteAnnouncement(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	result := h.DB.Delete(&model.SiteAnnouncement{}, id)
	if result.Error != nil {
		writeLogicError(c, result.Error)
		return
	}
	if result.RowsAffected == 0 {
		writeLogicError(c, logic.ErrNotFound)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func invalidWindow(startedAt, endedAt *time.Time) bool {
	return startedAt != nil && endedAt != nil && endedAt.Before(*startedAt)
}

func writeDBNotFound(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		writeLogicError(c, logic.ErrNotFound)
		return
	}
	writeLogicError(c, err)
}
