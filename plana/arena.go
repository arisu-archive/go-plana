package plana

import (
	"context"
	"fmt"

	"github.com/arisu-archive/plana-protos/protos"
)

type ArenaService service

type ArenaRankListRequestWrapper struct {
	*protos.ArenaRankListRequest
}

func (w ArenaRankListRequestWrapper) Packet() *protos.RequestPacket {
	return &w.RequestPacket
}

func (s *ArenaService) GetRanks(
	ctx context.Context,
	rank int32,
	count int32,
) (*protos.ArenaRankListResponse, error) {
	param := ArenaRankListRequestWrapper{
		ArenaRankListRequest: &protos.ArenaRankListRequest{
			StartIndex: rank,
			Count:      count,
		},
	}
	req, err := s.client.R().Game(ctx, protos.Protocol_Arena_RankList, param)
	if err != nil {
		return nil, fmt.Errorf("failed to create arena rank list request: %w", err)
	}
	result := new(protos.ArenaRankListResponse)
	_, err = s.client.Do(req, result)
	if err != nil {
		return nil, fmt.Errorf("failed to get arena rank list response: %w", err)
	}
	return result, nil
}
