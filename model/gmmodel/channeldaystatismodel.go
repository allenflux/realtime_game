package gmmodel

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	ChannelDayStatis_F_id              = "id"
	ChannelDayStatis_F_client_id       = "client_id"
	ChannelDayStatis_F_game_name       = "game_name"
	ChannelDayStatis_F_player_num      = "player_num"
	ChannelDayStatis_F_play_num        = "play_num"
	ChannelDayStatis_F_pre_bet_amt     = "pre_bet_amt"
	ChannelDayStatis_F_pre_cashout_amt = "pre_cashout_amt"
	ChannelDayStatis_F_ing_bet_amt     = "ing_bet_amt"
	ChannelDayStatis_F_ing_cashout_amt = "ing_cashout_amt"
	ChannelDayStatis_F_fee_amt         = "fee_amt"
	ChannelDayStatis_F_total_bet_amt   = "total_bet_amt"
	ChannelDayStatis_F_total_bonus     = "total_bonus"
	ChannelDayStatis_F_bonus_rate      = "bonus_rate"
	ChannelDayStatis_F_one_num         = "one_num"
	ChannelDayStatis_F_two_num         = "two_num"
	ChannelDayStatis_F_five_num        = "five_num"
	ChannelDayStatis_F_ten_num         = "ten_num"
	ChannelDayStatis_F_ten_more_num    = "ten_more_num"
	ChannelDayStatis_F_record_date     = "record_date"
	ChannelDayStatis_F_create_time     = "create_time"
)

// ChannelDayStatisSummary 汇总统计数据结构
type ChannelDayStatisSummary struct {
	TotalDays          int     `db:"total_days"`            // 统计天数
	TotalChannels      int     `db:"total_channels"`        // 总渠道数
	TotalPlayerNum     int64   `db:"total_player_num"`      // 总游戏人次
	TotalPlayNum       int64   `db:"total_play_num"`        // 总局数
	TotalPreBetAmt     float64 `db:"total_pre_bet_amt"`     // 总游戏前下注金额
	TotalPreCashoutAmt float64 `db:"total_pre_cashout_amt"` // 总游戏前兑现金额
	TotalIngBetAmt     float64 `db:"total_ing_bet_amt"`     // 总游戏中下注金额
	TotalIngCashoutAmt float64 `db:"total_ing_cashout_amt"` // 总游戏中兑现金额
	TotalFeeAmt        float64 `db:"total_fee_amt"`         // 总服务费
	TotalBetAmt        float64 `db:"total_bet_amt"`         // 总下注金额
	TotalBonus         float64 `db:"total_bonus"`           // 总返奖金额
	AvgBonusRate       float64 `db:"avg_bonus_rate"`        // 平均返奖率
	TotalOneNum        int64   `db:"total_one_num"`         // 总倍数=1的局数
	TotalTwoNum        int64   `db:"total_two_num"`         // 总倍数(1,2)的局数
	TotalFiveNum       int64   `db:"total_five_num"`        // 总倍数[2,5)的局数
	TotalTenNum        int64   `db:"total_ten_num"`         // 总倍数[5,10)的局数
	TotalTenMoreNum    int64   `db:"total_ten_more_num"`    // 总倍数[10,MAX]的局数
}

var _ ChannelDayStatisModel = (*customChannelDayStatisModel)(nil)

type (
	// ChannelDayStatisModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChannelDayStatisModel.
	ChannelDayStatisModel interface {
		channelDayStatisModel
		withSession(session sqlx.Session) ChannelDayStatisModel
		GetPage(ctx context.Context, channelIds []int64, startDate, endDate int, gameName, bonusRateSort string, pageSize int, start int) ([]ChannelDayStatis, error)
		GetPageNum(ctx context.Context, channelIds []int64, startDate, endDate int, gameName string) (int, error)
		FindOneByRecordDateClientIdCurrency(ctx context.Context, recordDate int64, clientId int64, currency string) (*ChannelDayStatis, error)
		GetSummaryData(ctx context.Context, channelIds []int64, startDate, endDate int, gameName, currency string) (*ChannelDayStatisSummary, error)
		GetDataWithTimeZoneOffset(ctx context.Context, channelIds []int64, startDate, endDate int, gameName, currency string, agentTimeZone int64) ([]ChannelDayStatis, error)
		GetPageWithFilters(ctx context.Context, channelIds []int64, startDate, endDate int, gameName, bonusRateSort string, pageSize, offset int, filters interface{}) ([]ChannelDayStatis, error)
		GetPageNumWithFilters(ctx context.Context, channelIds []int64, startDate, endDate int, gameName string, filters interface{}) (int, error)
	}

	customChannelDayStatisModel struct {
		*defaultChannelDayStatisModel
		rds *redis.Redis
	}
)

// NewChannelDayStatisModel returns a model for the database table.
func NewChannelDayStatisModel(conn sqlx.SqlConn, rds *redis.Redis) ChannelDayStatisModel {
	return &customChannelDayStatisModel{
		defaultChannelDayStatisModel: newChannelDayStatisModel(conn, rds),
		rds:                          rds,
	}
}

