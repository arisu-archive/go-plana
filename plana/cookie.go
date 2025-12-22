package plana

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type CookieService service

var ErrCookieURLNotConfigured = errors.New("GetCookieURL is not configured in the client")

type GetCookieOptions struct {
	UserID    string `json:"user_id"`
	Seed      string `json:"seed"`
	AuthToken string `json:"-"`
}

type Cookie struct {
	Success   bool      `json:"success"`
	Cookie    string    `json:"cookie"`
	Timestamp time.Time `json:"timestamp"`
}

func (c *CookieService) GetCookie(ctx context.Context, opts GetCookieOptions) (*Cookie, error) {
	if c.client.GetCookieURL == nil {
		return nil, ErrCookieURLNotConfigured
	}

	u, err := c.client.GetCookieURL.Parse("/cookie")
	if err != nil {
		return nil, fmt.Errorf("failed to parse cookie URL: %w", err)
	}

	payload, err := c.client.JSONSerializer.Serialize(opts, "")
	if err != nil {
		return nil, fmt.Errorf("failed to serialize protocol encoder request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create protocol encoder request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Add authentication header if provided
	if opts.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+opts.AuthToken)
	}

	resp, err := c.client.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send protocol encoder request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var out Cookie
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode cookie response as JSON: %w", err)
	}
	return &out, nil
}
