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

var ErrUnavailable = errors.New("dependency unavailable")
var ErrNotFound = errors.New("dependency object not found")
var ErrTimeout = errors.New("dependency timeout")
var ErrBadGateway = errors.New("dependency invalid response")

type Client struct {
	baseURL, token string
	http           *http.Client
}

type requestIDKey struct{}

func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func New(baseURL, token string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: &http.Client{Timeout: timeout}}
}

type envelope[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

func (c *Client) get(ctx context.Context, path string, target any) error {
	if c.baseURL == "" {
		return fmt.Errorf("%w: base URL not configured", ErrUnavailable)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("X-Internal-Token", c.token)
	}
	if requestID, _ := ctx.Value(requestIDKey{}).(string); requestID != "" {
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("%w: status %d", ErrUnavailable, resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("%w: %v", ErrBadGateway, err)
	}
	return nil
}

type UserSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

func (c *Client) User(ctx context.Context, id uint) (UserSummary, error) {
	var out envelope[UserSummary]
	err := c.get(ctx, "/internal/v1/users/"+fmt.Sprint(id), &out)
	return out.Data, err
}
func (c *Client) IsBlocked(ctx context.Context, blockerID, blockedID uint) (bool, error) {
	var out envelope[struct {
		Blocked bool `json:"blocked"`
	}]
	err := c.get(ctx, fmt.Sprintf("/internal/v1/relationships/blocked?blockerId=%d&blockedId=%d", blockerID, blockedID), &out)
	return out.Data.Blocked, err
}
func (c *Client) IsMember(ctx context.Context, userID, creatorID uint) (bool, error) {
	var out envelope[struct {
		Active bool `json:"active"`
	}]
	err := c.get(ctx, fmt.Sprintf("/internal/v1/memberships/status?userId=%d&creatorId=%d", userID, creatorID), &out)
	return out.Data.Active, err
}
func (c *Client) IsFollowing(ctx context.Context, followerID, followeeID uint) (bool, error) {
	var out envelope[struct {
		Following bool `json:"following"`
	}]
	err := c.get(ctx, fmt.Sprintf("/internal/v1/relationships/following?followerId=%d&followeeId=%d", followerID, followeeID), &out)
	return out.Data.Following, err
}

type VideoSummary struct {
	ID        uint   `json:"id"`
	CreatorID uint   `json:"creatorId"`
	Title     string `json:"title"`
	CoverURL  string `json:"coverUrl"`
	VideoURL  string `json:"videoUrl"`
	Duration  int    `json:"duration"`
	Status    string `json:"status"`
	Playable  bool   `json:"playable"`
}

func (c *Client) Videos(ctx context.Context, ids []uint) ([]VideoSummary, error) {
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
	err := c.get(ctx, "/internal/v1/videos/batch?ids="+url.QueryEscape(strings.Join(parts, ",")), &out)
	return out.Data.Items, err
}

func (c *Client) Video(ctx context.Context, id uint) (VideoSummary, error) {
	var out envelope[VideoSummary]
	err := c.get(ctx, "/internal/v1/videos/"+url.PathEscape(fmt.Sprint(id)), &out)
	return out.Data, err
}
