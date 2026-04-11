package gmmodel

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	GameDayStatis_F_id                       = "id"
	GameDayStatis_F_client_id                = "client_id"
	GameDayStatis_F_game_name                = "game_name"
	GameDayStatis_F_term_id                  = "term_id"
	GameDayStatis_F_player_num               = "player_num"
	GameDayStatis_F_total_bet_amt            = "total_bet_amt"
	GameDayStatis_F_game_result              = "game_result"
	GameDayStatis_F_pre_bet_num              = "pre_bet_num"
	GameDayStatis_F_ing_bet_num              = "ing_bet_num"
	GameDayStatis_F_pre_bet_amt              = "pre_bet_amt"
	GameDayStatis_F_ing_bet_amt              = "ing_bet_amt"
	GameDayStatis_F_fee_amt                  = "fee_amt"
	GameDayStatis_F_pre_cashout_num          = "pre_cashout_num"
	GameDayStatis_F_ing_cashout_num          = "ing_cashout_num"
	GameDayStatis_F_pre_cashout_amt          = "pre_cashout_amt"
	GameDayStatis_F_ing_cashout_amt          = "ing_cashout_amt"
	GameDayStatis_F_total_bonus              = "total_bonus"
	GameDayStatis_F_bonus_rate               = "bonus_rate"
	GameDayStatis_F_ctrl_status              = "ctrl_status"
	GameDayStatis_F_record_date              = "record_date"
	GameDayStatis_F_create_time              = "create_time"
	GameDayStatis_F_currency                 = "currency"
	GameDayStatis_F_total_player_num         = "total_player_num"
	GameDayStatis_F_total_valid_bet_amt      = "total_valid_bet_amt"
	GameDayStatis_F_prize_pool_add_amt       = "prize_pool_add_amt"
	GameDayStatis_F_prize_pool_payout_amt    = "prize_pool_payout_amt"
	GameDayStatis_F_prize_pool_valid_bet_amt = "prize_pool_valid_bet_amt"
	GameDayStatis_F_prize_pool_rate          = "prize_pool_rate"
	GameDayStatis_F_match_bet_player_num     = "match_bet_player_num"
	GameDayStatis_F_match_valid_bet_amt      = "match_valid_bet_amt"
	GameDayStatis_F_match_profit_amt         = "match_profit_amt"
	GameDayStatis_F_match_kill_rate          = "match_kill_rate"
	GameDayStatis_F_rolling_bet_player_num   = "rolling_bet_player_num"
	GameDayStatis_F_rolling_valid_bet_amt    = "rolling_valid_bet_amt"
	GameDayStatis_F_rolling_profit_amt       = "rolling_profit_amt"
	GameDayStatis_F_rolling_kill_rate        = "rolling_kill_rate"
	GameDayStatis_F_total_profit_amt         = "total_profit_amt"
	GameDayStatis_F_total_kill_rate          = "total_kill_rate"
)

const (
	GameDayStatis_CTRL_STATUS_normal = 1
	GameDayStatis_CTRL_STATUS_fail   = 2
	GameDayStatis_CTRL_STATUS_win    = 3
)

var _ GameDayStatisModel = (*customGameDayStatisModel)(nil)

type (
	// GameDayStatisModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGameDayStatisModel.
	GameDayStatisModel interface {
		gameDayStatisModel
		withSession(session sqlx.Session) GameDayStatisModel
		GetPage(ctx context.Context, termId int64, channelIds []int64, ctrlStatus int64, startDate, endDate int, gameName string, pageSize int, start int) ([]*GameDayStatis, error)
		GetPageNum(ctx context.Context, termId int64, channelIds []int64, ctrlStatus int64, startDate, endDate int, gameName string) (int, error)
		GetPageWithFilters(ctx context.Context, req GameStatisFilterReq, pageSize int, start int) ([]*GameDayStatis, error)
		GetPageNumWithFilters(ctx context.Context, req GameStatisFilterReq) (int, error)
	}

	// GameStatisFilterReq 游戏统计筛选条件
	GameStatisFilterReq struct {
		TermId     int64
		ChannelIds []int64
		CtrlStatus int64
		StartDate  int
		EndDate    int
		GameName   string
		Currency   string
		// 筛选条件
		TotalBetMin     int64 // 总投注最小值
		TotalBetMax     int64 // 总投注最大值
		TotalCashoutMin int64 // 总兑现最小值
		TotalCashoutMax int64 // 总兑现最大值
		TotalProfitMin  int64 // 总盈亏最小值
		TotalProfitMax  int64 // 总盈亏最大值
		KillRateMin     int64 // 总杀率最小值
		KillRateMax     int64 // 总杀率最大值
		PercentMin      int64 // 百分比最小值
		PercentMax      int64 // 百分比最大值
	}

	customGameDayStatisModel struct {
		*defaultGameDayStatisModel
	}
)

