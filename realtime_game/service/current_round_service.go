package service

import (
	"context"
	"crash/realtime_game/domain"
	rttypes "crash/realtime_game/types"
	"time"
)

// CurrentRoundService 负责查询当前局。
type CurrentRoundService struct{ *Services }

func NewCurrentRoundService(s *Services) *CurrentRoundService {
	return &CurrentRoundService{Services: s}
}

func (s *CurrentRoundService) Get(ctx context.Context, channelID int64) (*rttypes.CurrentRoundResponse, error) {
	pair, err := resolveRuntimeChannel(ctx, s.Services, channelID)
	if err != nil {
		return nil, err
	}

	snap, err := s.Ctx.SnapshotStore.GetSnapshot(ctx, pair.Runtime.Id)
	if err != nil {
		return nil, err
	}
	if snap == nil {
		return nil, nil
	}
	nowMs := time.Now().UnixMilli()
	snap.ServerTimeMs = nowMs
	return &rttypes.CurrentRoundResponse{
		RoundSnapshot:   snap,
		CurrentMultiple: s.currentMultiple(snap, nowMs),
	}, nil
}

func (s *CurrentRoundService) currentMultiple(snap *domain.RoundSnapshot, nowMs int64) int64 {
	switch snap.State {
	case domain.RoundStatePreStart, domain.RoundStateStarting:
		return domain.MultipleScale
	case domain.RoundStateFlying:
		v := domain.CalcCurrentMultiple(snap.IncNum, snap.FlyingStartAtMs, nowMs)
		if v > snap.CrashMultiple {
			return snap.CrashMultiple
		}
		return v
	case domain.RoundStateCrashed, domain.RoundStateClosed:
		return snap.CrashMultiple
	default:
		return domain.MultipleScale
	}
}
