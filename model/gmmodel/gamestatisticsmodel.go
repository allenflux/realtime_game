package gmmodel

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// 控制状态常量定义
const (
	CtrlStatus_Normal = 1 // 正常状态
	CtrlStatus_Fail   = 2 // 控输状态
	CtrlStatus_Win    = 3 // 控赢状态
)

// GameStatisticsFilter 游戏统计查询过滤条件
type GameStatisticsFilter struct {
	StartDate       *time.Time // 查询开始时间
	EndDate         *time.Time // 查询结束时间
	ClientId        string     // 代理ID
	GameName        string     // 趣投游戏
	Currency        string     // 币种
	StartBetAmt     int64      // 总投注最小值
	EndBetAmt       int64      // 总投注最大值
	TotalCashoutMin int64      // 总兑现最小值
	TotalCashoutMax int64      // 总兑现最大值
	StartProfitAmt  int64      // 总盈亏最小值
	EndProfitAmt    int64      // 总盈亏最大值
	StartRate       int64      // 总杀率最小值
	EndRate         int64      // 总杀率最大值
	SortField       string     // 排序字段
	SortOrder       string     // 排序方式 asc/desc
}

// GameStatistics 游戏统计数据
type GameStatistics struct {
	RecordDate         string `db:"record_date"`           // 日期
	ClientId           string `db:"client_id"`             // 代理ID
	ChannelName        string `db:"channel_name"`          // 代理名称
	GameName           string `db:"game_name"`             // 趣投游戏
	Currency           string `db:"currency"`              // 币种
	TotalBetTermNum    int64  `db:"total_bet_term_num"`    // 总期数
	PlayerNum          int64  `db:"player_num"`            // 总游戏人数
	TotalOrderNum      int64  `db:"total_order_num"`       // 总订单数
	TotalBetAmt        string `db:"total_bet_amt"`         // 总投注金额
	TotalValidBetAmt   string `db:"total_valid_bet_amt"`   // 总有效投注金额
	TotalCashoutAmt    string `db:"total_cashout_amt"`     // 总兑现金额
	JackpotInAmt       string `db:"jackpot_in_amt"`        // 奖池增加金额
	JackpotPayoutAmt   string `db:"jackpot_payout_amt"`    // 奖池派发金额
	TotalProfitAmt     string `db:"total_profit_amt"`      // 总盈亏
	TotalRate          string `db:"total_rate"`            // 总杀率%
	PreOrderNum        int64  `db:"pre_order_num"`         // 赛前订单数
	PreBetAmt          string `db:"pre_bet_amt"`           // 赛前投注金额
	PreValidBetAmt     string `db:"pre_valid_bet_amt"`     // 赛前有效投注金额
	PreCashoutAmt      string `db:"pre_cashout_amt"`       // 赛前兑现金额
	PreProfitAmt       string `db:"pre_profit_amt"`        // 赛前总盈亏
	PreKillRate        string `db:"pre_kill_rate"`         // 赛前投注杀率%
	RollingOrderNum    int64  `db:"rolling_order_num"`     // 滚盘订单数
	RollingBetAmt      string `db:"rolling_bet_amt"`       // 滚盘投注金额
	RollingValidBetAmt string `db:"rolling_valid_bet_amt"` // 滚盘有效投注金额
	TotalServiceFee    string `db:"total_service_fee"`     // 总服务费
	RollingCashoutAmt  string `db:"rolling_cashout_amt"`   // 滚盘兑现金额
	RollingProfitAmt   string `db:"rolling_profit_amt"`    // 滚盘总盈亏
	RollingKillRate    string `db:"rolling_kill_rate"`     // 滚盘投注杀率%
}

// GameStatisticsSummary 游戏统计汇总数据
type GameStatisticsSummary struct {
	TotalBetTermNum    int64  `db:"total_bet_term_num"`    // 总期数
	PlayerNum          int64  `db:"player_num"`            // 总游戏人数
	TotalOrderNum      int64  `db:"total_order_num"`       // 总订单数
	TotalBetAmt        string `db:"total_bet_amt"`         // 总投注金额
	TotalValidBetAmt   string `db:"total_valid_bet_amt"`   // 总有效投注金额
	TotalCashoutAmt    string `db:"total_cashout_amt"`     // 总兑现金额
	JackpotInAmt       string `db:"jackpot_in_amt"`        // 奖池增加金额
	JackpotPayoutAmt   string `db:"jackpot_payout_amt"`    // 奖池派发金额
	TotalProfitAmt     string `db:"total_profit_amt"`      // 总盈亏金额
	TotalRate          string `db:"total_rate"`            // 总杀率
	PreOrderNum        int64  `db:"pre_order_num"`         // 赛前订单数
	PreBetAmt          string `db:"pre_bet_amt"`           // 赛前投注金额
	PreValidBetAmt     string `db:"pre_valid_bet_amt"`     // 赛前有效投注金额
	PreCashoutAmt      string `db:"pre_cashout_amt"`       // 赛前兑现金额
	PreProfitAmt       string `db:"pre_profit_amt"`        // 赛前总盈亏
	PreKillRate        string `db:"pre_kill_rate"`         // 赛前投注杀率
	RollingOrderNum    int64  `db:"rolling_order_num"`     // 滚盘订单数
	RollingBetAmt      string `db:"rolling_bet_amt"`       // 滚盘投注金额
	RollingValidBetAmt string `db:"rolling_valid_bet_amt"` // 滚盘有效投注金额
	TotalServiceFee    string `db:"total_service_fee"`     // 总服务费
	RollingCashoutAmt  string `db:"rolling_cashout_amt"`   // 滚盘兑现金额
	RollingProfitAmt   string `db:"rolling_profit_amt"`    // 滚盘总盈亏
	RollingKillRate    string `db:"rolling_kill_rate"`     // 滚盘投注杀率
}

// TermStatisticsFilter 每期统计查询过滤条件
type TermStatisticsFilter struct {
	StartDate       string // 查询开始时间 YYYY-MM-DD HH:MM:SS
	EndDate         string // 查询结束时间 YYYY-MM-DD HH:MM:SS
	ClientId        string // 代理ID
	GameName        string // 趣投游戏
	Currency        string // 币种
	StartBetAmt     int64  // 总投注最小值
	EndBetAmt       int64  // 总投注最大值
	TotalCashoutMin int64  // 总兑现最小值
	TotalCashoutMax int64  // 总兑现最大值
	StartProfitAmt  int64  // 总盈亏最小值
	EndProfitAmt    int64  // 总盈亏最大值
	StartRate       int64  // 总杀率最小值
	EndRate         int64  // 总杀率最大值
	CtrlStatus      int64  // 期状态 1=正常 2=控输 3=控赢
	SortField       string // 排序字段
	SortOrder       string // 排序方式 asc/desc
	TermID          string // term id筛选
}

