package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"danmakustream/content-service/internal/middleware"
	"danmakustream/content-service/internal/model"
	"danmakustream/content-service/internal/response"
	"github.com/gin-gonic/gin"
)

func (h *Handler) ListDynamics(c *gin.Context) {
	page, pageSize := pageParams(c)
	var userID *uint
	if value := c.Query("userId"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 64)
		if err != nil || parsed == 0 {
			response.Error(c, http.StatusBadRequest, 40000, "invalid userId")
			return
		}
		id := uint(parsed)
		userID = &id
	}
	result, err := h.Logic.ListDynamics(page, pageSize, userID)
	if err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, result)
}

type dynamicInput struct {
	Content string          `json:"content" binding:"required"`
	Images  json.RawMessage `json:"images"`
}

func (h *Handler) CreateDynamic(c *gin.Context) {
	var input dynamicInput
	if err := c.ShouldBindJSON(&input); err != nil || strings.TrimSpace(input.Content) == "" {
		response.Error(c, http.StatusBadRequest, 40001, "invalid request body")
		return
	}
	encoded, err := normalizeImages(input.Images)
	if err != nil || len(encoded) > 1000 {
		response.Error(c, http.StatusBadRequest, 40004, "invalid images")
		return
	}
	post := model.DynamicPost{UserID: middleware.UserID(c), Content: strings.TrimSpace(input.Content), Images: encoded}
	if err := h.DB.Create(&post).Error; err != nil {
		writeLogicError(c, err)
		return
	}
	if err := h.notifyFollowers(c, post); err != nil {
		response.Error(c, http.StatusBadGateway, 50201, "notification service unavailable")
		return
	}
	response.Created(c, post)
}

func (h *Handler) notifyFollowers(c *gin.Context, post model.DynamicPost) error {
	payload, _ := json.Marshal(gin.H{"type": "dynamic", "title": "你关注的用户发布了新动态", "content": post.Content, "link": fmt.Sprintf("/user/%d", post.UserID)})
	ctx, cancel := context.WithTimeout(c.Request.Context(), h.Config.RequestTimeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/internal/v1/users/%d/followers/notifications", strings.TrimRight(h.Config.UserServiceURL, "/"), post.UserID), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Token", h.Config.InternalAPIToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("notification service status %d", resp.StatusCode)
	}
	return nil
}

func (h *Handler) DeleteDynamic(c *gin.Context) {
	id, ok := uintParam(c, "id")
	if !ok {
		return
	}
	if err := h.Logic.DeleteDynamic(id, middleware.UserID(c), c.GetString(middleware.ContextRole) == "admin"); err != nil {
		writeLogicError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": true})
}

func normalizeImages(raw json.RawMessage) (string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return "[]", nil
	}
	var encoded string
	if err := json.Unmarshal(raw, &encoded); err == nil {
		var values []string
		if encoded == "" {
			return "[]", nil
		}
		if err := json.Unmarshal([]byte(encoded), &values); err != nil {
			return "", err
		}
		canonical, err := json.Marshal(values)
		return string(canonical), err
	}
	var values []string
	if err := json.Unmarshal(raw, &values); err != nil {
		return "", err
	}
	canonical, err := json.Marshal(values)
	return string(canonical), err
}
