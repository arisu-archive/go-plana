package plana

import (
	"context"
	"errors"
	"fmt"

	"github.com/arisu-archive/plana-flatbuffers/go/flatdata"
	"github.com/arisu-archive/plana-protos/protos"
)

type RaidService service

type RaidOpponentListRequestWrapper struct {
	*protos.RaidOpponentListRequest
}

func (w RaidOpponentListRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

var ErrRaidSearchType = errors.New("invalid raid opponent search type")

type RaidOpponentListRequestBuilder struct {
	service *RaidService
	payload RaidOpponentListRequestWrapper
}

type RaidOpponentListOption func(*protos.RaidOpponentListRequest)

func (b RaidOpponentListRequestBuilder) Search(
	ctx context.Context,
	session *UserSession,
) (*protos.RaidOpponentListResponse, error) {
	return b.service.GetOpponents(ctx, session, b.payload)
}

func (s *RaidService) WithOpponentRank(rank int64) RaidOpponentListRequestBuilder {
	return RaidOpponentListRequestBuilder{
		service: s,
		payload: RaidOpponentListRequestWrapper{
			RaidOpponentListRequest: &protos.RaidOpponentListRequest{
				SearchType: flatdata.RankingSearchTypeRank,
				Rank:       &rank,
			},
		},
	}
}

func (s *RaidService) WithOpponentScore(score int64) RaidOpponentListRequestBuilder {
	return RaidOpponentListRequestBuilder{
		service: s,
		payload: RaidOpponentListRequestWrapper{
			RaidOpponentListRequest: &protos.RaidOpponentListRequest{
				SearchType: flatdata.RankingSearchTypeScore,
				Score:      &score,
			},
		},
	}
}

func (s *RaidService) GetOpponents(
	ctx context.Context,
	session *UserSession,
	payload RaidOpponentListRequestWrapper,
) (*protos.RaidOpponentListResponse, error) {
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_Raid_OpponentList, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create raid opponent list request: %w", err)
	}
	result := new(protos.RaidOpponentListResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("raid opponent list request failed: %w", err)
	}
	return result, nil
}

type RaidLobbyRequestWrapper struct {
	*protos.RaidLobbyRequest
}

func (w RaidLobbyRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *RaidService) Lobby(ctx context.Context, session *UserSession) (*protos.RaidLobbyResponse, error) {
	w := RaidLobbyRequestWrapper{
		RaidLobbyRequest: &protos.RaidLobbyRequest{},
	}
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_Raid_Lobby, w)
	if err != nil {
		return nil, fmt.Errorf("failed to create raid lobby request: %w", err)
	}
	result := new(protos.RaidLobbyResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("raid lobby request failed: %w", err)
	}
	return result, nil
}

type RaidGetBestTeamRequestWrapper struct {
	*protos.RaidGetBestTeamRequest
}

func (w RaidGetBestTeamRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *RaidService) GetBestTeam(ctx context.Context, session *UserSession, accountID int64) (*protos.RaidGetBestTeamResponse, error) {
	w := RaidGetBestTeamRequestWrapper{
		RaidGetBestTeamRequest: &protos.RaidGetBestTeamRequest{
			SearchAccountId: accountID,
		},
	}
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_Raid_GetBestTeam, w)
	if err != nil {
		return nil, fmt.Errorf("failed to create raid get best team request: %w", err)
	}
	result := new(protos.RaidGetBestTeamResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("raid get best team request failed: %w", err)
	}
	return result, nil
}

type RaidRankingIndexRequestWrapper struct {
	*protos.RaidRankingIndexRequest
}

func (w RaidRankingIndexRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *RaidService) GetRankingIndex(ctx context.Context, session *UserSession) (*protos.RaidRankingIndexResponse, error) {
	w := RaidRankingIndexRequestWrapper{
		RaidRankingIndexRequest: &protos.RaidRankingIndexRequest{},
	}
	req, err := s.client.R().WithSession(session).Game(ctx, protos.Protocol_Raid_RankingIndex, w)
	if err != nil {
		return nil, fmt.Errorf("failed to create raid ranking index request: %w", err)
	}
	result := new(protos.RaidRankingIndexResponse)
	_, err = s.client.Do(ctx, req, result)
	if err != nil {
		return nil, fmt.Errorf("raid ranking index request failed: %w", err)
	}
	return result, nil
}