// TermStatistics 每期统计数据
type TermStatistics struct {
	RecordDate            string `db:"record_date"`              // 日期
	ClientId              string `db:"client_id"`                // 代理ID
	ChannelName           string `db:"channel_name"`             // 代理名称
	ChannelId             int64  `db:"channel_id"`               //
	GameName              string `db:"game_name"`                // 趣投游戏
	Currency              string `db:"currency"`                 // 币种
	TermId                int64  `db:"term_id"`                  // 期数ID
	TotalBetTermNum       int64  `db:"total_bet_term_num"`       // 总期数
	PlayerNum             int64  `db:"player_num"`               // 总游戏人数
	TotalOrderNum         int64  `db:"total_order_num"`          // 总订单数
	TotalBetAmt           string `db:"total_bet_amt"`            // 总投注金额
	TotalValidBetAmt      string `db:"total_valid_bet_amt"`      // 总有效投注金额
	TotalCashoutAmt       string `db:"total_cashout_amt"`        // 总兑现金额
	JackpotInAmt          string `db:"jackpot_in_amt"`           // 奖池增加金额
	JackpotPayoutAmt      string `db:"jackpot_payout_amt"`       // 奖池派发金额
	TotalProfitAmt        string `db:"total_profit_amt"`         // 总盈亏
	TotalRate             string `db:"total_rate"`               // 总杀率%
	PreOrderNum           int64  `db:"pre_order_num"`            // 赛前订单数
	PreBetAmt             string `db:"pre_bet_amt"`              // 赛前投注金额
	PreValidBetAmt        string `db:"pre_valid_bet_amt"`        // 赛前有效投注金额
	PreCashoutAmt         string `db:"pre_cashout_amt"`          // 赛前兑现金额
	PreCashoutUserNum     int64  `db:"pre_cashout_user_num"`     // 赛前兑现人次
	PreProfitAmt          string `db:"pre_profit_amt"`           // 赛前总盈亏
	PreKillRate           string `db:"pre_kill_rate"`            // 赛前投注杀率%
	RollingOrderNum       int64  `db:"rolling_order_num"`        // 滚盘订单数
	RollingBetAmt         string `db:"rolling_bet_amt"`          // 滚盘投注金额
	RollingValidBetAmt    string `db:"rolling_valid_bet_amt"`    // 滚盘有效投注金额
	TotalServiceFee       string `db:"total_service_fee"`        // 总服务费
	RollingCashoutAmt     string `db:"rolling_cashout_amt"`      // 滚盘兑现金额
	RollingCashoutUserNum int64  `db:"rolling_cashout_user_num"` // 滚盘兑现人次
	RollingProfitAmt      string `db:"rolling_profit_amt"`       // 滚盘总盈亏
	RollingKillRate       string `db:"rolling_kill_rate"`        // 滚盘投注杀率%
	CtrlStatus            int64  `db:"ctrl_status"`              // 期状态 1=正常 2=控输 3=控赢
}

type BetAggStat struct {
	TermId    int64  `db:"term_id"`
	ChannelId int64  `db:"channel_id"`
	Currency  string `db:"currency"`

	TotalOrderNum int64 `db:"total_order_num"`
	PlayerNum     int64 `db:"player_num"`

	TotalBetAmt      int64 `db:"total_bet_amt"`
	TotalValidBetAmt int64 `db:"total_valid_bet_amt"`
	TotalCashoutAmt  int64 `db:"total_cashout_amt"`
	TotalProfitAmt   int64 `db:"total_profit_amt"`

	PreOrderNum       int64 `db:"pre_order_num"`
	PreBetAmt         int64 `db:"pre_bet_amt"`
	PreValidBetAmt    int64 `db:"pre_valid_bet_amt"`
	PreCashoutAmt     int64 `db:"pre_cashout_amt"`
	PreCashoutUserNum int64 `db:"pre_cashout_user_num"`
	PreProfitAmt      int64 `db:"pre_profit_amt"`

	RollingOrderNum       int64 `db:"rolling_order_num"`
	RollingBetAmt         int64 `db:"rolling_bet_amt"`
	RollingValidBetAmt    int64 `db:"rolling_valid_bet_amt"`
	RollingCashoutAmt     int64 `db:"rolling_cashout_amt"`
	RollingCashoutUserNum int64 `db:"rolling_cashout_user_num"`
	RollingProfitAmt      int64 `db:"rolling_profit_amt"`

	TotalServiceFee int64  `db:"total_service_fee"`
	CreateTime      string `db:"create_time"`
}

// TermStatisticsSummary 每期统计汇总数据
type TermStatisticsSummary struct {
	TotalBetTermNum       int64  `db:"total_bet_term_num"`       // 总期数
	PlayerNum             int64  `db:"player_num"`               // 总游戏人数
	TotalOrderNum         int64  `db:"total_order_num"`          // 总订单数
	TotalBetAmt           string `db:"total_bet_amt"`            // 总投注金额
	TotalValidBetAmt      string `db:"total_valid_bet_amt"`      // 总有效投注金额
	TotalCashoutAmt       string `db:"total_cashout_amt"`        // 总兑现金额
	JackpotInAmt          string `db:"jackpot_in_amt"`           // 奖池增加金额
	JackpotPayoutAmt      string `db:"jackpot_payout_amt"`       // 奖池派发金额
	TotalProfitAmt        string `db:"total_profit_amt"`         // 总盈亏金额
	TotalRate             string `db:"total_rate"`               // 总杀率
	PreOrderNum           int64  `db:"pre_order_num"`            // 赛前订单数
	PreBetAmt             string `db:"pre_bet_amt"`              // 赛前投注金额
	PreValidBetAmt        string `db:"pre_valid_bet_amt"`        // 赛前有效投注金额
	PreCashoutAmt         string `db:"pre_cashout_amt"`          // 赛前兑现金额
	PreCashoutUserNum     int64  `db:"pre_cashout_user_num"`     // 赛前兑现人次
	PreProfitAmt          string `db:"pre_profit_amt"`           // 赛前总盈亏
	PreKillRate           string `db:"pre_kill_rate"`            // 赛前投注杀率
	RollingOrderNum       int64  `db:"rolling_order_num"`        // 滚盘订单数
	RollingBetAmt         string `db:"rolling_bet_amt"`          // 滚盘投注金额
	RollingValidBetAmt    string `db:"rolling_valid_bet_amt"`    // 滚盘有效投注金额
	TotalServiceFee       string `db:"total_service_fee"`        // 总服务费
	RollingCashoutAmt     string `db:"rolling_cashout_amt"`      // 滚盘兑现金额
	RollingCashoutUserNum int64  `db:"rolling_cashout_user_num"` // 滚盘兑现人次
	RollingProfitAmt      string `db:"rolling_profit_amt"`       // 滚盘总盈亏
	RollingKillRate       string `db:"rolling_kill_rate"`        // 滚盘投注杀率
	CtrlStatus            int64  `db:"ctrl_status"`              // 期状态 1=正常 2=控输 3=控赢
}

type (
	// GameStatisticsModel 游戏统计模型接口
	GameStatisticsModel interface {
		// GetGameStatistics 获取游戏统计数据
		GetGameStatistics(ctx context.Context, filter GameStatisticsFilter, page, pageSize int) ([]*GameStatistics, int64, error)

		// GetGameStatisticsSummary 获取游戏统计汇总数据
		GetGameStatisticsSummary(ctx context.Context, filter GameStatisticsFilter) (*GameStatisticsSummary, error)

		// GetTermStatistics 获取每期统计数据
		GetTermStatistics(ctx context.Context, filter TermStatisticsFilter, page, pageSize int) ([]*TermStatistics, int64, error)

		// GetTermStatisticsSummary 获取每期统计汇总数据
		GetTermStatisticsSummary(ctx context.Context, filter TermStatisticsFilter) (*TermStatisticsSummary, error)
	}

	// defaultGameStatisticsModel 默认游戏统计模型实现
	defaultGameStatisticsModel struct {
		conn sqlx.SqlConn
	}
)

