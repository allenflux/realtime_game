package service

import (
	"context"
	"crash/model/gmmodel"
	"crash/model/servermodel"
	appctx "crash/realtime_game/context"
	"crash/realtime_game/domain"
	"crash/realtime_game/settlement"
	rttypes "crash/realtime_game/types"
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/logx"
)

// Services 聚合所有服务。
type Services struct {
	Ctx   *appctx.AppContext
	Hooks *GameHooks
}

func New(ctx *appctx.AppContext) *Services {
	svc := &Services{Ctx: ctx}
	svc.Hooks = NewGameHooks(svc)
	return svc
}

var orderSeq atomic.Int64

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
	seq := orderSeq.Add(1) % 1000
	return fmt.Sprintf("%d%03d", time.Now().UnixNano(), seq)
}

func isRobotBet(bet *servermodel.Bet) bool {
	if bet == nil {
		return false
	}
	return bet.UserId >= robotUserBase || strings.HasPrefix(strings.ToLower(bet.UserName), "robot_") || strings.HasPrefix(strings.ToLower(bet.ApiOrderNo), "robot_")
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

func GetApiSysUserData(ctx context.Context, token string, app *appctx.AppContext) (*settlement.ApiSysGetUserData, error) {
	return getApiSysUserData(ctx, token, app)
}
func getApiSysUserData(ctx context.Context, token string, app *appctx.AppContext) (*settlement.ApiSysGetUserData, error) {
	userData, err := app.TokenUserStore.Get(ctx, token)
	if err != nil {
		return nil, err
	}
	if userData != nil {
		return userData, nil
	}

	userData, err = app.Settlement.GetUserInfoByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if err = app.TokenUserStore.Set(ctx, token, userData, 100); err != nil {
		logx.Errorf("cache apisys user info failed, err=%v", err)
	}

	return userData, nil
}
