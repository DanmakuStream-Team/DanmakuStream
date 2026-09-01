package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrNotFound    = errors.New("content object not found")
	ErrUnavailable = errors.New("content service unavailable")
	ErrTimeout     = errors.New("content service timeout")
	ErrBadGateway  = errors.New("content service invalid response")
)

type ContentClient struct {
	baseURL string
	token   string
	http    *http.Client
}

type VideoSummary struct {
	ID        uint   `json:"id"`
	CreatorID uint   `json:"creatorId"`
	Title     string `json:"title"`
	CoverURL  string `json:"coverUrl"`
	Duration  int    `json:"duration"`
	Status    string `json:"status"`
	Playable  bool   `json:"playable"`
}

type envelope[T any] struct {
	Code int `json:"code"`
	Data T   `json:"data"`
}

func NewContent(baseURL, token string, timeout time.Duration) *ContentClient {
	return &ContentClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: timeout},
	}
}

func (c *ContentClient) Video(ctx context.Context, id uint, requestID string) (VideoSummary, error) {
	var out envelope[VideoSummary]
	err := c.get(ctx, "/internal/v1/videos/"+strconv.FormatUint(uint64(id), 10), requestID, &out)
	if err == nil && out.Code != 0 {
		err = fmt.Errorf("%w: business code %d", ErrBadGateway, out.Code)
	}
	return out.Data, err
}

func (c *ContentClient) Videos(ctx context.Context, ids []uint, requestID string) ([]VideoSummary, error) {
	if len(ids) == 0 {
		return []VideoSummary{}, nil
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatUint(uint64(id), 10))
	}
	var out envelope[struct {
		Items []VideoSummary `json:"items"`
	}]
	err := c.get(ctx, "/internal/v1/videos/batch?ids="+url.QueryEscape(strings.Join(parts, ",")), requestID, &out)
	if err == nil && out.Code != 0 {
		err = fmt.Errorf("%w: business code %d", ErrBadGateway, out.Code)
	}
	return out.Data.Items, err
}

func (c *ContentClient) get(ctx context.Context, path, requestID string, target any) error {
	if c == nil || c.baseURL == "" {
		return fmt.Errorf("%w: CONTENT_SERVICE_URL not configured", ErrUnavailable)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrBadGateway, err)
	}
	req.Header.Set("X-Internal-Token", c.token)
	if requestID != "" {
		req.Header.Set("X-Request-ID", requestID)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		var netErr net.Error
		if errors.As(err, &netErr) && netErr.Timeout() {
			return fmt.Errorf("%w: %v", ErrTimeout, err)
		}
		return fmt.Errorf("%w: %v", ErrUnavailable, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return ErrNotFound
	}
	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("%w: internal credential rejected", ErrBadGateway)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: status %d", ErrUnavailable, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("%w: %v", ErrBadGateway, err)
	}
	return nil
}