// NewGameStatisticsModel 创建游戏统计模型
func NewGameStatisticsModel(conn sqlx.SqlConn) GameStatisticsModel {
	return &defaultGameStatisticsModel{
		conn: conn,
	}
}

// GetGameStatistics 获取游戏统计数据
//
//	func (m *defaultGameStatisticsModel) GetGameStatistics(ctx context.Context, filter GameStatisticsFilter, page, pageSize int) ([]*GameStatistics, int64, error) {
//		// 生成SQL及参数
//		selectSQL := `
//			SELECT
//				DATE_FORMAT(b.create_time, '%Y-%m-%d') AS record_date,
//				c.client_id,
//				c.client_name AS channel_name,
//				c.game_name,
//				b.currency,
//				COUNT(DISTINCT b.term_id) AS total_bet_term_num,
//				COUNT(DISTINCT b.user_id) AS player_num,
//				COUNT(DISTINCT b.id) AS total_order_num,
//				SUM(b.amount)/10000 AS total_bet_amt,
//				SUM(CASE WHEN b.order_status = 4000 THEN b.amount ELSE 0 END)/10000 AS total_valid_bet_amt,
//				SUM(b.cashed_out_amount)/10000 AS total_cashout_amt,
//				COALESCE(SUM(j.jackpot_in_amount), 0)/10000 as jackpot_in_amt,
//				COALESCE(SUM(j.jackpot_prize_1 + j.jackpot_prize_2 + j.jackpot_prize_3), 0)/10000 as jackpot_payout_amt,
//				SUM(b.amount-b.cashed_out_amount)/10000 AS total_profit_amt,
//				CASE WHEN SUM(CASE WHEN b.order_status = 4000 THEN b.amount ELSE 0 END) > 0 THEN
//					CAST(SUM(b.amount-b.cashed_out_amount) * 100 / SUM(CASE WHEN b.order_status = 4000 THEN b.amount ELSE 0 END) AS CHAR)
//				ELSE '0' END AS total_rate,
//				COUNT(CASE WHEN b.bet_type = 1 THEN b.id ELSE NULL END) AS pre_order_num,
//				SUM(CASE WHEN b.bet_type = 1 THEN b.amount ELSE 0 END)/10000 AS pre_bet_amt,
//				SUM(CASE WHEN b.bet_type = 1 AND b.order_status = 4000 THEN b.amount ELSE 0 END)/10000 AS pre_valid_bet_amt,
//				SUM(CASE WHEN b.bet_type = 1 THEN b.cashed_out_amount ELSE 0 END)/10000 AS pre_cashout_amt,
//				SUM(CASE WHEN b.bet_type = 1 THEN b.amount-b.cashed_out_amount ELSE 0 END)/10000 AS pre_profit_amt,
//				CASE WHEN SUM(CASE WHEN b.bet_type = 1 AND b.order_status = 4000 THEN b.amount ELSE 0 END) > 0 THEN
//					CAST(SUM(CASE WHEN b.bet_type = 1 THEN b.amount-b.cashed_out_amount ELSE 0 END) * 100 /
//					SUM(CASE WHEN b.bet_type = 1 AND b.order_status = 4000 THEN b.amount ELSE 0 END) AS CHAR)
//				ELSE '0' END AS pre_kill_rate,
//				COUNT(CASE WHEN b.bet_type = 2 THEN b.id ELSE NULL END) AS rolling_order_num,
//				SUM(CASE WHEN b.bet_type = 2 THEN b.amount ELSE 0 END)/10000 AS rolling_bet_amt,
//				SUM(CASE WHEN b.bet_type = 2 AND b.order_status = 4000 THEN b.amount ELSE 0 END)/10000 AS rolling_valid_bet_amt,
//				SUM(b.service_fee)/10000 AS total_service_fee,
//				SUM(CASE WHEN b.bet_type = 2 THEN b.cashed_out_amount ELSE 0 END)/10000 AS rolling_cashout_amt,
//				SUM(CASE WHEN b.bet_type = 2 THEN b.amount-b.cashed_out_amount ELSE 0 END)/10000 AS rolling_profit_amt,
//				CASE WHEN SUM(CASE WHEN b.bet_type = 2 AND b.order_status = 4000 THEN b.amount ELSE 0 END) > 0 THEN
//					CAST(SUM(CASE WHEN b.bet_type = 2 THEN b.amount-b.cashed_out_amount ELSE 0 END) * 100 /
//					SUM(CASE WHEN b.bet_type = 2 AND b.order_status = 4000 THEN b.amount ELSE 0 END) AS CHAR)
//				ELSE '0' END AS rolling_kill_rate
//				FROM
//				crashv2.bet b
//				JOIN crashv2.crash_term ct ON b.term_id = ct.term_id
//				JOIN crashv2.channel c ON b.channel_id = c.id
//				LEFT JOIN gc_managerv2.jackpot_record j ON ct.id = j.term_id AND ct.channel_id = j.channel_id AND b.currency = j.currency COLLATE utf8mb4_unicode_ci
//		`
//
//		whereSQL, args := m.buildWhereClause(filter)
//		if whereSQL != "" {
//			selectSQL += " WHERE " + whereSQL
//		}
//
//		selectSQL += " GROUP BY record_date, c.client_id, c.client_name, c.game_name, b.currency"
//
//		// 排序
//		if filter.SortField != "" && filter.SortOrder != "" {
//			selectSQL += fmt.Sprintf(" ORDER BY %s %s", filter.SortField, filter.SortOrder)
//		} else {
//			selectSQL += " ORDER BY record_date DESC, total_bet_amt DESC"
//		}
//
//		// 获取总记录数
//		countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS t", selectSQL)
//		var total int64
//		err := m.conn.QueryRowCtx(ctx, &total, countSQL, args...)
//		if err != nil && err != sql.ErrNoRows {
//			return nil, 0, err
//		}
//
//		if total == 0 {
//			return []*GameStatistics{}, 0, nil
//		}
//
//		// 添加分页限制
//		offset := (page - 1) * pageSize
//		selectSQL += fmt.Sprintf(" LIMIT %d, %d", offset, pageSize)
//
//		// 执行查询
//		var result []*GameStatistics
//		err = m.conn.QueryRowsCtx(ctx, &result, selectSQL, args...)
//		if err != nil && err != sql.ErrNoRows {
//			return nil, 0, err
//		}
//
//		return result, total, nil
//	}
func (m *defaultGameStatisticsModel) GetGameStatistics(
	ctx context.Context,
	filter GameStatisticsFilter,
	page, pageSize int,
) ([]*GameStatistics, int64, error) {

	// ---------- WHERE ----------
	whereSQL, args := m.buildWhereClause(filter)
	if whereSQL != "" {
		whereSQL = "WHERE " + whereSQL
	}

	// ---------- bet 聚合（统一 currency collation） ----------
	betAggSQL := fmt.Sprintf(`
		SELECT
			DATE(b.create_time) AS record_date,
			b.channel_id,
			b.currency COLLATE utf8mb4_general_ci AS currency,

			COUNT(DISTINCT b.term_id) AS total_bet_term_num,
			COUNT(DISTINCT b.user_id) AS player_num,
			COUNT(*) AS total_order_num,

			SUM(b.amount) AS total_bet_amt,
			SUM(IF(b.order_status = 4000, b.amount, 0)) AS total_valid_bet_amt,
			SUM(b.cashed_out_amount) AS total_cashout_amt,
			SUM(b.amount - b.cashed_out_amount) AS total_profit_amt,

			COUNT(IF(b.bet_type = 1, 1, NULL)) AS pre_order_num,
			SUM(IF(b.bet_type = 1, b.amount, 0)) AS pre_bet_amt,
			SUM(IF(b.bet_type = 1 AND b.order_status = 4000, b.amount, 0)) AS pre_valid_bet_amt,
			SUM(IF(b.bet_type = 1, b.cashed_out_amount, 0)) AS pre_cashout_amt,
			SUM(IF(b.bet_type = 1, b.amount - b.cashed_out_amount, 0)) AS pre_profit_amt,

			COUNT(IF(b.bet_type = 2, 1, NULL)) AS rolling_order_num,
			SUM(IF(b.bet_type = 2, b.amount, 0)) AS rolling_bet_amt,
			SUM(IF(b.bet_type = 2 AND b.order_status = 4000, b.amount, 0)) AS rolling_valid_bet_amt,
			SUM(IF(b.bet_type = 2, b.cashed_out_amount, 0)) AS rolling_cashout_amt,
			SUM(IF(b.bet_type = 2, b.amount - b.cashed_out_amount, 0)) AS rolling_profit_amt,

			SUM(b.service_fee) AS total_service_fee
		FROM crashv2.bet b
		%s
		GROUP BY DATE(b.create_time), b.channel_id, currency
	`, whereSQL)

	// ---------- jackpot 聚合 ----------
	jackpotAggSQL := `
		SELECT
			ct.channel_id,
			j.currency COLLATE utf8mb4_general_ci AS currency,
			SUM(j.jackpot_in_amount) AS jackpot_in_amt,
			SUM(j.jackpot_prize_1 + j.jackpot_prize_2 + j.jackpot_prize_3) AS jackpot_payout_amt
		FROM gc_managerv2.jackpot_record j
		JOIN crashv2.crash_term ct ON ct.id = j.term_id
		GROUP BY ct.channel_id, currency
	`

	// ---------- 主查询（严格对齐 GameStatistics struct） ----------
	selectSQL := fmt.Sprintf(`
		SELECT
			t.record_date                                           AS record_date,
			CAST(c.client_id AS CHAR)                               AS client_id,
			c.client_name                                           AS channel_name,
			c.game_name                                             AS game_name,
			t.currency                                              AS currency,

			t.total_bet_term_num                                    AS total_bet_term_num,
			t.player_num                                            AS player_num,
			t.total_order_num                                       AS total_order_num,

			CAST(t.total_bet_amt / 10000 AS CHAR)                   AS total_bet_amt,
			CAST(t.total_valid_bet_amt / 10000 AS CHAR)             AS total_valid_bet_amt,
			CAST(t.total_cashout_amt / 10000 AS CHAR)               AS total_cashout_amt,
			CAST(IFNULL(j.jackpot_in_amt, 0) / 10000 AS CHAR)        AS jackpot_in_amt,
			CAST(IFNULL(j.jackpot_payout_amt, 0) / 10000 AS CHAR)    AS jackpot_payout_amt,
			CAST(t.total_profit_amt / 10000 AS CHAR)                AS total_profit_amt,

			CAST(
				IF(t.total_valid_bet_amt > 0,
				   t.total_profit_amt * 100 / t.total_valid_bet_amt,
				   0
				) AS CHAR
			)                                                       AS total_rate,

			t.pre_order_num                                         AS pre_order_num,
			CAST(t.pre_bet_amt / 10000 AS CHAR)                     AS pre_bet_amt,
			CAST(t.pre_valid_bet_amt / 10000 AS CHAR)               AS pre_valid_bet_amt,
			CAST(t.pre_cashout_amt / 10000 AS CHAR)                 AS pre_cashout_amt,
			CAST(t.pre_profit_amt / 10000 AS CHAR)                  AS pre_profit_amt,
			CAST(
				IF(t.pre_valid_bet_amt > 0,
				   t.pre_profit_amt * 100 / t.pre_valid_bet_amt,
				   0
				) AS CHAR
			)                                                       AS pre_kill_rate,

			t.rolling_order_num                                     AS rolling_order_num,
			CAST(t.rolling_bet_amt / 10000 AS CHAR)                 AS rolling_bet_amt,
			CAST(t.rolling_valid_bet_amt / 10000 AS CHAR)           AS rolling_valid_bet_amt,
			CAST(t.total_service_fee / 10000 AS CHAR)               AS total_service_fee,
			CAST(t.rolling_cashout_amt / 10000 AS CHAR)             AS rolling_cashout_amt,
			CAST(t.rolling_profit_amt / 10000 AS CHAR)              AS rolling_profit_amt,
			CAST(
				IF(t.rolling_valid_bet_amt > 0,
				   t.rolling_profit_amt * 100 / t.rolling_valid_bet_amt,
				   0
				) AS CHAR
			)                                                       AS rolling_kill_rate
		FROM (%s) t
		JOIN crashv2.channel c ON t.channel_id = c.id
		LEFT JOIN (%s) j
			ON j.channel_id = t.channel_id
		   AND j.currency = t.currency
		ORDER BY t.record_date DESC
	`, betAggSQL, jackpotAggSQL)

	// ---------- COUNT（只做维度统计，避免全量聚合） ----------
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*) FROM (
			SELECT 1
			FROM crashv2.bet b
			%s
			GROUP BY DATE(b.create_time), b.channel_id, b.currency
		) x
	`, whereSQL)

	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countSQL, args...); err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []*GameStatistics{}, 0, nil
	}

	// ---------- 分页 ----------
	offset := (page - 1) * pageSize
	selectSQL += " LIMIT ?, ?"
	args = append(args, offset, pageSize)

	var result []*GameStatistics
	if err := m.conn.QueryRowsCtx(ctx, &result, selectSQL, args...); err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

// GetGameStatisticsSummary 获取游戏统计汇总数据
func (m *defaultGameStatisticsModel) GetGameStatisticsSummary(
	ctx context.Context,
	filter GameStatisticsFilter,
) (*GameStatisticsSummary, error) {

	// ===============================
	// 1. bet 聚合子查询
	// ===============================
	betSQL := `
		SELECT
			b.term_id,
			b.channel_id,
			b.currency COLLATE utf8mb4_unicode_ci AS currency,

			COUNT(*) AS total_order_num,
			COUNT(DISTINCT b.user_id) AS player_num,
			COUNT(DISTINCT b.term_id) AS total_bet_term_num,

			SUM(b.amount) AS total_bet_amt,
			SUM(CASE WHEN b.order_status = 4000 THEN b.amount ELSE 0 END) AS total_valid_bet_amt,
			SUM(b.cashed_out_amount) AS total_cashout_amt,
			SUM(b.amount - b.cashed_out_amount) AS total_profit_amt,

			SUM(CASE WHEN b.bet_type = 1 THEN 1 ELSE 0 END) AS pre_order_num,
			SUM(CASE WHEN b.bet_type = 1 THEN b.amount ELSE 0 END) AS pre_bet_amt,
			SUM(CASE WHEN b.bet_type = 1 AND b.order_status = 4000 THEN b.amount ELSE 0 END) AS pre_valid_bet_amt,
			SUM(CASE WHEN b.bet_type = 1 THEN b.cashed_out_amount ELSE 0 END) AS pre_cashout_amt,
			SUM(CASE WHEN b.bet_type = 1 THEN b.amount - b.cashed_out_amount ELSE 0 END) AS pre_profit_amt,

			SUM(CASE WHEN b.bet_type = 2 THEN 1 ELSE 0 END) AS rolling_order_num,
			SUM(CASE WHEN b.bet_type = 2 THEN b.amount ELSE 0 END) AS rolling_bet_amt,
			SUM(CASE WHEN b.bet_type = 2 AND b.order_status = 4000 THEN b.amount ELSE 0 END) AS rolling_valid_bet_amt,
			SUM(CASE WHEN b.bet_type = 2 THEN b.cashed_out_amount ELSE 0 END) AS rolling_cashout_amt,
			SUM(CASE WHEN b.bet_type = 2 THEN b.amount - b.cashed_out_amount ELSE 0 END) AS rolling_profit_amt,

			SUM(b.service_fee) AS total_service_fee
		FROM crashv2.bet b
	`

	whereSQL, args := m.buildWhereClause(filter)
	if whereSQL != "" {
		betSQL += " WHERE " + whereSQL
	}

	betSQL += " GROUP BY b.term_id, b.channel_id, currency "

	// ===============================
	// 2. jackpot 聚合子查询
	// ===============================
	jackpotSQL := `
		SELECT
			j.term_id,
			j.channel_id,
			j.currency COLLATE utf8mb4_unicode_ci AS currency,
			SUM(j.jackpot_in_amount) AS jackpot_in_amt,
			SUM(j.jackpot_prize_1 + j.jackpot_prize_2 + j.jackpot_prize_3) AS jackpot_payout_amt
		FROM gc_managerv2.jackpot_record j
		GROUP BY j.term_id, j.channel_id, currency
	`

	// ===============================
	// 3. 最终汇总
	// ===============================
	finalSQL := `
		SELECT
			SUM(a.total_bet_term_num) AS total_bet_term_num,
			SUM(a.player_num) AS player_num,
			SUM(a.total_order_num) AS total_order_num,

			SUM(a.total_bet_amt)/10000 AS total_bet_amt,
			SUM(a.total_valid_bet_amt)/10000 AS total_valid_bet_amt,
			SUM(a.total_cashout_amt)/10000 AS total_cashout_amt,

			COALESCE(SUM(j.jackpot_in_amt),0)/10000 AS jackpot_in_amt,
			COALESCE(SUM(j.jackpot_payout_amt),0)/10000 AS jackpot_payout_amt,

			SUM(a.total_profit_amt)/10000 AS total_profit_amt,

			CASE
				WHEN SUM(a.total_valid_bet_amt) > 0
				THEN CAST(SUM(a.total_profit_amt) * 100 / SUM(a.total_valid_bet_amt) AS CHAR)
				ELSE '0'
			END AS total_rate,

			SUM(a.pre_order_num) AS pre_order_num,
			SUM(a.pre_bet_amt)/10000 AS pre_bet_amt,
			SUM(a.pre_valid_bet_amt)/10000 AS pre_valid_bet_amt,
			SUM(a.pre_cashout_amt)/10000 AS pre_cashout_amt,
			SUM(a.pre_profit_amt)/10000 AS pre_profit_amt,

			CASE
				WHEN SUM(a.pre_valid_bet_amt) > 0
				THEN CAST(SUM(a.pre_profit_amt) * 100 / SUM(a.pre_valid_bet_amt) AS CHAR)
				ELSE '0'
			END AS pre_kill_rate,

			SUM(a.rolling_order_num) AS rolling_order_num,
			SUM(a.rolling_bet_amt)/10000 AS rolling_bet_amt,
			SUM(a.rolling_valid_bet_amt)/10000 AS rolling_valid_bet_amt,
			SUM(a.rolling_cashout_amt)/10000 AS rolling_cashout_amt,
			SUM(a.rolling_profit_amt)/10000 AS rolling_profit_amt,

			CASE
				WHEN SUM(a.rolling_valid_bet_amt) > 0
				THEN CAST(SUM(a.rolling_profit_amt) * 100 / SUM(a.rolling_valid_bet_amt) AS CHAR)
				ELSE '0'
			END AS rolling_kill_rate,

			SUM(a.total_service_fee)/10000 AS total_service_fee
		FROM (` + betSQL + `) a
		JOIN crashv2.channel c
		  ON a.channel_id = c.id
		LEFT JOIN (` + jackpotSQL + `) j
		  ON a.term_id = j.term_id
		 AND a.channel_id = j.channel_id
		 AND a.currency = j.currency
	`

	// 外层字符串条件统一 COLLATE
	var outerWhere []string
	if filter.ClientId != "" {
		outerWhere = append(outerWhere, "c.client_id COLLATE utf8mb4_unicode_ci = ?")
		args = append(args, filter.ClientId)
	}
	if filter.GameName != "" {
		outerWhere = append(outerWhere, "c.game_name COLLATE utf8mb4_unicode_ci = ?")
		args = append(args, filter.GameName)
	}

	if len(outerWhere) > 0 {
		finalSQL += " WHERE " + strings.Join(outerWhere, " AND ")
	}

	var result GameStatisticsSummary
	err := m.conn.QueryRowCtx(ctx, &result, finalSQL, args...)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return &result, nil
}

// buildWhereClause 构建WHERE子句和参数
func (m *defaultGameStatisticsModel) buildWhereClause(filter GameStatisticsFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	// 时间范围过滤
	if filter.StartDate != nil {
		conditions = append(conditions, "b.create_time >= ?")
		args = append(args, filter.StartDate.Format("2006-01-02 15:04:05"))
	}

	if filter.EndDate != nil {
		conditions = append(conditions, "b.create_time <= ?")
		args = append(args, filter.EndDate.Format("2006-01-02 15:04:05"))
	}

	// 代理ID过滤
	if filter.ClientId != "" {
		conditions = append(conditions, "c.client_id = ?")
		args = append(args, filter.ClientId)
	}

	// 游戏名称过滤
	if filter.GameName != "" {
		conditions = append(conditions, "c.game_name = ?")
		args = append(args, filter.GameName)
	}

	// 币种过滤
	if filter.Currency != "" && filter.Currency != "0" {
		conditions = append(conditions, "b.currency COLLATE utf8mb4_unicode_ci = ?")
		args = append(args, filter.Currency)
	}

	// 总投注金额范围过滤 - 这些条件需要在HAVING子句中处理，但为简化实现，我们在WHERE子句中处理
	if filter.StartBetAmt > 0 {
		conditions = append(conditions, "b.amount >= ?")
		args = append(args, filter.StartBetAmt)
	}
	if filter.EndBetAmt > 0 {
		conditions = append(conditions, "b.amount <= ?")
		args = append(args, filter.EndBetAmt)
	}

	// 总兑现金额范围过滤
	if filter.TotalCashoutMin > 0 {
		conditions = append(conditions, "b.cashed_out_amount >= ?")
		args = append(args, filter.TotalCashoutMin)
	}
	if filter.TotalCashoutMax > 0 {
		conditions = append(conditions, "b.cashed_out_amount <= ?")
		args = append(args, filter.TotalCashoutMax)
	}

	// 总盈亏金额范围过滤
	if filter.StartProfitAmt != 0 {
		conditions = append(conditions, "(b.amount-b.cashed_out_amount) >= ?")
		args = append(args, filter.StartProfitAmt)
	}
	if filter.EndProfitAmt != 0 {
		conditions = append(conditions, "(b.amount-b.cashed_out_amount) <= ?")
		args = append(args, filter.EndProfitAmt)
	}

	// 总杀率范围过滤 - 这些条件需要在HAVING子句中处理，但为简化实现，我们省略此处理

	if len(conditions) == 0 {
		return "", args
	}

	return strings.Join(conditions, " AND "), args
}

// GetTermStatistics 获取每期统计数据
func (m *defaultGameStatisticsModel) GetTermStatistics(
	ctx context.Context,
	filter TermStatisticsFilter,
	page, pageSize int,
) ([]*TermStatistics, int64, error) {

	// ===============================
	// Step 0：参数 & where
	// ===============================
	whereSQL, whereArgs := m.buildTermStatisticsWhere(filter)

	// ===============================
	// Step 1：COUNT（只数 term）
	// ===============================
	logx.Info("Step 1：COUNT")
	countSQL := `
		SELECT COUNT(*)
		FROM (
			SELECT DISTINCT b.term_id
    		FROM crashv2.bet b FORCE INDEX (term_id)
	`

	if whereSQL != "" {
		countSQL += " WHERE " + whereSQL
	}
	countSQL += "  LIMIT 10001) t"

	var total int64
	if err := m.conn.QueryRowCtx(ctx, &total, countSQL, whereArgs...); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []*TermStatistics{}, 0, nil
	}

	// ===============================
	// Step 2：分页拿 term_id（极快）
	// ===============================
	logx.Info("Step 2：分页拿 term_id")

	// ===============================
	// 从 bet 表快速拿到所有 term_id（去重）
	// ===============================

	type termIDRow struct {
		TermID int64 `db:"term_id"`
	}

	// whereArgs 里应包含：startTime, endTime
	// 例如：2026-01-08 00:00:00 / 2026-01-09 00:00:00
	termIDSQL := `
	SELECT DISTINCT b.term_id
	FROM crashv2.bet b FORCE INDEX (term_id)
	WHERE b.create_time >= ?
	  AND b.create_time < ?
`
	if filter.TermID != "" {
		termIDSQL += " AND b.term_id = ?"
	}

	var termIDRows []termIDRow
	if err := m.conn.QueryRowsCtx(ctx, &termIDRows, termIDSQL, whereArgs...); err != nil {
		return nil, 0, err
	}
	if len(termIDRows) == 0 {
		return []*TermStatistics{}, total, nil
	}

	allTermIds := make([]int64, 0, len(termIDRows))
	for _, r := range termIDRows {
		allTermIds = append(allTermIds, r.TermID)
	}

	// ===============================
	// 回查 crash_term 拿 create_time（小表，极快）
	// ===============================

	type termRow struct {
		TermID     int64     `db:"term_id"`
		CreateTime time.Time `db:"create_time"`
	}

	placeholders := make([]string, len(allTermIds))
	args := make([]interface{}, len(allTermIds))
	for i, id := range allTermIds {
		placeholders[i] = "?"
		args[i] = id
	}

	termInfoSQL := fmt.Sprintf(`
	SELECT term_id, create_time
	FROM crashv2.crash_term
	WHERE term_id IN (%s)
`, strings.Join(placeholders, ","))

	var termRows []termRow
	if err := m.conn.QueryRowsCtx(ctx, &termRows, termInfoSQL, args...); err != nil {
		return nil, 0, err
	}

	// ===============================
	// Go 层排序（按 create_time 倒序）
	// ===============================

	sort.Slice(termRows, func(i, j int) bool {
		return termRows[i].CreateTime.After(termRows[j].CreateTime)
	})

	// ===============================
	// Go 层分页（彻底摆脱 SQL OFFSET）
	// ===============================

	start := (page - 1) * pageSize
	if start >= len(termRows) {
		return []*TermStatistics{}, total, nil
	}

	end := start + pageSize
	if end > len(termRows) {
		end = len(termRows)
	}

	pagedTerms := termRows[start:end]

	// ===============================
	// 得到最终需要的 termIds
	// ===============================

	termIds := make([]int64, 0, len(pagedTerms))
	for _, t := range pagedTerms {
		termIds = append(termIds, t.TermID)
	}

	// 👉 termIds 就是你后续所有 bet / jackpot 聚合 SQL 的输入

	if len(termIds) == 0 {
		return []*TermStatistics{}, total, nil
	}

	// 构造 IN (?, ?, ?)
	placeholders2 := make([]string, 0, len(termIds))
	inArgs := make([]interface{}, 0, len(termIds))
	for _, id := range termIds {
		placeholders2 = append(placeholders2, "?")
		inArgs = append(inArgs, id)
	}
	inClause := strings.Join(placeholders2, ",")

	// ===============================
	// Step 3：bet 聚合（核心但可控）
	// ===============================
	logx.Info("Step 3：bet 聚合")
	betAggSQL := `
		SELECT
			b.term_id,
			b.channel_id,
			b.currency,
			b.create_time,

			COUNT(*) AS total_order_num,
			COUNT(DISTINCT b.user_id) AS player_num,

			SUM(b.amount) AS total_bet_amt,
			SUM(CASE WHEN b.order_status = 4000 THEN b.amount ELSE 0 END) AS total_valid_bet_amt,
			SUM(b.cashed_out_amount) AS total_cashout_amt,
			SUM(b.amount - b.cashed_out_amount) AS total_profit_amt,

			SUM(CASE WHEN b.bet_type = 1 THEN 1 ELSE 0 END) AS pre_order_num,
			SUM(CASE WHEN b.bet_type = 1 THEN b.amount ELSE 0 END) AS pre_bet_amt,
			SUM(CASE WHEN b.bet_type = 1 AND b.order_status = 4000 THEN b.amount ELSE 0 END) AS pre_valid_bet_amt,
			SUM(CASE WHEN b.bet_type = 1 THEN b.cashed_out_amount ELSE 0 END) AS pre_cashout_amt,
			COUNT(DISTINCT CASE WHEN b.bet_type = 1 AND b.cashed_out_amount > 0 THEN b.user_id END) AS pre_cashout_user_num,
			SUM(CASE WHEN b.bet_type = 1 THEN b.amount - b.cashed_out_amount ELSE 0 END) AS pre_profit_amt,

			SUM(CASE WHEN b.bet_type = 2 THEN 1 ELSE 0 END) AS rolling_order_num,
			SUM(CASE WHEN b.bet_type = 2 THEN b.amount ELSE 0 END) AS rolling_bet_amt,
			SUM(CASE WHEN b.bet_type = 2 AND b.order_status = 4000 THEN b.amount ELSE 0 END) AS rolling_valid_bet_amt,
			SUM(CASE WHEN b.bet_type = 2 THEN b.cashed_out_amount ELSE 0 END) AS rolling_cashout_amt,
			COUNT(DISTINCT CASE WHEN b.bet_type = 2 AND b.cashed_out_amount > 0 THEN b.user_id END) AS rolling_cashout_user_num,
			SUM(CASE WHEN b.bet_type = 2 THEN b.amount - b.cashed_out_amount ELSE 0 END) AS rolling_profit_amt,

			SUM(b.service_fee) AS total_service_fee
		FROM crashv2.bet b
		WHERE b.term_id IN (` + inClause + `)
		GROUP BY b.term_id, b.channel_id, b.currency
	`

	var betAggStats []*BetAggStat
	if err := m.conn.QueryRowsCtx(ctx, &betAggStats, betAggSQL, inArgs...); err != nil {
		return nil, 0, err
	}

	betStats := make([]*TermStatistics, 0, len(betAggStats))

	for _, a := range betAggStats {
		ts := &TermStatistics{
			RecordDate: a.CreateTime,

			// === 核心维度 ===
			TermId:    a.TermId,
			ChannelId: a.ChannelId,
			Currency:  a.Currency,

			// === 计数类 ===
			PlayerNum:     a.PlayerNum,
			TotalOrderNum: a.TotalOrderNum,

			PreOrderNum:     a.PreOrderNum,
			RollingOrderNum: a.RollingOrderNum,

			PreCashoutUserNum:     a.PreCashoutUserNum,
			RollingCashoutUserNum: a.RollingCashoutUserNum,

			// === 金额类（统一 /10000 转 string）===
			TotalBetAmt:      strconv.FormatInt(a.TotalBetAmt/10000, 10),
			TotalValidBetAmt: strconv.FormatInt(a.TotalValidBetAmt/10000, 10),
			TotalCashoutAmt:  strconv.FormatInt(a.TotalCashoutAmt/10000, 10),
			TotalProfitAmt:   strconv.FormatInt(a.TotalProfitAmt/10000, 10),

			PreBetAmt:      strconv.FormatInt(a.PreBetAmt/10000, 10),
			PreValidBetAmt: strconv.FormatInt(a.PreValidBetAmt/10000, 10),
			PreCashoutAmt:  strconv.FormatInt(a.PreCashoutAmt/10000, 10),
			PreProfitAmt:   strconv.FormatInt(a.PreProfitAmt/10000, 10),

			RollingBetAmt:      strconv.FormatInt(a.RollingBetAmt/10000, 10),
			RollingValidBetAmt: strconv.FormatInt(a.RollingValidBetAmt/10000, 10),
			RollingCashoutAmt:  strconv.FormatInt(a.RollingCashoutAmt/10000, 10),
			RollingProfitAmt:   strconv.FormatInt(a.RollingProfitAmt/10000, 10),

			TotalServiceFee: strconv.FormatInt(a.TotalServiceFee/10000, 10),
		}

		// === 杀率（Go 里算，避免 SQL 再做表达式）===

		if a.TotalValidBetAmt > 0 {
			ts.TotalRate = strconv.FormatInt(
				a.TotalProfitAmt*100/a.TotalValidBetAmt,
				10,
			)
		} else {
			ts.TotalRate = "0"
		}

		if a.PreValidBetAmt > 0 {
			ts.PreKillRate = strconv.FormatInt(
				a.PreProfitAmt*100/a.PreValidBetAmt,
				10,
			)
		} else {
			ts.PreKillRate = "0"
		}

		if a.RollingValidBetAmt > 0 {
			ts.RollingKillRate = strconv.FormatInt(
				a.RollingProfitAmt*100/a.RollingValidBetAmt,
				10,
			)
		} else {
			ts.RollingKillRate = "0"
		}

		// === 期数统计（你之前是常量 1）===
		ts.TotalBetTermNum = 1

		betStats = append(betStats, ts)
	}

	// ===============================
	// Step 4：jackpot 聚合（独立，快）
	// ===============================
	logx.Info("Step 4：jackpot")
	jackpotSQL := `
		SELECT
			j.term_id,
			j.channel_id,
			j.currency,
			SUM(j.jackpot_in_amount) AS jackpot_in_amt,
			SUM(j.jackpot_prize_1 + j.jackpot_prize_2 + j.jackpot_prize_3) AS jackpot_payout_amt
		FROM gc_managerv2.jackpot_record j
		WHERE j.term_id IN (` + inClause + `)
		GROUP BY j.term_id, j.channel_id, j.currency
	`

	type jackpotAgg struct {
		TermID           int64  `db:"term_id"`
		ChannelID        int64  `db:"channel_id"`
		Currency         string `db:"currency"`
		JackpotInAmt     int64  `db:"jackpot_in_amt"`
		JackpotPayoutAmt int64  `db:"jackpot_payout_amt"`
	}

	var jackpots []jackpotAgg
	if err := m.conn.QueryRowsCtx(ctx, &jackpots, jackpotSQL, inArgs...); err != nil {
		return nil, 0, err
	}

	// ===============================
	// Step 5：Go Map 合并（极快）
	// ===============================
	logx.Info("Step 5：Go Map")
	type key struct {
		TermID    int64
		ChannelID int64
		Currency  string
	}

	jackpotMap := make(map[key]jackpotAgg, len(jackpots))
	for _, j := range jackpots {
		jackpotMap[key{j.TermID, j.ChannelID, j.Currency}] = j
	}

	for _, s := range betStats {
		if j, ok := jackpotMap[key{s.TermId, s.ChannelId, s.Currency}]; ok {
			s.JackpotInAmt = fmt.Sprintf("%d", j.JackpotInAmt)
			s.JackpotPayoutAmt = fmt.Sprintf("%d", j.JackpotPayoutAmt)
		}
	}

	return betStats, total, nil
}

// GetTermStatisticsSummary 获取每期统计汇总数据
func (m *defaultGameStatisticsModel) GetTermStatisticsSummary(
	ctx context.Context,
	filter TermStatisticsFilter,
) (*TermStatisticsSummary, error) {

	// ===============================
	// bet 聚合
	// ===============================
	betAggSQL := `
		SELECT
			b.term_id,
			b.channel_id,
			b.currency COLLATE utf8mb4_unicode_ci AS currency,

			COUNT(*) AS total_order_num,
			COUNT(DISTINCT b.user_id) AS player_num,

			SUM(b.amount) AS total_bet_amt,
			SUM(CASE WHEN b.order_status = 4000 THEN b.amount ELSE 0 END) AS total_valid_bet_amt,
			SUM(b.cashed_out_amount) AS total_cashout_amt,
			SUM(b.amount - b.cashed_out_amount) AS total_profit_amt,

			SUM(CASE WHEN b.bet_type = 1 THEN 1 ELSE 0 END) AS pre_order_num,
			SUM(CASE WHEN b.bet_type = 1 THEN b.amount ELSE 0 END) AS pre_bet_amt,
			SUM(CASE WHEN b.bet_type = 1 AND b.order_status = 4000 THEN b.amount ELSE 0 END) AS pre_valid_bet_amt,
			SUM(CASE WHEN b.bet_type = 1 THEN b.cashed_out_amount ELSE 0 END) AS pre_cashout_amt,
			COUNT(DISTINCT CASE WHEN b.bet_type = 1 AND b.cashed_out_amount > 0 THEN b.user_id END) AS pre_cashout_user_num,
			SUM(CASE WHEN b.bet_type = 1 THEN b.amount - b.cashed_out_amount ELSE 0 END) AS pre_profit_amt,

			SUM(CASE WHEN b.bet_type = 2 THEN 1 ELSE 0 END) AS rolling_order_num,
			SUM(CASE WHEN b.bet_type = 2 THEN b.amount ELSE 0 END) AS rolling_bet_amt,
			SUM(CASE WHEN b.bet_type = 2 AND b.order_status = 4000 THEN b.amount ELSE 0 END) AS rolling_valid_bet_amt,
			SUM(CASE WHEN b.bet_type = 2 THEN b.cashed_out_amount ELSE 0 END) AS rolling_cashout_amt,
			COUNT(DISTINCT CASE WHEN b.bet_type = 2 AND b.cashed_out_amount > 0 THEN b.user_id END) AS rolling_cashout_user_num,
			SUM(CASE WHEN b.bet_type = 2 THEN b.amount - b.cashed_out_amount ELSE 0 END) AS rolling_profit_amt,

			SUM(b.service_fee) AS total_service_fee
		FROM crashv2.bet b
	`

	whereSQL, args := m.buildTermStatisticsWhere(filter)
	if whereSQL != "" {
		betAggSQL += " WHERE " + whereSQL
	}

	betAggSQL += " GROUP BY b.term_id, b.channel_id, currency "

	// ===============================
	// jackpot 聚合
	// ===============================
	jackpotAggSQL := `
		SELECT
			term_id,
			channel_id,
			currency COLLATE utf8mb4_unicode_ci AS currency,
			SUM(jackpot_in_amount) AS jackpot_in_amt,
			SUM(jackpot_prize_1 + jackpot_prize_2 + jackpot_prize_3) AS jackpot_payout_amt
		FROM gc_managerv2.jackpot_record
		GROUP BY term_id, channel_id, currency
	`

	// ===============================
	// 汇总
	// ===============================
	finalSQL := `
		SELECT
			COUNT(*) AS total_bet_term_num,
			SUM(a.player_num) AS player_num,
			SUM(a.total_order_num) AS total_order_num,

			SUM(a.total_bet_amt)/10000 AS total_bet_amt,
			SUM(a.total_valid_bet_amt)/10000 AS total_valid_bet_amt,
			SUM(a.total_cashout_amt)/10000 AS total_cashout_amt,
			COALESCE(SUM(j.jackpot_in_amt),0)/10000 AS jackpot_in_amt,
			COALESCE(SUM(j.jackpot_payout_amt),0)/10000 AS jackpot_payout_amt,
			SUM(a.total_profit_amt)/10000 AS total_profit_amt,

			CASE
				WHEN SUM(a.total_valid_bet_amt) > 0
				THEN CAST(SUM(a.total_profit_amt) * 100 / SUM(a.total_valid_bet_amt) AS CHAR)
				ELSE '0'
			END AS total_rate,

			SUM(a.pre_order_num) AS pre_order_num,
			SUM(a.pre_bet_amt)/10000 AS pre_bet_amt,
			SUM(a.pre_valid_bet_amt)/10000 AS pre_valid_bet_amt,
			SUM(a.pre_cashout_amt)/10000 AS pre_cashout_amt,
			SUM(a.pre_cashout_user_num) AS pre_cashout_user_num,
			SUM(a.pre_profit_amt)/10000 AS pre_profit_amt,

			CASE
				WHEN SUM(a.pre_valid_bet_amt) > 0
				THEN CAST(SUM(a.pre_profit_amt) * 100 / SUM(a.pre_valid_bet_amt) AS CHAR)
				ELSE '0'
			END AS pre_kill_rate,

			SUM(a.rolling_order_num) AS rolling_order_num,
			SUM(a.rolling_bet_amt)/10000 AS rolling_bet_amt,
			SUM(a.rolling_valid_bet_amt)/10000 AS rolling_valid_bet_amt,
			SUM(a.rolling_cashout_amt)/10000 AS rolling_cashout_amt,
			SUM(a.rolling_cashout_user_num) AS rolling_cashout_user_num,
			SUM(a.rolling_profit_amt)/10000 AS rolling_profit_amt,

			CASE
				WHEN SUM(a.rolling_valid_bet_amt) > 0
				THEN CAST(SUM(a.rolling_profit_amt) * 100 / SUM(a.rolling_valid_bet_amt) AS CHAR)
				ELSE '0'
			END AS rolling_kill_rate,

			SUM(a.total_service_fee)/10000 AS total_service_fee
		FROM (` + betAggSQL + `) a
		JOIN crashv2.channel c ON a.channel_id = c.id
		LEFT JOIN (` + jackpotAggSQL + `) j
		  ON a.term_id = j.term_id
		 AND a.channel_id = j.channel_id
		 AND a.currency = j.currency
	`

	if filter.ClientId != "" {
		finalSQL += " WHERE c.client_id COLLATE utf8mb4_unicode_ci = ?"
		args = append(args, filter.ClientId)
	}

	var result TermStatisticsSummary
	if err := m.conn.QueryRowCtx(ctx, &result, finalSQL, args...); err != nil {
		return nil, err
	}

	return &result, nil
}

func (m *defaultGameStatisticsModel) buildTermStatisticsHaving(
	filter TermStatisticsFilter,
) (string, []interface{}) {

	var (
		conds []string
		args  []interface{}
	)

	if filter.StartBetAmt > 0 {
		conds = append(conds, "SUM(a.total_bet_amt) >= ?")
		args = append(args, filter.StartBetAmt)
	}
	if filter.EndBetAmt > 0 {
		conds = append(conds, "SUM(a.total_bet_amt) <= ?")
		args = append(args, filter.EndBetAmt)
	}

	if filter.TotalCashoutMin > 0 {
		conds = append(conds, "SUM(a.total_cashout_amt) >= ?")
		args = append(args, filter.TotalCashoutMin)
	}
	if filter.TotalCashoutMax > 0 {
		conds = append(conds, "SUM(a.total_cashout_amt) <= ?")
		args = append(args, filter.TotalCashoutMax)
	}

	if filter.StartProfitAmt != 0 {
		conds = append(conds, "SUM(a.total_profit_amt) >= ?")
		args = append(args, filter.StartProfitAmt)
	}
	if filter.EndProfitAmt != 0 {
		conds = append(conds, "SUM(a.total_profit_amt) <= ?")
		args = append(args, filter.EndProfitAmt)
	}

	// 杀率（注意除 0）
	if filter.StartRate > 0 {
		conds = append(conds,
			"(SUM(a.total_profit_amt) * 100 / NULLIF(SUM(a.total_valid_bet_amt),0)) >= ?",
		)
		args = append(args, filter.StartRate)
	}
	if filter.EndRate > 0 {
		conds = append(conds,
			"(SUM(a.total_profit_amt) * 100 / NULLIF(SUM(a.total_valid_bet_amt),0)) <= ?",
		)
		args = append(args, filter.EndRate)
	}

	if len(conds) == 0 {
		return "", args
	}

	return " HAVING " + strings.Join(conds, " AND "), args
}
func (m *defaultGameStatisticsModel) buildTermStatisticsWhere(
	filter TermStatisticsFilter,
) (string, []interface{}) {

	var (
		conds []string
		args  []interface{}
	)

	// ===== 时间范围（走索引，最重要）=====
	if filter.StartDate != "" {
		conds = append(conds, "b.create_time >= ?")
		//args = append(args, filter.StartDate+" 00:00:00")
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != "" {
		conds = append(conds, "b.create_time <= ?")
		//args = append(args, filter.EndDate+" 23:59:59")
		args = append(args, filter.EndDate)
	}

	// ===== 币种（显式 COLLATE，避免 MySQL 8 报错）=====
	if filter.Currency != "" && filter.Currency != "0" {
		conds = append(conds, "b.currency COLLATE utf8mb4_unicode_ci = ?")
		args = append(args, filter.Currency)
	}

	if filter.TermID != "" {
		conds = append(conds, "b.term_id = ?")
		args = append(args, filter.TermID)
	}

	// ⚠️ 注意：
	// client_id / game_name 不能在这里放
	// 因为 bet 表里没有这些字段
	// 必须在外层 JOIN channel 后再过滤

	return strings.Join(conds, " AND "), args
}
