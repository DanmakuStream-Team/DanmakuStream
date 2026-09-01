package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	if timeout <= 0 || timeout > 2*time.Second {
		timeout = 2 * time.Second
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: timeout, KeepAlive: 30 * time.Second}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          20,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       60 * time.Second,
		TLSHandshakeTimeout:   timeout,
		ResponseHeaderTimeout: timeout,
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: &http.Client{Timeout: timeout, Transport: transport}}
}

type envelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
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
	req.Header.Set("Accept", "application/json")
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
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return dependencyHTTPError(resp)
	}
	var body envelope
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 2<<20))
	if err := decoder.Decode(&body); err != nil {
		return fmt.Errorf("%w: %v", ErrBadGateway, err)
	}
	if body.Code != 0 {
		status := body.Code
		if status >= 1000 {
			status /= 100
		}
		return fmt.Errorf("%w: downstream code %d: %s", errorForStatus(status), body.Code, body.Message)
	}
	if target == nil {
		return nil
	}
	if len(body.Data) == 0 || string(body.Data) == "null" {
		return fmt.Errorf("%w: missing data", ErrBadGateway)
	}
	if err := json.Unmarshal(body.Data, target); err != nil {
		return fmt.Errorf("%w: %v", ErrBadGateway, err)
	}
	return nil
}

func dependencyHTTPError(resp *http.Response) error {
	mapped := errorForStatus(resp.StatusCode)
	var body envelope
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64<<10)).Decode(&body); err == nil {
		// A missing internal route is an integration contract failure, not a
		// missing business object. Content-service currently reserves 40400 for
		// its router-level "route not found" response.
		if resp.StatusCode == http.StatusNotFound && body.Code == 40400 && strings.Contains(strings.ToLower(body.Message), "route") {
			mapped = ErrBadGateway
		}
		return fmt.Errorf("%w: status %d downstream code %d", mapped, resp.StatusCode, body.Code)
	}
	return fmt.Errorf("%w: status %d", mapped, resp.StatusCode)
}

func errorForStatus(status int) error {
	switch status {
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusGatewayTimeout:
		return ErrTimeout
	case http.StatusBadGateway, http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden:
		return ErrBadGateway
	default:
		return ErrUnavailable
	}
}

type UserSummary struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     string `json:"role"`
}

func (c *Client) User(ctx context.Context, id uint) (UserSummary, error) {
	var out struct {
		Items []UserSummary `json:"items"`
	}
	query := url.Values{}
	query.Add("id", strconv.FormatUint(uint64(id), 10))
	if err := c.get(ctx, "/internal/v1/users?"+query.Encode(), &out); err != nil {
		return UserSummary{}, err
	}
	for _, user := range out.Items {
		if user.ID == id {
			return user, nil
		}
	}
	return UserSummary{}, ErrNotFound
}
func (c *Client) IsBlocked(ctx context.Context, blockerID, blockedID uint) (bool, error) {
	var out struct {
		Blocked bool `json:"blocked"`
	}
	query := url.Values{}
	query.Set("firstId", strconv.FormatUint(uint64(blockerID), 10))
	query.Set("secondId", strconv.FormatUint(uint64(blockedID), 10))
	err := c.get(ctx, "/internal/v1/relationships/blocked?"+query.Encode(), &out)
	return out.Blocked, err
}
func (c *Client) IsMember(ctx context.Context, userID, creatorID uint) (bool, error) {
	var out struct {
		Active bool `json:"active"`
	}
	query := url.Values{}
	query.Set("subscriberId", strconv.FormatUint(uint64(userID), 10))
	query.Set("creatorId", strconv.FormatUint(uint64(creatorID), 10))
	err := c.get(ctx, "/internal/v1/memberships/status?"+query.Encode(), &out)
	return out.Active, err
}
func (c *Client) IsFollowing(ctx context.Context, followerID, followeeID uint) (bool, error) {
	var out struct {
		Following bool `json:"following"`
	}
	err := c.get(ctx, fmt.Sprintf("/internal/v1/relationships/following?followerId=%d&followeeId=%d", followerID, followeeID), &out)
	return out.Following, err
}

type VideoSummary struct {
	ID              uint   `json:"id"`
	CreatorID       uint   `json:"creatorId"`
	Title           string `json:"title"`
	CoverURL        string `json:"coverUrl"`
	VideoURL        string `json:"videoUrl"`
	Duration        int    `json:"duration"`
	Status          string `json:"status"`
	TranscodeStatus string `json:"transcodeStatus,omitempty"`
	Playable        bool   `json:"playable"`
}

func (v *VideoSummary) UnmarshalJSON(data []byte) error {
	var wire struct {
		ID              uint   `json:"id"`
		CreatorID       uint   `json:"creatorId"`
		AuthorID        uint   `json:"authorId"`
		Title           string `json:"title"`
		CoverURL        string `json:"coverUrl"`
		VideoURL        string `json:"videoUrl"`
		Duration        int    `json:"duration"`
		Status          string `json:"status"`
		TranscodeStatus string `json:"transcodeStatus"`
		Playable        *bool  `json:"playable"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		return err
	}
	creatorID := wire.CreatorID
	if creatorID == 0 {
		creatorID = wire.AuthorID
	}
	playable := wire.Status == "approved" && (wire.TranscodeStatus == "" || wire.TranscodeStatus == "ready")
	if wire.Playable != nil {
		playable = *wire.Playable
	}
	*v = VideoSummary{ID: wire.ID, CreatorID: creatorID, Title: wire.Title, CoverURL: wire.CoverURL, VideoURL: wire.VideoURL, Duration: wire.Duration, Status: wire.Status, TranscodeStatus: wire.TranscodeStatus, Playable: playable}
	return nil
}

func (c *Client) Videos(ctx context.Context, ids []uint) ([]VideoSummary, error) {
	if len(ids) == 0 {
		return []VideoSummary{}, nil
	}
	parts := make([]string, 0, len(ids))
	for _, id := range ids {
		parts = append(parts, strconv.FormatUint(uint64(id), 10))
	}
	var out struct {
		Items []VideoSummary `json:"items"`
	}
	err := c.get(ctx, "/internal/v1/videos/batch?ids="+url.QueryEscape(strings.Join(parts, ",")), &out)
	return out.Items, err
}

func (c *Client) Video(ctx context.Context, id uint) (VideoSummary, error) {
	var out VideoSummary
	err := c.get(ctx, "/internal/v1/videos/"+url.PathEscape(fmt.Sprint(id)), &out)
	return out, err
}