func (m *customChannelDayStatisModel) withSession(session sqlx.Session) ChannelDayStatisModel {
	return NewChannelDayStatisModel(sqlx.NewSqlConnFromSession(session), m.rds)
}

// 获取一页数据
func (m *customChannelDayStatisModel) GetPage(ctx context.Context, channelIds []int64, startDate, endDate int, gameName, bonusRateSort string, pageSize int, start int) ([]ChannelDayStatis, error) {
	sqb := sq.Select("*")
	if len(channelIds) > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_client_id: channelIds,
		})
	}
	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_game_name: gameName,
		})
	}
	if startDate == endDate && startDate > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_record_date: startDate,
		})
	} else {
		if startDate > 0 {
			sqb = sqb.Where(sq.GtOrEq{
				ChannelDayStatis_F_record_date: startDate,
			})
		}
		if endDate > 0 {
			sqb = sqb.Where(sq.LtOrEq{
				ChannelDayStatis_F_record_date: endDate,
			})
		}
	}
	if bonusRateSort == "asc" {
		sqb = sqb.OrderBy(ChannelDayStatis_F_bonus_rate + " asc")
	} else if bonusRateSort == "desc" {
		sqb = sqb.OrderBy(ChannelDayStatis_F_bonus_rate + " desc")
	}
	sqlStr, sqlParams, _ := sqb.From(m.table).Offset((uint64(start))).Limit(uint64(pageSize)).ToSql()
	resp := []ChannelDayStatis{}
	err := m.QueryRowsNoCacheCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

// 获取总条数
func (m *customChannelDayStatisModel) GetPageNum(ctx context.Context, channelIds []int64, startDate, endDate int, gameName string) (int, error) {
	sqb := sq.Select("count(*) as num")
	if len(channelIds) > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_client_id: channelIds,
		})
	}
	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_game_name: gameName,
		})
	}
	if startDate == endDate && startDate > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_record_date: startDate,
		})
	} else {
		if startDate > 0 {
			sqb = sqb.Where(sq.GtOrEq{
				ChannelDayStatis_F_record_date: startDate,
			})
		}
		if endDate > 0 {
			sqb = sqb.Where(sq.LtOrEq{
				ChannelDayStatis_F_record_date: endDate,
			})
		}
	}

	sqlStr, sqlParams, _ := sqb.From(m.table).ToSql()
	resp := 0
	err := m.QueryRowNoCacheCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}

	return resp, nil
}

// FindOneByRecordDateClientIdCurrency 根据记录日期、渠道ID和币种查询
func (m *customChannelDayStatisModel) FindOneByRecordDateClientIdCurrency(ctx context.Context, recordDate int64, clientId int64, currency string) (*ChannelDayStatis, error) {
	// 使用已生成的方法
	return m.defaultChannelDayStatisModel.FindOneByRecordDateClientIdCurrency(ctx, recordDate, clientId, currency)
}

// GetSummaryData 获取汇总统计数据
func (m *customChannelDayStatisModel) GetSummaryData(ctx context.Context, channelIds []int64, startDate, endDate int, gameName, currency string) (*ChannelDayStatisSummary, error) {
	sqb := sq.Select(`
		COUNT(DISTINCT record_date) as total_days,
		COUNT(DISTINCT client_id) as total_channels,
		SUM(player_num) as total_player_num,
		SUM(play_num) as total_play_num,
		SUM(pre_bet_amt) as total_pre_bet_amt,
		SUM(pre_cashout_amt) as total_pre_cashout_amt,
		SUM(ing_bet_amt) as total_ing_bet_amt,
		SUM(ing_cashout_amt) as total_ing_cashout_amt,
		SUM(fee_amt) as total_fee_amt,
		SUM(total_bet_amt) as total_bet_amt,
		SUM(total_bonus) as total_bonus,
		SUM(one_num) as total_one_num,
		SUM(two_num) as total_two_num,
		SUM(five_num) as total_five_num,
		SUM(ten_num) as total_ten_num,
		SUM(ten_more_num) as total_ten_more_num
	`).From(m.table)

	if len(channelIds) > 0 {
		sqb = sqb.Where(sq.Eq{"client_id": channelIds})
	}
	if startDate > 0 {
		sqb = sqb.Where(sq.GtOrEq{"record_date": startDate})
	}
	if endDate > 0 {
		sqb = sqb.Where(sq.LtOrEq{"record_date": endDate})
	}
	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{"game_name": gameName})
	}
	if len(currency) > 0 && currency != "0" {
		sqb = sqb.Where(sq.Eq{"currency": currency})
	}

	query, args, err := sqb.ToSql()
	if err != nil {
		return nil, err
	}

	var summary ChannelDayStatisSummary
	err = m.QueryRowNoCacheCtx(ctx, &summary, query, args...)
	if err != nil {
		return nil, err
	}

	return &summary, nil
}

