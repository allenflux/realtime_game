package service

import (
	"context"
	"crash/model/servermodel"
	"crash/realtime_game/domain"
	"crash/realtime_game/settlement"
	"time"
)

// AutoCashoutService 负责自动兑现。
type AutoCashoutService struct{ *Services }

func NewAutoCashoutService(s *Services) *AutoCashoutService { return &AutoCashoutService{Services: s} }

func (s *AutoCashoutService) RunOnce(ctx context.Context, snap *domain.RoundSnapshot) error {
	if snap == nil || snap.State != domain.RoundStateFlying {
		return nil
	}
	current := domain.CalcCurrentMultiple(snap.IncNum, snap.FlyingStartAtMs, time.Now().UnixMilli())
	if current > snap.CrashMultiple {
		current = snap.CrashMultiple
	}
	currentComparable := current * domain.MultipleTail

	orderNos, err := s.Ctx.SnapshotStore.ListDueAutoCashouts(ctx, snap.ChannelID, currentComparable, s.Ctx.Config.Runtime.AutoCashoutBatchSize)
	if err != nil || len(orderNos) == 0 {
		return err
	}

	channel, err := s.Ctx.ChannelModel.FindOne(ctx, snap.ChannelID)
	if err != nil {
		return err
	}

	for _, orderNo := range orderNos {
		if err := s.handleAutoOne(ctx, channel, snap, orderNo, currentComparable); err != nil {
			continue
		}
	}
	return nil
}

func (s *AutoCashoutService) handleAutoOne(ctx context.Context, channel *servermodel.Channel, snap *domain.RoundSnapshot, orderNo string, currentComparable int64) error {
	lockOK, err := s.Ctx.SnapshotStore.AcquireOpLock(ctx, "cashout:"+orderNo, 5)
	if err != nil || !lockOK {
		return err
	}

	bet, err := s.Ctx.BetModel.GetByApiOrderNo(ctx, orderNo)
	if err != nil || bet == nil {
		return err
	}
	if bet.OrderStatus != servermodel.OrderStatusCreated {
		_ = s.Ctx.SnapshotStore.RemoveAutoCashout(ctx, channel.Id, orderNo)
		return nil
	}
	if bet.TermId != snap.TermID {
		_ = s.Ctx.SnapshotStore.RemoveAutoCashout(ctx, channel.Id, orderNo)
		return nil
	}

	payout := calcCashoutAmount(bet, currentComparable, channel.MaxCashoutPerBet, false)
	if payout <= 0 {
		return nil
	}
	if err := s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{
		servermodel.Bet_F_order_status:      servermodel.OrderStatusCashingOut,
		servermodel.Bet_F_cashed_out_amount: payout,
	}); err != nil {
		return err
	}
	bet.CashedOutAmount = payout

	var billErr error
	if bet.GamePlay == domain.GamePlayPreMatch {
		billErr = s.Ctx.Settlement.BillPreMatch(ctx, settlement.BillRequest{
			ChannelID:      bet.ChannelId,
			UserID:         bet.UserId,
			OrderNo:        bet.ApiOrderNo,
			Currency:       bet.Currency,
			Amount:         domain.DBAmountToString(payout),
			Metadata:       domain.BuildBetMetadata(bet),
			IsSystemReward: false,
		}, false, 0)
	} else {
		billErr = s.Ctx.Settlement.BillRolling(ctx, settlement.BillRequest{
			ChannelID:      bet.ChannelId,
			UserID:         bet.UserId,
			OrderNo:        bet.ApiOrderNo,
			Currency:       bet.Currency,
			Amount:         domain.DBAmountToString(payout),
			Metadata:       domain.BuildBetMetadata(bet),
			IsSystemReward: false,
		})
	}
	if billErr != nil {
		//_ = s.addCashoutRetry(ctx, bet)
		_ = s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{servermodel.Bet_F_order_status: servermodel.OrderStatusOverRetry})
		return billErr
	}
	cashoutService := NewCashoutService(s.Services)
	return cashoutService.markBetSettled(ctx, bet, payout)
}
