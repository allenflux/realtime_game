package service

import (
	"context"
	"crash/model/servermodel"
	"crash/realtime_game/domain"
	"crash/realtime_game/settlement"
	rttypes "crash/realtime_game/types"
	"errors"
	"time"
)

// CashoutService 负责手动兑现。
type CashoutService struct{ *Services }

func NewCashoutService(s *Services) *CashoutService { return &CashoutService{Services: s} }

func (s *CashoutService) Cashout(ctx context.Context, req *rttypes.CashoutRequest) (*rttypes.CashoutResponse, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}
	if req.OrderNo == "" || req.UserID <= 0 {
		return nil, errors.New("order_no 或 user_id 非法")
	}

	lockOK, err := s.Ctx.SnapshotStore.AcquireOpLock(ctx, "cashout:"+req.OrderNo, 5)
	if err != nil {
		return nil, err
	}
	if !lockOK {
		return nil, errors.New("该订单正在处理")
	}

	bet, err := s.Ctx.BetModel.GetByApiOrderNo(ctx, req.OrderNo)
	if err != nil || bet == nil {
		return nil, errors.New("订单不存在")
	}
	if bet.UserId != req.UserID {
		return nil, errors.New("不能操作他人订单")
	}
	if bet.OrderStatus != servermodel.OrderStatusCreated && bet.OrderStatus != servermodel.OrderStatusCashingOut {
		return nil, errors.New("当前订单状态不可兑现")
	}

	snap, err := s.Ctx.SnapshotStore.GetSnapshot(ctx, bet.ChannelId)
	if err != nil {
		return nil, err
	}
	if snap == nil || snap.TermID != bet.TermId {
		return nil, errors.New("当前局不存在")
	}
	if snap.State == domain.RoundStateCrashed || snap.State == domain.RoundStateClosed {
		return nil, errors.New("本局已结束，不能手动兑现")
	}

	channel, err := s.Ctx.ChannelModel.FindOne(ctx, bet.ChannelId)
	if err != nil {
		return nil, err
	}

	currentMultiple := s.currentComparable(snap)
	bet.ManualCashoutMultiple = currentMultiple

	switch req.GamePlay {
	case domain.GamePlayRollingPlate:
		return s.cashoutRolling(ctx, bet, channel)
	case domain.GamePlayPreMatch:
		if req.SettlementMode == domain.CashoutHalf {
			return s.cashoutPreHalf(ctx, bet, channel)
		}
		return s.cashoutPreAll(ctx, bet, channel)
	default:
		return nil, errors.New("不支持的玩法")
	}
}

func (s *CashoutService) currentComparable(snap *domain.RoundSnapshot) int64 {
	current := domain.MultipleScale
	if snap.State == domain.RoundStateFlying {
		current = domain.CalcCurrentMultiple(snap.IncNum, snap.FlyingStartAtMs, time.Now().UnixMilli())
		if current > snap.CrashMultiple {
			current = snap.CrashMultiple
		}
	} else if snap.State == domain.RoundStateCrashed || snap.State == domain.RoundStateClosed {
		current = snap.CrashMultiple
	}
	return current * domain.MultipleTail
}

func (s *CashoutService) cashoutRolling(ctx context.Context, bet *servermodel.Bet, channel *servermodel.Channel) (*rttypes.CashoutResponse, error) {
	payout := calcCashoutAmount(bet, bet.ManualCashoutMultiple, channel.MaxCashoutPerBet, true)
	bet.CashedOutAmount = payout
	if err := s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{
		servermodel.Bet_F_order_status:            servermodel.OrderStatusCashingOut,
		servermodel.Bet_F_cashed_out_amount:       payout,
		servermodel.Bet_F_manual_cashout_multiple: bet.ManualCashoutMultiple,
	}); err != nil {
		return nil, err
	}

	err := s.Ctx.Settlement.BillRolling(ctx, settlement.BillRequest{
		ChannelID:      bet.ChannelId,
		UserID:         bet.UserId,
		OrderNo:        bet.ApiOrderNo,
		Currency:       bet.Currency,
		Amount:         domain.DBAmountToString(payout),
		Metadata:       domain.BuildBetMetadata(bet),
		IsSystemReward: false,
	})
	if err != nil {
		_ = s.addCashoutRetry(ctx, bet)
		_ = s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{servermodel.Bet_F_order_status: servermodel.OrderStatusOverRetry})
		return nil, err
	}

	if err := s.markBetSettled(ctx, bet, payout); err != nil {
		return nil, err
	}
	return buildCashoutResp(bet, 0), nil
}