// NewGameDayStatisModel returns a model for the database table.
func NewGameDayStatisModel(conn sqlx.SqlConn) GameDayStatisModel {
	return &customGameDayStatisModel{
		defaultGameDayStatisModel: newGameDayStatisModel(conn),
	}
}

func (m *customGameDayStatisModel) withSession(session sqlx.Session) GameDayStatisModel {
	return NewGameDayStatisModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customGameDayStatisModel) GetPage(ctx context.Context, termId int64, channelIds []int64, ctrlStatus int64, startDate, endDate int, gameName string, pageSize int, start int) (resp []*GameDayStatis, err error) {
	resp = make([]*GameDayStatis, 0)
	sqb := sq.Select("*").From(m.table)
	if termId > 0 {
		sqb = sqb.Where(sq.Eq{
			GameDayStatis_F_term_id: termId,
		})
	}
	if len(channelIds) > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_client_id: channelIds,
		})
	}
	if ctrlStatus > 0 {
		sqb = sqb.Where(sq.Eq{
			GameDayStatis_F_ctrl_status: ctrlStatus,
		})
	}
	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{
			GameDayStatis_F_game_name: gameName,
		})
	}
	if startDate > 0 {
		sqb = sqb.Where(sq.GtOrEq{
			GameDayStatis_F_record_date: startDate,
		})
	}
	if endDate > 0 {
		sqb = sqb.Where(sq.LtOrEq{
			GameDayStatis_F_record_date: endDate,
		})
	}

	sqlStr, sqlParams, _ := sqb.Offset(uint64(start)).Limit(uint64(pageSize)).ToSql()
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}

	return resp, err
}

// GetPageWithFilters 带筛选条件的分页查询
func (m *customGameDayStatisModel) GetPageWithFilters(ctx context.Context, req GameStatisFilterReq, pageSize int, start int) (resp []*GameDayStatis, err error) {
	resp = make([]*GameDayStatis, 0)
	sqb := sq.Select("*").From(m.table)

	// 基础条件
	if req.TermId > 0 {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_term_id: req.TermId})
	}
	if len(req.ChannelIds) > 0 {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_client_id: req.ChannelIds})
	}
	if req.CtrlStatus > 0 {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_ctrl_status: req.CtrlStatus})
	}
	if len(req.GameName) > 0 {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_game_name: req.GameName})
	}
	if len(req.Currency) > 0 && req.Currency != "0" {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_currency: req.Currency})
	}
	if req.StartDate > 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_record_date: req.StartDate})
	}
	if req.EndDate > 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_record_date: req.EndDate})
	}

	// 筛选条件
	if req.TotalBetMin > 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_total_bet_amt: req.TotalBetMin})
	}
	if req.TotalBetMax > 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_total_bet_amt: req.TotalBetMax})
	}

	// 总兑现区间 (总兑现 = 游戏前兑现 + 游戏中兑现)
	if req.TotalCashoutMin != 0 || req.TotalCashoutMax != 0 {
		cashoutExpr := "(pre_cashout_amt + ing_cashout_amt)"
		if req.TotalCashoutMin != 0 {
			sqb = sqb.Where(sq.Expr(cashoutExpr+" >= ?", req.TotalCashoutMin))
		}
		if req.TotalCashoutMax != 0 {
			sqb = sqb.Where(sq.Expr(cashoutExpr+" <= ?", req.TotalCashoutMax))
		}
	}

	// 总盈亏区间
	if req.TotalProfitMin != 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_total_profit_amt: req.TotalProfitMin})
	}
	if req.TotalProfitMax != 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_total_profit_amt: req.TotalProfitMax})
	}

	// 总杀率区间
	if req.KillRateMin != 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_total_kill_rate: req.KillRateMin})
	}
	if req.KillRateMax != 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_total_kill_rate: req.KillRateMax})
	}

	// 百分比区间 (这里用返奖率)
	if req.PercentMin != 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_bonus_rate: req.PercentMin})
	}
	if req.PercentMax != 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_bonus_rate: req.PercentMax})
	}

	sqlStr, sqlParams, _ := sqb.Offset(uint64(start)).Limit(uint64(pageSize)).ToSql()
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}

	return resp, err
}

