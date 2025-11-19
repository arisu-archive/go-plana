package plana

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/arisu-archive/plana-protos/protos"
)

type QueuingService service

type GetCryptoKeysOptions struct {
	KeyBundle AESKeyBundle
}

type QueuingGetCryptoKeysRequestWrapper struct {
	*protos.QueuingGetCryptoKeysRequest
}

func (p QueuingGetCryptoKeysRequestWrapper) Packet() *protos.RequestPacket {
	return &p.RequestPacket
}

func (s *QueuingService) GetCryptoKeys(ctx context.Context, data GetCryptoKeysOptions) (*protos.QueuingGetCryptoKeysResponse, error) {
	param := QueuingGetCryptoKeysRequestWrapper{
		QueuingGetCryptoKeysRequest: &protos.QueuingGetCryptoKeysRequest{
			ClientGeneratedKey: base64.StdEncoding.EncodeToString(rsaEncrypt(data.KeyBundle.Key, s.client.publicKey)),
			ClientGeneratedIV:  base64.StdEncoding.EncodeToString(rsaEncrypt(data.KeyBundle.IV, s.client.publicKey)),
		},
	}
	req, err := s.client.R().Gateway(
		ctx,
		protos.Protocol_Queuing_GetCryptoKeys,
		param,
		WithHash(0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create get crypto keys request: %w", err)
	}
	result := new(protos.QueuingGetCryptoKeysResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("get crypto keys request failed: %w", err)
	}
	return result, nil
}

type GetTicketOptions struct {
	ClientVersion string
	YostarUID     int64
	YostarToken   string
}

type QueuingGetTicketRequestWrapper struct {
	*protos.QueuingGetTicketRequest
}

func (p QueuingGetTicketRequestWrapper) Packet() *protos.RequestPacket {
	return &p.RequestPacket
}

func (s *QueuingService) GetTicket(ctx context.Context, data GetTicketOptions) (*protos.QueuingGetTicketResponse, error) {
	param := QueuingGetTicketRequestWrapper{
		QueuingGetTicketRequest: &protos.QueuingGetTicketRequest{
			YostarUID:     data.YostarUID,
			YostarToken:   data.YostarToken,
			ClientVersion: data.ClientVersion,
			OSType:        defaultFullOSType,
		},
	}
	req, err := s.client.R().Gateway(
		ctx,
		protos.Protocol_Queuing_GetTicket,
		param,
		WithHash(0),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create get ticket request: %w", err)
	}
	result := new(protos.QueuingGetTicketResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("get ticket request failed: %w", err)
	}
	return result, nil
}