func (s *CashoutService) cashoutPreHalf(ctx context.Context, bet *servermodel.Bet, channel *servermodel.Channel) (*rttypes.CashoutResponse, error) {
	if bet.FirstCashoutAmount > 0 {
		return nil, errors.New("该订单已经半兑过")
	}
	total := calcCashoutAmount(bet, bet.ManualCashoutMultiple, channel.MaxCashoutPerBet, true)
	half := total / 2
	bet.CashedOutAmount = half
	bet.FirstCashoutAmount = half
	bet.FirstManualCashoutMultiple = bet.ManualCashoutMultiple
	bet.ManualCashoutTimes = 1

	if err := s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{
		servermodel.Bet_F_order_status:                  servermodel.OrderStatusCashedout,
		servermodel.Bet_F_cashed_out_amount:             bet.CashedOutAmount,
		servermodel.Bet_F_manual_cashout_multiple:       bet.ManualCashoutMultiple,
		servermodel.Bet_F_First_Cashout_Amount:          bet.FirstCashoutAmount,
		servermodel.Bet_F_First_Manual_Cashout_Multiple: bet.FirstManualCashoutMultiple,
		servermodel.Bet_F_Cashout_Times:                 bet.ManualCashoutTimes,
	}); err != nil {
		return nil, err
	}

	err := s.Ctx.Settlement.BillPreMatch(ctx, settlement.BillRequest{
		ChannelID:      bet.ChannelId,
		UserID:         bet.UserId,
		OrderNo:        bet.ApiOrderNo,
		Currency:       bet.Currency,
		Amount:         domain.DBAmountToString(half),
		Metadata:       domain.BuildBetMetadataWithHalf(bet, true),
		IsSystemReward: false,
	}, true, 0)
	if err != nil {
		_ = s.addCashoutRetry(ctx, bet)
		return nil, err
	}

	// 第一段结算成功后，这笔单仍然从自动兑现队列中移除，避免后续再次自动兑。
	_ = s.Ctx.SnapshotStore.RemoveAutoCashout(ctx, bet.ChannelId, bet.ApiOrderNo)
	return buildCashoutResp(bet, 1), nil
}

func (s *CashoutService) cashoutPreAll(ctx context.Context, bet *servermodel.Bet, channel *servermodel.Channel) (*rttypes.CashoutResponse, error) {
	total := calcCashoutAmount(bet, bet.ManualCashoutMultiple, channel.MaxCashoutPerBet, true)
	if bet.FirstCashoutAmount > 0 && total < bet.FirstCashoutAmount {
		total = bet.FirstCashoutAmount
	}
	bet.CashedOutAmount = total
	bet.ManualCashoutTimes += 1

	if err := s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{
		servermodel.Bet_F_order_status:            servermodel.OrderStatusCashingOut,
		servermodel.Bet_F_cashed_out_amount:       bet.CashedOutAmount,
		servermodel.Bet_F_manual_cashout_multiple: bet.ManualCashoutMultiple,
		servermodel.Bet_F_Cashout_Times:           bet.ManualCashoutTimes,
	}); err != nil {
		return nil, err
	}

	partialCount := int8(0)
	if bet.FirstCashoutAmount > 0 || bet.FirstManualCashoutMultiple > 0 {
		partialCount = 1
	}
	err := s.Ctx.Settlement.BillPreMatch(ctx, settlement.BillRequest{
		ChannelID:      bet.ChannelId,
		UserID:         bet.UserId,
		OrderNo:        bet.ApiOrderNo,
		Currency:       bet.Currency,
		Amount:         domain.DBAmountToString(total),
		Metadata:       domain.BuildBetMetadata(bet),
		IsSystemReward: false,
	}, false, partialCount)
	if err != nil {
		_ = s.addCashoutRetry(ctx, bet)
		_ = s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{servermodel.Bet_F_order_status: servermodel.OrderStatusOverRetry})
		return nil, err
	}

	if err := s.markBetSettled(ctx, bet, total); err != nil {
		return nil, err
	}
	return buildCashoutResp(bet, 0), nil
}