// GetPageNumWithFilters 带筛选条件的总数查询
func (m *customGameDayStatisModel) GetPageNumWithFilters(ctx context.Context, req GameStatisFilterReq) (int, error) {
	sqb := sq.Select("count(*) as num").From(m.table)

	// 基础条件
	if req.TermId > 0 {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_term_id: req.TermId})
	}
	if len(req.ChannelIds) > 0 {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_client_id: req.ChannelIds})
	}
	if req.CtrlStatus > 0 {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_ctrl_status: req.CtrlStatus})
	}
	if len(req.GameName) > 0 {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_game_name: req.GameName})
	}
	if len(req.Currency) > 0 && req.Currency != "0" {
		sqb = sqb.Where(sq.Eq{GameDayStatis_F_currency: req.Currency})
	}
	if req.StartDate > 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_record_date: req.StartDate})
	}
	if req.EndDate > 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_record_date: req.EndDate})
	}

	// 筛选条件
	if req.TotalBetMin > 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_total_bet_amt: req.TotalBetMin})
	}
	if req.TotalBetMax > 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_total_bet_amt: req.TotalBetMax})
	}

	// 总兑现区间 (总兑现 = 游戏前兑现 + 游戏中兑现)
	if req.TotalCashoutMin != 0 || req.TotalCashoutMax != 0 {
		cashoutExpr := "(pre_cashout_amt + ing_cashout_amt)"
		if req.TotalCashoutMin != 0 {
			sqb = sqb.Where(sq.Expr(cashoutExpr+" >= ?", req.TotalCashoutMin))
		}
		if req.TotalCashoutMax != 0 {
			sqb = sqb.Where(sq.Expr(cashoutExpr+" <= ?", req.TotalCashoutMax))
		}
	}

	// 总盈亏区间
	if req.TotalProfitMin != 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_total_profit_amt: req.TotalProfitMin})
	}
	if req.TotalProfitMax != 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_total_profit_amt: req.TotalProfitMax})
	}

	// 总杀率区间
	if req.KillRateMin != 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_total_kill_rate: req.KillRateMin})
	}
	if req.KillRateMax != 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_total_kill_rate: req.KillRateMax})
	}

	// 百分比区间 (这里用返奖率)
	if req.PercentMin != 0 {
		sqb = sqb.Where(sq.GtOrEq{GameDayStatis_F_bonus_rate: req.PercentMin})
	}
	if req.PercentMax != 0 {
		sqb = sqb.Where(sq.LtOrEq{GameDayStatis_F_bonus_rate: req.PercentMax})
	}

	resp := 0
	sqlStr, sqlParams, _ := sqb.ToSql()
	err := m.conn.QueryRowCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}

	return resp, err
}

func (m *customGameDayStatisModel) GetPageNum(ctx context.Context, termId int64, channelIds []int64, ctrlStatus int64, startDate, endDate int, gameName string) (int, error) {
	sqb := sq.Select("count(*) as num").From(m.table)
	if termId > 0 {
		sqb = sqb.Where(sq.Eq{
			GameDayStatis_F_term_id: termId,
		})
	}
	if len(channelIds) > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_client_id: channelIds,
		})
	}
	if ctrlStatus > 0 {
		sqb = sqb.Where(sq.Eq{
			GameDayStatis_F_ctrl_status: ctrlStatus,
		})
	}
	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{
			GameDayStatis_F_game_name: gameName,
		})
	}
	if startDate > 0 {
		sqb = sqb.Where(sq.GtOrEq{
			GameDayStatis_F_record_date: startDate,
		})
	}
	if endDate > 0 {
		sqb = sqb.Where(sq.LtOrEq{
			GameDayStatis_F_record_date: endDate,
		})
	}
	resp := 0
	sqlStr, sqlParams, _ := sqb.ToSql()
	err := m.conn.QueryRowCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}

	return resp, err
}