// GetDataWithTimeZoneOffset 根据代理商时区偏移获取数据
func (m *customChannelDayStatisModel) GetDataWithTimeZoneOffset(ctx context.Context, channelIds []int64, startDate, endDate int, gameName, currency string, agentTimeZone int64) ([]ChannelDayStatis, error) {
	// 根据时区偏移调整查询日期范围
	adjustedStartDate := startDate
	adjustedEndDate := endDate

	// 如果有时区偏移，需要调整查询的日期范围
	if agentTimeZone != 0 {
		// 向前偏移需要包含更早的数据，向后偏移需要包含更晚的数据
		if agentTimeZone > 0 {
			// 正向偏移，数据时间往后推，查询范围需要往前推
			adjustedStartDate = startDate - 1
		} else {
			// 负向偏移，数据时间往前推，查询范围需要往后推
			adjustedEndDate = endDate + 1
		}
	}

	sqb := sq.Select("*")
	if len(channelIds) > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_client_id: channelIds,
		})
	}
	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_game_name: gameName,
		})
	}
	if len(currency) > 0 && currency != "0" {
		sqb = sqb.Where(sq.Eq{
			"currency": currency,
		})
	}
	if adjustedStartDate == adjustedEndDate && adjustedStartDate > 0 {
		sqb = sqb.Where(sq.Eq{
			ChannelDayStatis_F_record_date: adjustedStartDate,
		})
	} else {
		if adjustedStartDate > 0 {
			sqb = sqb.Where(sq.GtOrEq{
				ChannelDayStatis_F_record_date: adjustedStartDate,
			})
		}
		if adjustedEndDate > 0 {
			sqb = sqb.Where(sq.LtOrEq{
				ChannelDayStatis_F_record_date: adjustedEndDate,
			})
		}
	}

	sqlStr, sqlParams, _ := sqb.From(m.table).OrderBy(ChannelDayStatis_F_record_date + " desc").ToSql()
	resp := []ChannelDayStatis{}
	err := m.QueryRowsNoCacheCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

// GetPageWithFilters 获取分页数据（支持新的筛选条件）
func (m *customChannelDayStatisModel) GetPageWithFilters(ctx context.Context, channelIds []int64, startDate, endDate int, gameName, bonusRateSort string, pageSize, offset int, filters interface{}) ([]ChannelDayStatis, error) {
	sqb := sq.Select(channelDayStatisRows).From(m.table)

	if len(channelIds) > 0 {
		sqb = sqb.Where(sq.Eq{"client_id": channelIds})
	}
	if startDate > 0 {
		sqb = sqb.Where(sq.GtOrEq{"record_date": startDate})
	}
	if endDate > 0 {
		sqb = sqb.Where(sq.LtOrEq{"record_date": endDate})
	}
	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{"game_name": gameName})
	}

	// 添加新的筛选条件 - 这里需要根据具体的filters类型进行断言和处理
	// 暂时简化处理，后续可以根据实际需求优化

	// 如果有返奖率排序，添加排序条件
	if len(bonusRateSort) > 0 {
		if bonusRateSort == "asc" {
			sqb = sqb.OrderBy("bonus_rate ASC")
		} else if bonusRateSort == "desc" {
			sqb = sqb.OrderBy("bonus_rate DESC")
		}
	} else {
		// 默认按记录日期倒序排列
		sqb = sqb.OrderBy("record_date DESC, id DESC")
	}

	if pageSize > 0 {
		sqb = sqb.Limit(uint64(pageSize))
	}
	if offset > 0 {
		sqb = sqb.Offset(uint64(offset))
	}

	query, args, err := sqb.ToSql()
	if err != nil {
		return nil, err
	}

	var rs []ChannelDayStatis
	err = m.QueryRowsNoCacheCtx(ctx, &rs, query, args...)
	if err != nil {
		return nil, err
	}

	return rs, nil
}

// GetPageNumWithFilters 获取符合条件的总数（支持新的筛选条件）
func (m *customChannelDayStatisModel) GetPageNumWithFilters(ctx context.Context, channelIds []int64, startDate, endDate int, gameName string, filters interface{}) (int, error) {
	sqb := sq.Select("COUNT(*)").From(m.table)

	if len(channelIds) > 0 {
		sqb = sqb.Where(sq.Eq{"client_id": channelIds})
	}
	if startDate > 0 {
		sqb = sqb.Where(sq.GtOrEq{"record_date": startDate})
	}
	if endDate > 0 {
		sqb = sqb.Where(sq.LtOrEq{"record_date": endDate})
	}
	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{"game_name": gameName})
	}

	// 添加新的筛选条件 - 这里需要根据具体的filters类型进行断言和处理
	// 暂时简化处理，后续可以根据实际需求优化

	query, args, err := sqb.ToSql()
	if err != nil {
		return 0, err
	}

	var count int
	err = m.QueryRowNoCacheCtx(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}
