package service

import (
	"context"
	"crash/model/servermodel"
	"crash/realtime_game/domain"
	"crash/realtime_game/settlement"
	"time"
)

// RetryService 负责兑现 / 退款失败后的重试。
type RetryService struct{ *Services }

func NewRetryService(s *Services) *RetryService { return &RetryService{Services: s} }

func (s *RetryService) RunCashoutRetry(ctx context.Context) error {
	tasks, err := s.Ctx.RetryCashoutTaskModel.GetPageNeedRetry(ctx, s.Ctx.Config.Runtime.RetryPageSize)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		_ = s.retryCashoutOne(ctx, task)
	}
	return nil
}

func (s *RetryService) retryCashoutOne(ctx context.Context, task *servermodel.RetryCashoutTask) error {
	bet, err := s.Ctx.BetModel.GetById(ctx, task.BetId)
	if err != nil || bet == nil {
		return err
	}
	var billErr error
	if bet.GamePlay == domain.GamePlayPreMatch {
		partialCount := int8(0)
		if bet.FirstCashoutAmount > 0 {
			partialCount = 1
		}
		billErr = s.Ctx.Settlement.BillPreMatch(ctx, settlement.BillRequest{
			ChannelID: bet.ChannelId,
			UserID:    bet.UserId,
			OrderNo:   bet.ApiOrderNo,
			Currency:  bet.Currency,
			Amount:    domain.DBAmountToString(bet.CashedOutAmount),
			Metadata:  domain.BuildBetMetadata(bet),
		}, false, partialCount)
	} else {
		billErr = s.Ctx.Settlement.BillRolling(ctx, settlement.BillRequest{
			ChannelID: bet.ChannelId,
			UserID:    bet.UserId,
			OrderNo:   bet.ApiOrderNo,
			Currency:  bet.Currency,
			Amount:    domain.DBAmountToString(bet.CashedOutAmount),
			Metadata:  domain.BuildBetMetadata(bet),
		})
	}
	if billErr != nil {
		task.RetryNum += 1
		task.NextRetryTime = time.Now().Unix() + 5
		if task.RetryNum >= 20 {
			task.Status = servermodel.RetryCashoutTask_Status_close
		}
		return s.Ctx.RetryCashoutTaskModel.Update(ctx, task)
	}
	_ = s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{servermodel.Bet_F_order_status: servermodel.OrderStatusCashedout})
	task.Status = servermodel.RetryCashoutTask_Status_suc
	return s.Ctx.RetryCashoutTaskModel.Update(ctx, task)
}

func (s *RetryService) RunRefundRetry(ctx context.Context) error {
	tasks, err := s.Ctx.RetryRefundTaskModel.GetPageNeedRetry(ctx, s.Ctx.Config.Runtime.RetryPageSize)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		_ = s.retryRefundOne(ctx, task)
	}
	return nil
}

func (s *RetryService) retryRefundOne(ctx context.Context, task *servermodel.RetryRefundTask) error {
	bet, err := s.Ctx.BetModel.GetById(ctx, task.BetId)
	if err != nil || bet == nil {
		return err
	}
	if err := s.Ctx.Settlement.Refund(ctx, settlement.RefundRequest{
		ChannelID: bet.ChannelId,
		OrderNo:   bet.ApiOrderNo,
		Metadata:  domain.BuildBetMetadata(bet),
	}); err != nil {
		task.RetryNum += 1
		task.NextRetryTime = time.Now().Unix() + 5
		if task.RetryNum >= 20 {
			task.Status = servermodel.RetryRefundTask_Status_close
		}
		return s.Ctx.RetryRefundTaskModel.Update(ctx, task)
	}
	_ = s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{servermodel.Bet_F_order_status: servermodel.OrderStatusRefunded})
	task.Status = servermodel.RetryRefundTask_Status_suc
	return s.Ctx.RetryRefundTaskModel.Update(ctx, task)
}
