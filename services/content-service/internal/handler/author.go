package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"danmakustream/content-service/internal/logic"
	"github.com/gin-gonic/gin"
)

type userSummaryEnvelope struct {
	Data struct {
		Items []logic.AuthorDTO `json:"items"`
	} `json:"data"`
}

// enrichVideos resolves user-owned author data through user-service. Content
// remains available when that dependency is down, but the response carries an
// explicit, stable fallback instead of empty fields that the UI labels anonymous.
func (h *Handler) enrichVideos(c *gin.Context, videos []logic.VideoDTO) {
	if len(videos) == 0 {
		return
	}
	ids := make(map[uint]struct{}, len(videos))
	for _, video := range videos {
		if video.Author.ID > 0 {
			ids[video.Author.ID] = struct{}{}
		}
	}
	authors, err := h.fetchAuthors(c.Request.Context(), c.GetHeader("X-Request-ID"), ids)
	for i := range videos {
		id := videos[i].Author.ID
		if author, ok := authors[id]; err == nil && ok {
			videos[i].Author = author
			continue
		}
		videos[i].Author = unavailableAuthor(id)
	}
}

func (h *Handler) fetchAuthors(ctx context.Context, requestID string, ids map[uint]struct{}) (map[uint]logic.AuthorDTO, error) {
	query := url.Values{}
	for id := range ids {
		query.Add("id", strconv.FormatUint(uint64(id), 10))
	}
	endpoint := strings.TrimRight(h.Config.UserServiceURL, "/") + "/internal/v1/users?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Internal-Token", h.Config.InternalAPIToken)
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}
	resp, err := (&http.Client{Timeout: h.Config.RequestTimeout}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("user-service returned %s", resp.Status)
	}
	var envelope userSummaryEnvelope
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		return nil, err
	}
	result := make(map[uint]logic.AuthorDTO, len(envelope.Data.Items))
	for _, author := range envelope.Data.Items {
		result[author.ID] = author
	}
	return result, nil
}

func unavailableAuthor(id uint) logic.AuthorDTO {
	return logic.AuthorDTO{
		ID:       id,
		Username: fmt.Sprintf("user-%d", id),
		Nickname: fmt.Sprintf("用户 #%d（资料暂不可用）", id),
	}
}
