package plana

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type GetCookieRequest struct {
	Seed string `json:"seed"`
}

type CookieService service

var ErrCookieURLNotConfigured = errors.New("GetCookieURL is not configured in the client")

func (c *CookieService) GetCookie(ctx context.Context, serverSeed string) (string, error) {
	if c.client.GetCookieURL == nil {
		return "", ErrCookieURLNotConfigured
	}

	data := GetCookieRequest{Seed: serverSeed}
	payload, err := c.client.JSONSerializer.Serialize(data, "")
	if err != nil {
		return "", fmt.Errorf("failed to serialize protocol encoder request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.client.GetCookieURL.String(), bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create protocol encoder request: %w", err)
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
