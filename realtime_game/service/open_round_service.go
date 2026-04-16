package service

import (
	"context"
	"crash/model/servermodel"
	"crash/realtime_game/domain"
	"time"
)

// OpenRoundService 负责开新局。
type OpenRoundService struct{ *Services }

func NewOpenRoundService(s *Services) *OpenRoundService { return &OpenRoundService{Services: s} }

// OpenIfNeeded 在没有可用当前局时创建新局。
func (s *OpenRoundService) OpenIfNeeded(ctx context.Context, channel *servermodel.Channel) (*domain.RoundSnapshot, error) {
	current, err := s.Ctx.SnapshotStore.GetSnapshot(ctx, channel.Id)
	if err != nil {
		return nil, err
	}
	if current != nil && current.State != domain.RoundStateClosed {
		return current, nil
	}

	termID, err := s.Ctx.CrashTermModel.GetNextTermNo(ctx, channel.Id)
	if err != nil {
		return nil, err
	}

	nowMs := time.Now().UnixMilli()
	ctime := time.Now().Unix()

	applyRiskControl(channel)
	crashMultiple := channel.NextRandMultiple
	if crashMultiple <= 0 || channel.NextRandMultipleUsed == servermodel.CHANNEL_next_rand_multiple_used_yes {
		crashMultiple = domain.RandCrashMultiple(channel.Divisor, channel.CtrlCoef, channel.MaxCashoutMultiple)
	}
	if crashMultiple <= 0 {
		crashMultiple = domain.MultipleScale
	}

	crashDurationMs := domain.CalcCrashDurationMs(channel.IncNum, crashMultiple)

	term := &servermodel.CrashTerm{
		TermId:            termID,
		ChannelId:         channel.Id,
		Multiple:          domain.MultipleScale,
		IsControl:         servermodel.CrashTerm_Is_Control_no,
		IsCrashed:         servermodel.CrashTermIsCrashedNo,
		BounsPoolStart:    0,
		BounsPool:         0,
		TermHash:          "",
		Sha512Seed:        "",
		TotalBetAmt:       0,
		FeeAmt:            0,
		RakeAmt:           0,
		CashedAmt:         0,
		CtrlCashedAmt:     0,
		ManualSquib:       servermodel.ManualSquib_no,
		ManualSquibState:  servermodel.CrashTerm_Manual_Squib_State_null,
		BreakPayoutRate:   0,
		ProfitAmt:         0,
		MaxCashedoutBetId: 0,
		MaxMultiple:       0,
		PreStartTime:      nowMs / 1000,
		StartingTime:      0,
		FlyingTime:        0,
		CrashedTime:       0,
		Ctime:             ctime,
	}

	res, err := s.Ctx.CrashTermModel.Insert(ctx, term)
	if err != nil {
		return nil, err
	}
	termDBID, _ := res.LastInsertId()

	snap := &domain.RoundSnapshot{
		ChannelID: channel.Id,
		GameCode:  channel.GameName,
		TermID:    termID,
		TermDBID:  termDBID,
		State:     domain.RoundStatePreStart,
		Version:   1,

		OpenAtMs:          nowMs,
		BetCloseAtMs:      nowMs + s.Ctx.Config.Runtime.PreStartMs,
		FlyingStartAtMs:   nowMs + s.Ctx.Config.Runtime.PreStartMs + s.Ctx.Config.Runtime.StartingMs,
		CrashAtMs:         nowMs + s.Ctx.Config.Runtime.PreStartMs + s.Ctx.Config.Runtime.StartingMs + crashDurationMs,
		CloseAtMs:         nowMs + s.Ctx.Config.Runtime.PreStartMs + s.Ctx.Config.Runtime.StartingMs + crashDurationMs + s.Ctx.Config.Runtime.CloseDelayMs,
		IncNum:            channel.IncNum,
		CrashMultiple:     crashMultiple,
		Hash:              domain.BuildTermHash(termID, ctime),
		IsControl:         term.IsControl,
		IsCrashed:         term.IsCrashed,
		BounsPoolStart:    0,
		BounsPool:         0,
		TotalBetAmt:       0,
		FeeAmt:            0,
		RakeAmt:           0,
		CashedAmt:         0,
		CtrlCashedAmt:     0,
		ProfitAmt:         0,
		UserProfitCorrect: 0,
		BreakPayoutRate:   0,
		MaxCashedoutBetID: 0,
		MaxMultiple:       0,
		Seed:              "",
	}

	if err := s.Ctx.SnapshotStore.SaveSnapshot(ctx, snap); err != nil {
		return nil, err
	}
	return snap, nil
}
