package service

import (
	"context"
	"crash/model/servermodel"
	"crash/realtime_game/domain"
	"crash/realtime_game/settlement"
	rttypes "crash/realtime_game/types"
	"errors"
	"fmt"
	"strings"
	"time"
)

// PlaceBetService 负责下单。
type PlaceBetService struct{ *Services }

func NewPlaceBetService(s *Services) *PlaceBetService { return &PlaceBetService{Services: s} }

func (s *PlaceBetService) Place(ctx context.Context, req *rttypes.CreateBetRequest) (*rttypes.CreateBetResponse, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}
	if req.ChannelID <= 0 || req.UserID <= 0 {
		return nil, errors.New("channel_id 或 user_id 非法")
	}
	if req.Amount == "" || req.Currency == "" || req.AutoCashoutMultiple == "" {
		return nil, errors.New("amount/currency/auto_cashout_multiple 不能为空")
	}

	// 幂等：优先使用调用方传入 request_id。
	if req.RequestID != "" {
		if value, _ := s.Ctx.SnapshotStore.GetIdempotent(ctx, "bet:"+req.RequestID); value != "" {
			bet, err := s.Ctx.BetModel.GetByApiOrderNo(ctx, value)
			if err == nil && bet != nil {
				return buildCreateBetResponse(bet), nil
			}
		}
	}

	snap, err := s.Ctx.SnapshotStore.GetSnapshot(ctx, req.ChannelID)
	if err != nil {
		return nil, err
	}
	if snap == nil {
		return nil, errors.New("当前局不存在")
	}
	if snap.State == domain.RoundStateCrashed || snap.State == domain.RoundStateClosed {
		return nil, errors.New("本局已结束")
	}
	if snap.State != domain.RoundStatePreStart && req.GamePlay == domain.GamePlayPreMatch {
		return nil, errors.New("赛前盘当前不可下注")
	}

	channel, err := s.Ctx.ChannelModel.FindOne(ctx, req.ChannelID)
	if err != nil {
		return nil, err
	}

	limit, err := loadCurrencyLimit(ctx, s.Services, req.ChannelID, req.Currency)
	if err != nil || limit == nil {
		return nil, fmt.Errorf("币种限额未配置: %w", err)
	}

	amountDB, err := domain.ParseAmountToDB(req.Amount)
	if err != nil {
		return nil, errors.New("amount 格式非法")
	}
	if amountDB <= 0 {
		return nil, errors.New("amount 必须大于 0")
	}

	// 这里延续旧系统习惯：限额表直接是自然金额，不是 *10000。
	if amountDB < limit.MinBet*domain.AmountScale {
		return nil, errors.New("投注金额小于最小限额")
	}
	if amountDB > limit.MaxBet*domain.AmountScale {
		return nil, errors.New("投注金额大于最大限额")
	}

	autoField, err := domain.ParseUserMultipleToBetField(req.AutoCashoutMultiple)
	if err != nil {
		return nil, errors.New("auto_cashout_multiple 格式非法")
	}
	if channel.MaxCashoutMultiple > 0 {
		maxField := channel.MaxCashoutMultiple * domain.MultipleScale * domain.MultipleTail
		if autoField > maxField {
			autoField = maxField
		}
	}

	betType := chooseBetType(snap)
	serviceFee := calcServiceFee(channel, amountDB, betType)
	amountAfterFee := amountDB - serviceFee
	if amountAfterFee <= 0 {
		return nil, errors.New("扣除服务费后金额非法")
	}

	orderNo := nextOrderNo()
	if req.RequestID != "" {
		lockOK, err := s.Ctx.SnapshotStore.AcquireOpLock(ctx, "bet:"+req.RequestID, 10)
		if err != nil {
			return nil, err
		}
		if !lockOK {
			return nil, errors.New("请求正在处理中")
		}
	}

	// 下注写 DB：先写 creating，再扣款，再改 created。
	bet := &servermodel.Bet{
		ApiOrderNo:                 orderNo,
		ChannelId:                  req.ChannelID,
		TermId:                     snap.TermID,
		UserId:                     req.UserID,
		UserName:                   req.UserName,
		UserSeed:                   strings.TrimSpace(req.UserSeed),
		BetType:                    betType,
		Amount:                     amountAfterFee,
		Currency:                   req.Currency,
		AutoCashoutMultiple:        autoField,
		ManualCashoutMultiple:      0,
		BetAtMultiple:              domain.CurrentMultipleToBetField(s.currentBetMultiple(snap)),
		ServiceFee:                 serviceFee,
		Rake:                       channel.Rake,
		RakeAmt:                    amountDB * channel.Rake / domain.AmountScale,
		CashedOutAmount:            0,
		OrderStatus:                servermodel.OrderStatusCreating,
		InRetry:                    servermodel.BET_IN_RETRY_no,
		Ctime:                      time.Now().Unix(),
		GamePlay:                   req.GamePlay,
		ManualCashoutTimes:         0,
		FirstCashoutAmount:         0,
		FirstManualCashoutMultiple: 0,
		IsCashoutAmountMerged:      0,
	}
	if bet.UserSeed == "" {
		bet.UserSeed = randUserSeed()
	}

	res, err := s.Ctx.BetModel.Insert(ctx, bet)
	if err != nil {
		return nil, err
	}
	betID, _ := res.LastInsertId()
	bet.Id = betID

	// 外部扣款金额按原始自然金额传递。
	_, err = s.Ctx.Settlement.Deduct(ctx, settlement.DeductRequest{
		ChannelID:      req.ChannelID,
		OrderNo:        orderNo,
		UserID:         req.UserID,
		Currency:       req.Currency,
		Amount:         req.Amount,
		Metadata:       domain.BuildBetMetadata(bet),
		IsSystemReward: false,
	})
	if err != nil {
		// 扣款失败时，订单置失败；同时保守补打一笔退款。
		_ = s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{
			servermodel.Bet_F_order_status: servermodel.OrderStatusCreationFailed,
		})
		_ = s.Ctx.Settlement.Refund(ctx, settlement.RefundRequest{
			ChannelID: req.ChannelID,
			OrderNo:   orderNo,
			Metadata:  domain.BuildBetMetadata(bet),
		})
		return nil, err
	}

	// 改成已创建。
	if err := s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{
		servermodel.Bet_F_order_status: servermodel.OrderStatusCreated,
	}); err != nil {
		return nil, err
	}
	bet.OrderStatus = servermodel.OrderStatusCreated

	// 更新当前局聚合统计。
	snap.TotalBetAmt += amountDB
	snap.FeeAmt += serviceFee
	snap.RakeAmt += bet.RakeAmt
	snap.BounsPool += amountDB - bet.RakeAmt
	snap.Version = domain.NextVersion(snap.Version)
	if err := s.Ctx.SnapshotStore.SaveSnapshot(ctx, snap); err != nil {
		return nil, err
	}

	// 保存热状态并加入自动兑现队列。
	if err := s.Ctx.SnapshotStore.SaveBetHot(ctx, req.ChannelID, &domain.BetHotState{
		BetID:       bet.Id,
		OrderNo:     bet.ApiOrderNo,
		ChannelID:   bet.ChannelId,
		TermID:      bet.TermId,
		UserID:      bet.UserId,
		GamePlay:    bet.GamePlay,
		OrderStatus: bet.OrderStatus,
		Settled:     false,
		AutoTarget:  bet.AutoCashoutMultiple,
	}); err != nil {
		return nil, err
	}
	if err := s.Ctx.SnapshotStore.EnqueueAutoCashout(ctx, req.ChannelID, bet.ApiOrderNo, bet.AutoCashoutMultiple); err != nil {
		return nil, err
	}

	if req.RequestID != "" {
		_ = s.Ctx.SnapshotStore.MarkIdempotent(ctx, "bet:"+req.RequestID, 3600, orderNo)
	}
	return buildCreateBetResponse(bet), nil
}

func (s *PlaceBetService) currentBetMultiple(snap *domain.RoundSnapshot) int64 {
	switch snap.State {
	case domain.RoundStateFlying:
		nowMs := time.Now().UnixMilli()
		v := domain.CalcCurrentMultiple(snap.IncNum, snap.FlyingStartAtMs, nowMs)
		if v > snap.CrashMultiple {
			return snap.CrashMultiple
		}
		return v
	default:
		return domain.MultipleScale
	}
}
