package domain

// RoundSnapshot 是当前局在 Redis 中保存的权威热状态。
type RoundSnapshot struct {
	ChannelID int64  `json:"channel_id"`
	GameCode  string `json:"game_code"`
	TermID    int64  `json:"term_id"`
	TermDBID  int64  `json:"term_db_id"`

	State   string `json:"state"`
	Version int64  `json:"version"`

	OpenAtMs        int64   `json:"open_at_ms"`
	BetCloseAtMs    int64   `json:"bet_close_at_ms"`
	FlyingStartAtMs int64   `json:"flying_start_at_ms"`
	CrashAtMs       int64   `json:"crash_at_ms"`
	CloseAtMs       int64   `json:"close_at_ms"`
	CrashedAtMs     int64   `json:"crashed_at_ms"`
	ClosedAtMs      int64   `json:"closed_at_ms"`
	IncNum          float64 `json:"inc_num"`

	// 当前局最终爆点，单位 *100。
	CrashMultiple int64 `json:"crash_multiple"`
	// 当前局最终 hash。
	Hash string `json:"hash"`

	// 这些值用于映射到 crash_term，保证字段语义兼容。
	IsControl         int64  `json:"is_control"`
	IsCrashed         int64  `json:"is_crashed"`
	BounsPoolStart    int64  `json:"bouns_pool_start"`
	BounsPool         int64  `json:"bouns_pool"`
	TotalBetAmt       int64  `json:"total_bet_amt"`
	FeeAmt            int64  `json:"fee_amt"`
	RakeAmt           int64  `json:"rake_amt"`
	CashedAmt         int64  `json:"cashed_amt"`
	CtrlCashedAmt     int64  `json:"ctrl_cashed_amt"`
	ProfitAmt         int64  `json:"profit_amt"`
	UserProfitCorrect int64  `json:"user_profit_correct"`
	BreakPayoutRate   int64  `json:"break_payout_rate"`
	MaxCashedoutBetID int64  `json:"max_cashedout_bet_id"`
	MaxMultiple       int64  `json:"max_multiple"`
	Seed              string `json:"seed"`

	ServerTimeMs int64 `json:"server_time_ms,omitempty"`
}

// BetHotState 是当前局中一笔订单的热状态。
type BetHotState struct {
	BetID        int64  `json:"bet_id"`
	OrderNo      string `json:"order_no"`
	ChannelID    int64  `json:"channel_id"`
	TermID       int64  `json:"term_id"`
	UserID       int64  `json:"user_id"`
	GamePlay     int64  `json:"game_play"`
	OrderStatus  int64  `json:"order_status"`
	Settled      bool   `json:"settled"`
	SettledAtMs  int64  `json:"settled_at_ms"`
	AutoTarget   int64  `json:"auto_target"`
	ManualTarget int64  `json:"manual_target"`
}
