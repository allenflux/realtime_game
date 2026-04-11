package service

import (
	"context"
	"crash/model/servermodel"
	"crash/realtime_game/domain"
	"crash/realtime_game/settlement"
	"database/sql"
	"time"
)

// CloseRoundService 负责收局。
type CloseRoundService struct{ *Services }

func NewCloseRoundService(s *Services) *CloseRoundService { return &CloseRoundService{Services: s} }

func (s *CloseRoundService) Finalize(ctx context.Context, snap *domain.RoundSnapshot) error {
	//if snap == nil || snap.State == domain.RoundStateClosed {
	//	return nil
	//}
	//
	//// 先把局置为 closed，防止重复推进。
	//snap.State = domain.RoundStateClosed
	//snap.IsCrashed = servermodel.CrashTermIsCrashedYes
	//snap.CrashedAtMs = snap.CrashAtMs
	//snap.ClosedAtMs = time.Now().UnixMilli()
	//snap.Version = domain.NextVersion(snap.Version)
	//if err := s.Ctx.SnapshotStore.SaveSnapshot(ctx, snap); err != nil {
	//	return err
	//}

	if err := s.finalizePendingBets(ctx, snap); err != nil {
		return err
	}

	// 投影到 crash_term 表。
	return s.Ctx.CrashTermModel.UpdateById(ctx, snap.TermDBID, map[string]any{
		servermodel.CrashTerm_F_multiple:             snap.CrashMultiple,
		servermodel.CrashTerm_F_is_crashed:           servermodel.CrashTermIsCrashedYes,
		servermodel.CrashTerm_F_term_hash:            snap.Hash,
		servermodel.CrashTerm_F_total_bet_amt:        snap.TotalBetAmt,
		servermodel.CrashTerm_F_fee_amt:              snap.FeeAmt,
		servermodel.CrashTerm_F_rake_amt:             snap.RakeAmt,
		servermodel.CrashTerm_F_cashed_amt:           snap.CashedAmt,
		servermodel.CrashTerm_F_ctrl_cashed_amt:      snap.CtrlCashedAmt,
		servermodel.CrashTerm_F_bouns_pool:           snap.BounsPool,
		servermodel.CrashTerm_F_bouns_pool_start:     snap.BounsPoolStart,
		servermodel.CrashTerm_F_profit_amt:           snap.ProfitAmt,
		servermodel.CrashTerm_F_user_profit_correct:  snap.UserProfitCorrect,
		servermodel.CrashTerm_F_break_payout_rate:    snap.BreakPayoutRate,
		servermodel.CrashTerm_F_max_cashedout_bet_id: snap.MaxCashedoutBetID,
		servermodel.CrashTerm_F_max_multiple:         snap.MaxMultiple,
		servermodel.CrashTerm_F_pre_start_time:       snap.OpenAtMs / 1000,
		servermodel.CrashTerm_F_starting_time:        snap.BetCloseAtMs / 1000,
		servermodel.CrashTerm_F_flying_time:          snap.FlyingStartAtMs / 1000,
		servermodel.CrashTerm_F_crashed_time:         snap.CrashAtMs / 1000,
		servermodel.CrashTerm_F_seed:                 snap.Seed,
	})
}

func (s *CloseRoundService) finalizePendingBets(ctx context.Context, snap *domain.RoundSnapshot) error {
	bets, err := s.Ctx.BetModel.GetAllBetsByItemId(ctx, snap.ChannelID, snap.TermID)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if len(bets) == 0 {
		return nil
	}

	zeroReqs := make([]settlement.BillRequest, 0)
	needClose := make([]int64, 0)
	for _, bet := range bets {
		// 未兑现订单做 0 元收口。
		if bet.OrderStatus == servermodel.OrderStatusCreated {
			zeroReqs = append(zeroReqs, settlement.BillRequest{
				ChannelID: bet.ChannelId,
				UserID:    bet.UserId,
				OrderNo:   bet.ApiOrderNo,
				Currency:  bet.Currency,
				Amount:    "0",
				Metadata:  domain.BuildBetMetadata(bet),
			})
			needClose = append(needClose, bet.Id)
			continue
		}
		// 赛前半兑成功后，需要再打一笔 0 元 final settle，保持与旧系统对外语义一致。
		if bet.GamePlay == domain.GamePlayPreMatch &&
			bet.OrderStatus == servermodel.OrderStatusCashedout &&
			bet.FirstCashoutAmount > 0 &&
			bet.FirstCashoutAmount == bet.CashedOutAmount &&
			bet.FirstManualCashoutMultiple == bet.ManualCashoutMultiple {
			zeroReqs = append(zeroReqs, settlement.BillRequest{
				ChannelID: bet.ChannelId,
				UserID:    bet.UserId,
				OrderNo:   bet.ApiOrderNo,
				Currency:  bet.Currency,
				Amount:    "0",
				Metadata:  domain.BuildBetMetadata(bet),
			})
		}
	}

	if len(zeroReqs) > 0 {
		suc, fail, err := s.Ctx.Settlement.BatchBill(ctx, zeroReqs)
		if err != nil {
			_ = err
		}
		sucMap := map[string]struct{}{}
		for _, no := range suc {
			sucMap[no] = struct{}{}
		}
		failMap := map[string]struct{}{}
		for _, no := range fail {
			failMap[no] = struct{}{}
		}
		for _, bet := range bets {
			if _, ok := sucMap[bet.ApiOrderNo]; ok {
				_ = s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{servermodel.Bet_F_order_status: servermodel.OrderStatusCashedout})
			}
			if _, ok := failMap[bet.ApiOrderNo]; ok {
				_ = s.Ctx.BetModel.UpdateById(ctx, bet.Id, map[string]any{servermodel.Bet_F_order_status: servermodel.OrderStatusOverRetry})
				_, _ = s.Ctx.RetryCashoutTaskModel.Insert(ctx, &servermodel.RetryCashoutTask{
					BetId:         bet.Id,
					Status:        servermodel.RetryCashoutTask_Status_need,
					RetryNum:      0,
					NextRetryTime: time.Now().Unix() + 5,
				})
			}
		}
	}
	return nil
}

func (s *CloseRoundService) MarkClosed(ctx context.Context, round *domain.RoundSnapshot) error {
	round.State = domain.RoundStateClosed
	round.Version = domain.NextVersion(round.Version)
	round.ClosedAtMs = time.Now().UnixMilli()

	return s.Ctx.SnapshotStore.SaveSnapshot(ctx, round)
}
