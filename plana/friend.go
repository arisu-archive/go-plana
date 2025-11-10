package plana

import (
	"context"
	"fmt"

	"github.com/arisu-archive/plana-flatbuffers/go/flatdata"
	"github.com/arisu-archive/plana-protos/protos"
)

type FriendService service

type FriendGetFriendDetailedInfoRequestWrapper struct {
	*protos.FriendGetFriendDetailedInfoRequest
}

func (w FriendGetFriendDetailedInfoRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *FriendService) GetDetail(ctx context.Context, session *UserSession, accountID int64) (*protos.FriendGetFriendDetailedInfoResponse, error) {
	param := FriendGetFriendDetailedInfoRequestWrapper{
		&protos.FriendGetFriendDetailedInfoRequest{
			FriendAccountId: accountID,
		},
	}
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_Friend_GetFriendDetailedInfo, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create get friend detail request: %w", err)
	}
	result := new(protos.FriendGetFriendDetailedInfoResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("get friend detail request failed: %w", err)
	}
	return result, nil
}

type FriendSearchRequestWrapper struct {
	*protos.FriendSearchRequest
}

func (w FriendSearchRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

type FriendSearchRequestBuilder struct {
	service *FriendService
	payload FriendSearchRequestWrapper
}

func (s *FriendService) Search() FriendSearchRequestBuilder {
	return FriendSearchRequestBuilder{
		service: s,
		payload: FriendSearchRequestWrapper{
			FriendSearchRequest: &protos.FriendSearchRequest{},
		},
	}
}

func (b FriendSearchRequestBuilder) ByCode(friendCode string) FriendSearchRequestBuilder {
	b.payload.FriendCode = friendCode
	return b
}

func (b FriendSearchRequestBuilder) WithLevelOption(option flatdata.FriendSearchLevelOption) FriendSearchRequestBuilder {
	b.payload.LevelOption = option
	return b
}

func (b FriendSearchRequestBuilder) Execute(ctx context.Context, session *UserSession) (*protos.FriendSearchResponse, error) {
	return b.service.submitSearch(ctx, session, b.payload)
}

func (s *FriendService) submitSearch(ctx context.Context, session *UserSession, w FriendSearchRequestWrapper) (*protos.FriendSearchResponse, error) {
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_Friend_Search, w)
	if err != nil {
		return nil, fmt.Errorf("failed to create friend search request: %w", err)
	}
	result := new(protos.FriendSearchResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("friend search request failed: %w", err)
	}
	return result, nil
}
