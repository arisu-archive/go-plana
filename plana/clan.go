package plana

import (
	"context"
	"fmt"

	"github.com/arisu-archive/plana-flatbuffers/go/flatdata"
	"github.com/arisu-archive/plana-protos/protos"
)

type ClanService service

type ClanSearchRequestWrapper struct {
	*protos.ClanSearchRequest
}

func (w ClanSearchRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

type ClanSearchRequestBuilder struct {
	service *ClanService
	payload ClanSearchRequestWrapper
}

func (s *ClanService) Search() ClanSearchRequestBuilder {
	return ClanSearchRequestBuilder{
		service: s,
		payload: ClanSearchRequestWrapper{
			ClanSearchRequest: &protos.ClanSearchRequest{},
		},
	}
}

func (b ClanSearchRequestBuilder) ByName(name string) ClanSearchRequestBuilder {
	// Mutually exclusive with ByCode
	b.payload.ClanUniqueCode = ""
	b.payload.SearchString = name
	return b
}

func (b ClanSearchRequestBuilder) ByCode(code string) ClanSearchRequestBuilder {
	// Mutually exclusive with ByName
	b.payload.SearchString = ""
	b.payload.ClanUniqueCode = code
	return b
}

func (b ClanSearchRequestBuilder) WithJoinOption(option flatdata.ClanJoinOption) ClanSearchRequestBuilder {
	b.payload.ClanJoinOption = option
	return b
}

func (b ClanSearchRequestBuilder) Execute(
	ctx context.Context,
	session *UserSession,
) (*protos.ClanSearchResponse, error) {
	return b.service.submitSearch(ctx, session, b.payload)
}

func (s *ClanService) submitSearch(
	ctx context.Context,
	session *UserSession,
	param ClanSearchRequestWrapper,
) (*protos.ClanSearchResponse, error) {
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_Clan_Search, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create clan search request: %w", err)
	}
	result := new(protos.ClanSearchResponse)
	_, err = s.client.Do(req, result)
	if err != nil {
		return nil, fmt.Errorf("clan search request failed: %w", err)
	}
	return result, nil
}

type ClanMemberListRequestWrapper struct {
	*protos.ClanMemberListRequest
}

func (w ClanMemberListRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *ClanService) GetMembers(
	ctx context.Context,
	session *UserSession,
	clanID int64,
) (*protos.ClanMemberListResponse, error) {
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_Clan_MemberList, ClanMemberListRequestWrapper{
		ClanMemberListRequest: &protos.ClanMemberListRequest{
			ClanDBId: clanID,
		},
	})
	if err != nil {
		return nil, err
	}
	result := new(protos.ClanMemberListResponse)
	_, err = s.client.Do(req, result)
	if err != nil {
		return nil, fmt.Errorf("clan member list request failed: %w", err)
	}
	return result, nil
}
