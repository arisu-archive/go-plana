package plana

import (
	"context"
	"fmt"

	"github.com/arisu-archive/plana-flatbuffers/go/flatdata"
	"github.com/arisu-archive/plana-protos/protos"
)

type EliminateRaidService service

type EliminateRaidOpponentListRequestWrapper struct {
	*protos.EliminateRaidOpponentListRequest
}

func (w EliminateRaidOpponentListRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

type EliminateRaidOpponentListOption func(*protos.EliminateRaidOpponentListRequest)

type EliminateRaidOpponentListRequestBuilder struct {
	service *EliminateRaidService
	payload EliminateRaidOpponentListRequestWrapper
}

func (b EliminateRaidOpponentListRequestBuilder) WithBossGroup(bossGroup int32) EliminateRaidOpponentListRequestBuilder {
	b.payload.BossGroupIndex = &bossGroup
	return b
}

func (b EliminateRaidOpponentListRequestBuilder) Search(
	ctx context.Context,
	session *UserSession,
	opts ...EliminateRaidOpponentListOption,
) (*protos.EliminateRaidOpponentListResponse, error) {
	return b.service.GetOpponents(ctx, session, b.payload, opts...)
}

func (s *EliminateRaidService) WithOpponentRank(rank int64) EliminateRaidOpponentListRequestBuilder {
	return EliminateRaidOpponentListRequestBuilder{
		service: s,
		payload: EliminateRaidOpponentListRequestWrapper{
			EliminateRaidOpponentListRequest: &protos.EliminateRaidOpponentListRequest{
				SearchType: flatdata.RankingSearchTypeRank,
				Rank:       &rank,
			},
		},
	}
}

func (s *EliminateRaidService) WithOpponentScore(score int64) EliminateRaidOpponentListRequestBuilder {
	return EliminateRaidOpponentListRequestBuilder{
		service: s,
		payload: EliminateRaidOpponentListRequestWrapper{
			EliminateRaidOpponentListRequest: &protos.EliminateRaidOpponentListRequest{
				SearchType: flatdata.RankingSearchTypeScore,
				Score:      &score,
			},
		},
	}
}

func (s *EliminateRaidService) GetOpponents(
	ctx context.Context,
	session *UserSession,
	payload EliminateRaidOpponentListRequestWrapper,
	opts ...EliminateRaidOpponentListOption,
) (*protos.EliminateRaidOpponentListResponse, error) {
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_EliminateRaid_OpponentList, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create eliminate raid opponent list request: %w", err)
	}
	for _, o := range opts {
		o(payload.EliminateRaidOpponentListRequest)
	}
	result := new(protos.EliminateRaidOpponentListResponse)
	_, err = s.client.Do(req, result)
	if err != nil {
		return nil, fmt.Errorf("eliminate raid opponent list request failed: %w", err)
	}
	return result, nil
}

type EliminateRaidLobbyRequestWrapper struct {
	*protos.EliminateRaidLobbyRequest
}

func (w EliminateRaidLobbyRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *EliminateRaidService) GetLobby(ctx context.Context, session *UserSession) (*protos.EliminateRaidLobbyResponse, error) {
	param := EliminateRaidLobbyRequestWrapper{
		EliminateRaidLobbyRequest: &protos.EliminateRaidLobbyRequest{},
	}
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_EliminateRaid_Lobby, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create eliminate raid lobby request: %w", err)
	}
	result := new(protos.EliminateRaidLobbyResponse)
	_, err = s.client.Do(req, result)
	if err != nil {
		return nil, fmt.Errorf("eliminate raid lobby request failed: %w", err)
	}
	return result, nil
}

type EliminateRaidGetBestTeamRequestWrapper struct {
	*protos.EliminateRaidGetBestTeamRequest
}

func (w EliminateRaidGetBestTeamRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *EliminateRaidService) GetBestTeam(ctx context.Context, session *UserSession, accountID int64) (*protos.EliminateRaidGetBestTeamResponse, error) {
	w := EliminateRaidGetBestTeamRequestWrapper{
		EliminateRaidGetBestTeamRequest: &protos.EliminateRaidGetBestTeamRequest{
			SearchAccountId: accountID,
		},
	}
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_EliminateRaid_GetBestTeam, w)
	if err != nil {
		return nil, fmt.Errorf("failed to create eliminate raid get best team request: %w", err)
	}
	result := new(protos.EliminateRaidGetBestTeamResponse)
	_, err = s.client.Do(req, result)
	if err != nil {
		return nil, fmt.Errorf("eliminate raid get best team request failed: %w", err)
	}
	return result, nil
}

type EliminateRaidRankingIndexRequestWrapper struct {
	*protos.EliminateRaidRankingIndexRequest
}

func (w EliminateRaidRankingIndexRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *EliminateRaidService) GetRankingIndex(ctx context.Context, session *UserSession) (*protos.EliminateRaidRankingIndexResponse, error) {
	param := EliminateRaidRankingIndexRequestWrapper{
		EliminateRaidRankingIndexRequest: &protos.EliminateRaidRankingIndexRequest{},
	}
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_EliminateRaid_RankingIndex, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create eliminate raid ranking index request: %w", err)
	}
	result := new(protos.EliminateRaidRankingIndexResponse)
	_, err = s.client.Do(req, result)
	if err != nil {
		return nil, fmt.Errorf("eliminate raid ranking index request failed: %w", err)
	}
	return result, nil
}
