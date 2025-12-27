package plana

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/arisu-archive/plana-protos/protos"
)

var ErrProtocolEncoderNotConfigured = errors.New("protocol encoder is not configured")

type encoderRequest struct {
	Server   string `json:"server"`
	Protocol uint64 `json:"protocol"`
	Crc32    uint64 `json:"crc32"`
}

func (c *Client) encodeProtocol(ctx context.Context, crc32 uint32, p protos.Protocol) (uint32, error) {
	if c.ProtocolEncoderConfig == nil {
		return 0, ErrProtocolEncoderNotConfigured
	}
	data := encoderRequest{
		Server:   "japan",
		Protocol: uint64(p), //nolint:gosec // This is how the protocol works
		Crc32:    uint64(crc32),
	}
	payload, err := c.JSONSerializer.Serialize(data, "")
	if err != nil {
		return 0, fmt.Errorf("failed to serialize protocol encoder request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.ProtocolEncoderConfig.URL.String(), bytes.NewBuffer(payload))
	if err != nil {
		return 0, fmt.Errorf("failed to create protocol encoder request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.ProtocolEncoderConfig.ClientID != "" && c.ProtocolEncoderConfig.ClientSecret != "" {
		req.Header.Set("Cf-Access-Client-Id", c.ProtocolEncoderConfig.ClientID)
		req.Header.Set("Cf-Access-Client-Secret", c.ProtocolEncoderConfig.ClientSecret)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to send protocol encoder request: %w", err)
	}
	defer resp.Body.Close()

	var encoderResp struct {
		Result uint64 `json:"result"`
	}
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read protocol encoder response: %w", err)
	}
	if err := c.JSONSerializer.Deserialize(responseData, &encoderResp); err != nil {
		return 0, fmt.Errorf("failed to deserialize protocol encoder response: %w", err)
	}

	return uint32(encoderResp.Result), nil //nolint:gosec // This is how the protocol works
}