func (s *CashoutService) markBetSettled(ctx context.Context, bet *servermodel.Bet, payout int64) error {
	if err := s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{
		servermodel.Bet_F_order_status:            servermodel.OrderStatusCashedout,
		servermodel.Bet_F_cashed_out_amount:       payout,
		servermodel.Bet_F_manual_cashout_multiple: bet.ManualCashoutMultiple,
	}); err != nil {
		return err
	}
	_ = s.Ctx.SnapshotStore.RemoveAutoCashout(ctx, bet.ChannelId, bet.ApiOrderNo)
	hot := &domain.BetHotState{
		BetID:        bet.Id,
		OrderNo:      bet.ApiOrderNo,
		ChannelID:    bet.ChannelId,
		TermID:       bet.TermId,
		UserID:       bet.UserId,
		GamePlay:     bet.GamePlay,
		OrderStatus:  servermodel.OrderStatusCashedout,
		Settled:      true,
		SettledAtMs:  time.Now().UnixMilli(),
		AutoTarget:   bet.AutoCashoutMultiple,
		ManualTarget: bet.ManualCashoutMultiple,
	}
	_ = s.Ctx.SnapshotStore.SaveBetHot(ctx, bet.ChannelId, hot)

	snap, err := s.Ctx.SnapshotStore.GetSnapshot(ctx, bet.ChannelId)
	if err == nil && snap != nil && snap.TermID == bet.TermId {
		snap.CashedAmt += payout
		if payout > 0 && (snap.MaxMultiple == 0 || bet.ManualCashoutMultiple > snap.MaxMultiple) {
			snap.MaxMultiple = bet.ManualCashoutMultiple
			snap.MaxCashedoutBetID = bet.Id
		}
		snap.Version = domain.NextVersion(snap.Version)
		_ = s.Ctx.SnapshotStore.SaveSnapshot(ctx, snap)
	}
	return nil
}

func (s *CashoutService) addCashoutRetry(ctx context.Context, bet *servermodel.Bet) error {
	old, _ := s.Ctx.RetryCashoutTaskModel.FindOneByBetId(ctx, bet.Id)
	if old != nil && old.Id > 0 {
		return s.Ctx.RetryCashoutTaskModel.Update(ctx, &servermodel.RetryCashoutTask{
			Id:            old.Id,
			BetId:         bet.Id,
			Status:        servermodel.RetryCashoutTask_Status_need,
			RetryNum:      old.RetryNum,
			NextRetryTime: time.Now().Unix() + 5,
		})
	}
	_, err := s.Ctx.RetryCashoutTaskModel.Insert(ctx, &servermodel.RetryCashoutTask{
		BetId:         bet.Id,
		Status:        servermodel.RetryCashoutTask_Status_need,
		RetryNum:      0,
		NextRetryTime: time.Now().Unix() + 5,
	})
	return err
}

func buildCashoutResp(bet *servermodel.Bet, isHalf int64) *rttypes.CashoutResponse {
	return &rttypes.CashoutResponse{
		Amount:     bet.Amount,
		ApiOrderNo: bet.ApiOrderNo,
		BetAtMutil: bet.BetAtMultiple / domain.MultipleTail,
		BetID:      bet.Id,
		BetType:    bet.BetType,
		CashoutAmt: bet.CashedOutAmount,
		Currency:   bet.Currency,
		Multipe:    bet.ManualCashoutMultiple / domain.MultipleTail,
		Type:       2,
		IsCashHalf: isHalf,
	}
}
