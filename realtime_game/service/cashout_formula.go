package service

import (
	"crash/model/servermodel"
	"crash/realtime_game/domain"

	"github.com/shopspring/decimal"
)

// calcCashoutAmount 统一计算兑现金额。
// 返回值单位是 *10000。
func calcCashoutAmount(bet *servermodel.Bet, currentComparable int64, maxCashoutPerBet int64, manual bool) int64 {
	multiple := bet.AutoCashoutMultiple
	if manual && bet.ManualCashoutMultiple > 0 && bet.ManualCashoutMultiple < multiple {
		multiple = bet.ManualCashoutMultiple
	}
	if currentComparable > 0 {
		multiple = currentComparable
	}

	// 赛前盘：金额 * 兑现倍数。
	if bet.BetType == servermodel.BET_TYPE_pre {
		amount := bet.Amount
		// 第二次赛前兑现时，只补另外一半。
		if bet.FirstCashoutAmount > 0 || bet.FirstManualCashoutMultiple > 0 {
			amount = amount / 2
		}
		return decimal.NewFromInt(amount).
			Div(domain.AmountScaleDec).
			Mul(decimal.NewFromInt(multiple).Div(domain.MultipleScaleDec).Div(domain.MultipleTailDec).Truncate(2)).
			Mul(domain.AmountScaleDec).
			IntPart()
	}

	// 滚盘：按旧项目的“奖金倍数 = 当前倍数 / 投注时倍数”规则。
	amt := decimal.NewFromInt(bet.Amount).Div(domain.AmountScaleDec)
	value := amt.Mul(decimal.NewFromInt(multiple).Div(decimal.NewFromInt(bet.BetAtMultiple)).Truncate(2)).Mul(domain.AmountScaleDec).IntPart()
	if maxCashoutPerBet > 0 {
		maxDB := maxCashoutPerBet * domain.AmountScale
		if value > maxDB {
			value = maxDB
		}
	}
	return value
}
