package domain

import "github.com/shopspring/decimal"

const (
	// 金额缩放位，数据库金额字段统一是 *10000。
	AmountScale = int64(10000)
	// 倍数缩放位，当前局倍数统一是 *100。
	MultipleScale = int64(100)
	// 历史系统里 bet 相关倍数字段统一多乘一个 10000。
	MultipleTail = int64(10000)
)

var (
	AmountScaleDec   = decimal.NewFromInt(AmountScale)
	MultipleScaleDec = decimal.NewFromInt(MultipleScale)
	MultipleTailDec  = decimal.NewFromInt(MultipleTail)
)

const (
	// 玩法常量。
	GamePlayRollingPlate = 0
	GamePlayPreMatch     = 1

	// 赛前盘兑现模式。
	CashoutHalf = 0
	CashoutAll  = 1
)

const (
	// 当前局状态。
	RoundStatePreStart = "pre_start"
	RoundStateStarting = "starting"
	RoundStateFlying   = "flying"
	RoundStateCrashed  = "crashed"
	RoundStateClosed   = "closed"
)
