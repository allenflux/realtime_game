package service

import (
	"context"
	"crash/model/gmmodel"
	"crash/model/servermodel"
	appctx "crash/realtime_game/context"
	"crash/realtime_game/domain"
	rttypes "crash/realtime_game/types"
	"fmt"
	"math/rand"

	"crash/realtime_game/common"

	"github.com/shopspring/decimal"
)

// Services 聚合所有服务。
type Services struct {
	Ctx *appctx.AppContext
}

func New(ctx *appctx.AppContext) *Services { return &Services{Ctx: ctx} }

func randUserSeed() string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	n := 8 + rand.Intn(8)
	out := make([]rune, n)
	for i := range out {
		out[i] = letters[rand.Intn(len(letters))]
	}
	return string(out)
}

func nextOrderNo() string {
	return fmt.Sprintf("%d", common.GenerateId())
}

func chooseBetType(snap *domain.RoundSnapshot) int64 {
	if snap.State == domain.RoundStatePreStart {
		return servermodel.BET_TYPE_pre
	}
	return servermodel.BET_TYPE_ing
}

func calcServiceFee(channel *servermodel.Channel, amountDB int64, betType int64) int64 {
	if betType == servermodel.BET_TYPE_pre {
		return 0
	}
	amountDec := decimal.NewFromInt(amountDB).Div(domain.AmountScaleDec)
	return decimal.NewFromInt(channel.ServiceFee).Div(domain.AmountScaleDec).Mul(amountDec).Truncate(2).Mul(domain.AmountScaleDec).IntPart()
}

func loadCurrencyLimit(ctx context.Context, svc *Services, channelID int64, currency string) (*gmmodel.CurrencyLimit, error) {
	return svc.Ctx.CurrencyLimitModel.FindByCidAndCurrency(ctx, channelID, currency)
}

func buildCreateBetResponse(bet *servermodel.Bet) *rttypes.CreateBetResponse {
	return &rttypes.CreateBetResponse{
		BetID:       bet.Id,
		ApiOrderNo:  bet.ApiOrderNo,
		BetAtMutil:  domain.BetFieldToHumanMultiple(bet.BetAtMultiple),
		ValidBetAmt: domain.DBAmountToString(bet.Amount),
	}
}
