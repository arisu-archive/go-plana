package plana

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/arisu-archive/plana-protos/protos"
)

type encoderRequest struct {
	Server   string `json:"server"`
	Protocol uint64 `json:"protocol"`
	Crc32    uint64 `json:"crc32"`
}

func (c *Client) encodeProtocol(ctx context.Context, crc32 uint32, p protos.Protocol) (uint32, error) {
	data := encoderRequest{
		Server:   "japan",
		Protocol: uint64(p), //nolint:gosec // This is how the protocol works
		Crc32:    uint64(crc32),
	}
	payload, err := c.JSONSerializer.Serialize(data, "")
	if err != nil {
		return 0, fmt.Errorf("failed to serialize protocol encoder request: %w", err)
	}

	u, err := c.ProtocolEncoderURL.Parse("/")
	if err != nil {
		return 0, fmt.Errorf("failed to parse protocol encoder URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewBuffer(payload))
	if err != nil {
		return 0, fmt.Errorf("failed to create protocol encoder request: %w", err)
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
