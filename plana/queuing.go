package plana

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/arisu-archive/plana-protos/protos"
)

type QueuingService service

type GetAuthTicketOptions struct {
	KeyBundle     AESKeyBundle
	ClientVersion string
	YostarUID     int64
	YostarToken   string
}

type QueuingGetAuthTicketRequestWrapper struct {
	*protos.QueuingGetAuthTicketRequest
}

func (p QueuingGetAuthTicketRequestWrapper) Packet() *protos.RequestPacket {
	return &p.RequestPacket
}

func (s *QueuingService) GetAuthTicket(ctx context.Context, data GetAuthTicketOptions) (*protos.QueuingGetAuthTicketResponse, error) {
	param := QueuingGetAuthTicketRequestWrapper{
		QueuingGetAuthTicketRequest: &protos.QueuingGetAuthTicketRequest{
			YostarUID:          data.YostarUID,
			YostarToken:        data.YostarToken,
			ClientVersion:      data.ClientVersion,
			OSType:             defaultFullOSType,
			ClientGeneratedKey: base64.StdEncoding.EncodeToString(data.KeyBundle.Key),
			ClientGeneratedIV:  base64.StdEncoding.EncodeToString(data.KeyBundle.IV),
		},
	}
	// It requires a session for this request
	req, err := s.client.R().Gateway(ctx, protos.Protocol_Queuing_GetAuthTicket, param, WithHash(0))
	if err != nil {
		return nil, fmt.Errorf("failed to create get auth ticket request: %w", err)
	}
	result := new(protos.QueuingGetAuthTicketResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("get auth ticket request failed: %w", err)
	}
	return result, nil
}

type QueuingProcessWaitingQueueRequestWrapper struct {
	*protos.QueuingProcessWaitingQueueRequest
}

func (p QueuingProcessWaitingQueueRequestWrapper) Packet() *protos.RequestPacket {
	return &p.RequestPacket
}

type ProcessWaitingQueueOptions struct {
	WaitingTicket string
	AuthTicket    string
	ClientVersion string
}

func (s *QueuingService) ProcessWaitingQueue(ctx context.Context, session *UserSession, data ProcessWaitingQueueOptions) (*protos.QueuingProcessWaitingQueueResponse, error) {
	param := QueuingProcessWaitingQueueRequestWrapper{
		QueuingProcessWaitingQueueRequest: &protos.QueuingProcessWaitingQueueRequest{
			WaitingTicket: data.WaitingTicket,
			AuthTicket:    data.AuthTicket,
			OSType:        defaultFullOSType,
			ClientVersion: data.ClientVersion,
		},
	}
	req, err := s.client.R().WithSession(session).Gateway(ctx, protos.Protocol_Queuing_ProcessWaitingQueue, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create queuing process request: %w", err)
	}
	result := new(protos.QueuingProcessWaitingQueueResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("queuing process request failed: %w", err)
	}
	return result, nil
}
