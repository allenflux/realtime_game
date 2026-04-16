package types

import "crash/realtime_game/domain"

// CreateBetRequest 是下单请求。
type CreateBetRequest struct {
	RequestID           string `json:"request_id,omitempty"`
	ChannelID           int64  `json:"channel_id"`
	ApiSysToken         string `json:"api_sys_token"`
	UserSeed            string `json:"user_seed,omitempty"`
	Amount              string `json:"amount"`
	Currency            string `json:"currency"`
	AutoCashoutMultiple string `json:"auto_cashout_multiple"`
	GamePlay            int64  `json:"game_play"`
}

// CreateBetResponse 是下单返回。
type CreateBetResponse struct {
	BetID       int64  `json:"bet_id"`
	ApiOrderNo  string `json:"api_order_no"`
	BetAtMutil  string `json:"bet_at_mutil"`
	ValidBetAmt string `json:"valid_bet_amt"`
}

// CashoutRequest 是手动兑现请求。
type CashoutRequest struct {
	RequestID      string `json:"request_id,omitempty"`
	ApiSysToken    string `json:"api_sys_token"`
	OrderNo        string `json:"order_no"`
	GamePlay       int64  `json:"game_play"`
	SettlementMode int64  `json:"settlement_mode"`
}

// CashoutResponse 是手动兑现返回。
type CashoutResponse struct {
	Amount     int64  `json:"amount"`
	ApiOrderNo string `json:"api_order_no"`
	BetAtMutil int64  `json:"bet_at_mutil"`
	BetID      int64  `json:"bet_id"`
	BetType    int64  `json:"bet_type"`
	CashoutAmt int64  `json:"cashout_amt"`
	Currency   string `json:"currency"`
	Multipe    int64  `json:"multipe"`
	Type       int64  `json:"type"`
	IsCashHalf int64  `json:"is_cash_half,omitempty"`
}

// CurrentRoundResponse 是当前局接口返回。
type CurrentRoundResponse struct {
	*domain.RoundSnapshot
	CurrentMultiple int64 `json:"current_multiple"`
}

type LeaderboardItem struct {
	OrderNo    string `json:"order_no"`
	UserName   string `json:"user_name"`
	Payout     string `json:"payout"`
	Multiplier string `json:"multiplier"`
}

type LeaderboardResponse struct {
	Items []LeaderboardItem `json:"items"`
}

type JackpotResponse struct {
	ChannelID      int64  `json:"channel_id"`
	GameName       string `json:"game_name"`
	Currency       string `json:"currency"`
	TermID         int64  `json:"term_id"`
	JackpotIn      string `json:"jackpot_in"`
	JackpotPrize1  string `json:"jackpot_prize_1"`
	JackpotPrize2  string `json:"jackpot_prize_2"`
	JackpotPrize3  string `json:"jackpot_prize_3"`
	JackpotBalance string `json:"jackpot_balance"`
}
