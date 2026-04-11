package domain

import (
	"crash/model/servermodel"
	"encoding/json"
)

// Metadata 是传给 ApiSys 的扩展信息。
type Metadata struct {
	ID              int64  `json:"id"`
	UserId          int64  `json:"user_id"`
	UserName        string `json:"user_name"`
	Currency        string `json:"currency"`
	CreateTime      int64  `json:"create_time"`
	Amount          string `json:"amount"`
	Multiple        string `json:"multiple"`
	BetAtMutil      string `json:"bet_at_mutil"`
	CashedOutAmount string `json:"cashed_out_amount"`
	OrderStatus     int64  `json:"order_status"`
	BetType         int64  `json:"bet_type"`
	ServiceFee      string `json:"service_fee"`
}

// BuildBetMetadata 生成标准注单 metadata。
func BuildBetMetadata(bet *servermodel.Bet) string {
	return BuildBetMetadataWithHalf(bet, false)
}

// BuildBetMetadataWithHalf 生成包含半兑视图的 metadata。
func BuildBetMetadataWithHalf(bet *servermodel.Bet, half bool) string {
	multiple := bet.AutoCashoutMultiple
	if bet.ManualCashoutMultiple > 0 {
		multiple = bet.ManualCashoutMultiple
	}
	multipleStr := BetFieldToHumanMultiple(multiple)
	if bet.CashedOutAmount == 0 {
		multipleStr = "-"
	}

	amount := bet.Amount
	if half {
		amount = amount / 2
	}

	meta := Metadata{
		ID:              bet.Id,
		UserId:          bet.UserId,
		UserName:        bet.UserName,
		Currency:        bet.Currency,
		CreateTime:      bet.Ctime,
		Amount:          DBAmountToString(amount),
		Multiple:        multipleStr,
		BetAtMutil:      BetFieldToHumanMultiple(bet.BetAtMultiple),
		CashedOutAmount: DBAmountToString(bet.CashedOutAmount),
		OrderStatus:     bet.OrderStatus,
		BetType:         bet.BetType,
		ServiceFee:      DBAmountToString(bet.ServiceFee),
	}
	buf, _ := json.Marshal(meta)
	return string(buf)
}
