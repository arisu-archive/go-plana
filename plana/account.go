package plana

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/arisu-archive/plana-protos/protos"
	"github.com/google/uuid"
)

const (
	defaultGooglePlayStoreMarketId = "gps"
	defaultAccessIP                = "192.168.0.16"
	defaultOSType                  = "A"
	defaultFullOSType              = "AOS"
	defaultOSVersion               = "Android OS 12 / API-32 (PQ3A.190605.09261202/3793265)"
	defaultDeviceModel             = "samsung SM-N976N"
	defaultDeviceSystemMemorySize  = 8192
)

type AccountService service

// AccountAuthOption is a functional option for Authenticate allowing callers to
// override default fields of protos.AccountAuthRequest. Example usage:
//
//	s.Authenticate(ctx, key, WithAccessIP("1.2.3.4"), WithDeviceModel("Pixel 7"))
type AccountAuthOption func(*protos.AccountAuthRequest)

type AccountAuthRequestWrapper struct {
	*protos.AccountAuthRequest
}

func (w AccountAuthRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *AccountService) Authenticate(
	ctx context.Context,
	credential *UserSession,
	opts ...AccountAuthOption,
) (*protos.AccountAuthResponse, error) {
	// Generate random bytes
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}
	param := AccountAuthRequestWrapper{
		&protos.AccountAuthRequest{
			AccessIP:               defaultAccessIP,
			MarketId:               defaultGooglePlayStoreMarketId,
			AdvertisementId:        uuid.NewString(),
			OSType:                 defaultOSType,
			OSVersion:              defaultOSVersion,
			DeviceUniqueId:         hex.EncodeToString(randomBytes),
			DeviceModel:            defaultDeviceModel,
			DeviceSystemMemorySize: defaultDeviceSystemMemorySize,
		},
	}
	// Apply functional options to allow callers to override defaults.
	for _, o := range opts {
		o(param.AccountAuthRequest)
	}
	req, err := s.client.R().WithSession(credential).Game(ctx, protos.Protocol_Account_Auth, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticate request: %w", err)
	}
	result := new(protos.AccountAuthResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("authenticate request failed: %w", err)
	}
	return result, nil
}

type AESKeyBundle struct {
	Key []byte
	IV  []byte
}

type YostarCheckOption struct {
	Cookie      string
	EnterTicket string
	KeyBundle   AESKeyBundle
}

type AccountCheckYostarRequestWrapper struct {
	*protos.AccountCheckYostarRequest
}

func (w AccountCheckYostarRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *AccountService) CheckYostar(
	ctx context.Context,
	ops YostarCheckOption,
) (*protos.AccountCheckYostarResponse, error) {
	param := AccountCheckYostarRequestWrapper{
		&protos.AccountCheckYostarRequest{
			EnterTicket: ops.EnterTicket,
			Cookie:      ops.Cookie,
		},
	}
	req, err := s.client.R().Game(ctx, protos.Protocol_Account_CheckYostar, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create check yostar request: %w", err)
	}
	result := new(protos.AccountCheckYostarResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("check nexon request failed: %w", err)
	}
	return result, nil
}
