package service

import (
	"context"
	"crash/model/servermodel"
	"crash/realtime_game/domain"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

// WorkerService 是 worker 侧主循环。
type WorkerService struct{ *Services }

func NewWorkerService(s *Services) *WorkerService { return &WorkerService{Services: s} }

func (s *WorkerService) TickOne(ctx context.Context, channel *servermodel.Channel, workerID string) error {
	// 抢 owner。
	owner, _ := s.Ctx.LeaseStore.Owner(ctx, channel.Id)
	if owner == "" {
		ok, err := s.Ctx.LeaseStore.Acquire(ctx, channel.Id, workerID, int(s.Ctx.Config.Runtime.LeaseTTLSeconds))
		if err != nil || !ok {
			return err
		}
	} else if owner == workerID {
		_, err := s.Ctx.LeaseStore.Renew(ctx, channel.Id, workerID, int(s.Ctx.Config.Runtime.LeaseTTLSeconds))
		if err != nil {
			return err
		}
	} else {
		return nil
	}

	openSvc := NewOpenRoundService(s.Services)
	current, err := openSvc.OpenIfNeeded(ctx, channel)
	if err != nil || current == nil {
		return err
	}

	nowMs := time.Now().UnixMilli()
	changed := false

	if current.State == domain.RoundStatePreStart && nowMs >= current.BetCloseAtMs {
		current.State = domain.RoundStateStarting
		current.Version = domain.NextVersion(current.Version)
		changed = true
	}
	if current.State == domain.RoundStateStarting && nowMs >= current.FlyingStartAtMs {
		current.State = domain.RoundStateFlying
		current.Version = domain.NextVersion(current.Version)
		changed = true
	}
	if current.State == domain.RoundStateFlying {
		autoSvc := NewAutoCashoutService(s.Services)
		_ = autoSvc.RunOnce(ctx, current)
		if nowMs >= current.CrashAtMs {
			current.State = domain.RoundStateCrashed
			current.IsCrashed = servermodel.CrashTermIsCrashedYes
			current.CrashedAtMs = nowMs
			current.Version = domain.NextVersion(current.Version)
			changed = true
		}
	}
	if current.State == domain.RoundStateCrashed && nowMs >= current.CloseAtMs {
		//closeSvc := NewCloseRoundService(s.Services)
		//return closeSvc.Close(ctx, current)

		// 1. 先准时切状态
		current.State = domain.RoundStateClosed
		current.Version = domain.NextVersion(current.Version)
		current.ClosedAtMs = nowMs

		_ = s.Ctx.SnapshotStore.SaveSnapshot(ctx, current)

		// 2. 异步收口
		go func(round *domain.RoundSnapshot) {
			closeSvc := NewCloseRoundService(s.Services)
			if err := closeSvc.Finalize(context.Background(), round); err != nil {
				logx.Errorf("finalize failed: %v", err)
			}
		}(current)
	}
	if changed {
		return s.Ctx.SnapshotStore.SaveSnapshot(ctx, current)
	}
	return nil
}
