package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Client handles HTTP communication with the now.ctx.st API.
type Client struct {
	Endpoint   string
	Token      string
	HTTPClient *http.Client
}

// NewClient creates a new API client.
func NewClient(endpoint, token string) *Client {
	return &Client{
		Endpoint: endpoint,
		Token:    token,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// StatusRequest is the body for POST /api/status.
type StatusRequest struct {
	Content string `json:"content"`
	Emoji   string `json:"emoji,omitempty"`
}

// MeResponse is the response from GET /api/auth/me.
type MeResponse struct {
	User struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}

// BoardEntry represents a user on the board.
type BoardEntry struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	Emoji      string `json:"emoji"`
	LastSeenAt string `json:"lastSeenAt"`
}

// BoardResponse is the response from GET /api/status.
type BoardResponse struct {
	Board []BoardEntry `json:"board"`
}

// PushStatus sends a status update.
func (c *Client) PushStatus(content, emoji string) error {
	body := StatusRequest{Content: content, Emoji: emoji}
	_, err := c.doJSON("POST", "/api/status", body)
	return err
}

// VerifyToken checks if the token is valid by calling /api/auth/me.
func (c *Client) VerifyToken() (*MeResponse, error) {
	data, err := c.doJSON("GET", "/api/auth/me", nil)
	if err != nil {
		return nil, err
	}
	var me MeResponse
	if err := json.Unmarshal(data, &me); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &me, nil
}

// GetBoard fetches the current board.
func (c *Client) GetBoard() (*BoardResponse, error) {
	data, err := c.doJSON("GET", "/api/status", nil)
	if err != nil {
		return nil, err
	}
	var board BoardResponse
	if err := json.Unmarshal(data, &board); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &board, nil
}

func (c *Client) doJSON(method, path string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.Endpoint+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	// Handle rate limiting
	if resp.StatusCode == 429 {
		retryAfter := resp.Header.Get("Retry-After")
		secs, _ := strconv.Atoi(retryAfter)
		if secs <= 0 {
			secs = 5
		}
		return nil, &RateLimitError{RetryAfter: time.Duration(secs) * time.Second}
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// RateLimitError indicates the API returned 429.
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limited, retry after %s", e.RetryAfter)
}

// AuthPendingError indicates the device code is not yet confirmed.
type AuthPendingError struct{}

func (e *AuthPendingError) Error() string {
	return "authorization pending"
}

// DeviceCodeResponse is the response from POST /api/auth/device.
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// DeviceTokenResponse is the response from POST /api/auth/device/token.
type DeviceTokenResponse struct {
	Token string `json:"token"`
	User  struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"user"`
}

// RequestDeviceCode initiates device flow authentication.
func (c *Client) RequestDeviceCode() (*DeviceCodeResponse, error) {
	data, err := c.doJSON("POST", "/api/auth/device", struct{}{})
	if err != nil {
		return nil, err
	}
	var resp DeviceCodeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &resp, nil
}

// PollDeviceToken polls for a device token. Returns AuthPendingError on 428.
func (c *Client) PollDeviceToken(deviceCode string) (*DeviceTokenResponse, error) {
	body := struct {
		DeviceCode string `json:"device_code"`
	}{DeviceCode: deviceCode}

	reqData, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", c.Endpoint+"/api/auth/device/token", bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode == 428 {
		return nil, &AuthPendingError{}
	}

	if resp.StatusCode == 429 {
		retryAfter := resp.Header.Get("Retry-After")
		secs, _ := strconv.Atoi(retryAfter)
		if secs <= 0 {
			secs = 5
		}
		return nil, &RateLimitError{RetryAfter: time.Duration(secs) * time.Second}
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if json.Unmarshal(respBody, &errResp) == nil && errResp.Error != "" {
			return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, errResp.Error)
		}
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var tokenResp DeviceTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}
	return &tokenResp, nil
}
