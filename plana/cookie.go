package plana

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type CookieService service

var ErrCookieURLNotConfigured = errors.New("GetCookieURL is not configured in the client")

type GetCookieOptions struct {
	UserID    string `json:"user_id"`
	Seed      string `json:"seed"`
	AuthToken string `json:"-"`
}

func (c *CookieService) GetCookie(ctx context.Context, opts GetCookieOptions) (string, error) {
	if c.client.GetCookieURL == nil {
		return "", ErrCookieURLNotConfigured
	}

	u, err := c.client.GetCookieURL.Parse("/cookie")
	if err != nil {
		return "", fmt.Errorf("failed to parse cookie URL: %w", err)
	}

	payload, err := c.client.JSONSerializer.Serialize(opts, "")
	if err != nil {
		return "", fmt.Errorf("failed to serialize protocol encoder request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create protocol encoder request: %w", err)
	}

	// Add authentication header if provided
	if opts.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+opts.AuthToken)
	}

	resp, err := c.client.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send protocol encoder request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	cookie, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read protocol encoder response: %w", err)
	}

	return string(cookie), nil
}
