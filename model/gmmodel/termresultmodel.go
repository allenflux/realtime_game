package gmmodel

import (
	"context"
	"crash/model/servermodel"

	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type TermResultModel interface {
	GetTermResultList(ctx context.Context, clientId string, gameName string, termId int64, termHash string, startTime, endTime int64, page, pageSize int, multipleMin, multipleMax float64) ([]*TermResult, error)
	GetTermResultCount(ctx context.Context, clientId string, gameName string, termId int64, termHash string, startTime, endTime int64, multipleMin, multipleMax float64) (int, error)
	GetTermResultStat(
		ctx context.Context,
		clientId string,
		gameName string,
		termId int64,
		termHash string,
		startTime, endTime int64,
	) (*TermResultStat, error)
}

type defaultTermResultModel struct {
	conn  sqlx.SqlConn
	table string
}

type TermResult struct {
	TermId      int64   `db:"id"`
	ChannelId   int64   `db:"channel_id"`
	ClientId    string  `db:"client_id"`
	ClientName  string  `db:"client_name"`
	GameName    string  `db:"game_name"`
	Multiple    float64 `db:"multiple"`
	TermHash    string  `db:"term_hash"`
	TotalBetAmt int64   `db:"total_bet_amt"`
	CashedAmt   int64   `db:"cashed_amt"`
	ProfitAmt   int64   `db:"profit_amt"`
	CrashedTime int64   `db:"crashed_time"`
}

// NewTermResultModel 返回期数结果模型
func NewTermResultModel(conn sqlx.SqlConn) TermResultModel {
	return &defaultTermResultModel{
		conn:  conn,
		table: "crash_term",
	}
}

func (m *defaultTermResultModel) GetTermResultList(
	ctx context.Context,
	clientId string,
	gameName string,
	termId int64,
	termHash string,
	startTime, endTime int64,
	page, pageSize int,
	multipleMin, multipleMax float64,
) ([]*TermResult, error) {

	start := (page - 1) * pageSize

	sqb := sq.Select(
		"ct.term_id AS term_id",
		"ct.channel_id",
		"c.client_id",
		"c.client_name",
		"c.game_name",
		"ct.multiple",
		"ct.term_hash",
		"ct.total_bet_amt",
		"ct.cashed_amt",
		"(ct.total_bet_amt - ct.cashed_amt) AS profit_amt",
		"ct.crashed_time",
	).
		From(m.table + " ct").
		Join("channel c ON ct.channel_id = c.id").
		Where(sq.Eq{"ct.is_crashed": servermodel.CrashTermIsCrashedYes})

	// 先过滤 channel（如果有传）
	if clientId != "" {
		sqb = sqb.Where(sq.Eq{"c.client_id": clientId})
	}
	if gameName != "" {
		sqb = sqb.Where(sq.Eq{"c.game_name": gameName})
	}

	// 再过滤 crash_term
	if termId > 0 {
		sqb = sqb.Where(sq.Eq{"ct.term_id": termId})
	}
	if termHash != "" {
		sqb = sqb.Where(sq.Eq{"ct.term_hash": termHash})
	}
	if startTime > 0 && endTime > 0 {
		sqb = sqb.Where(sq.And{
			sq.GtOrEq{"ct.crashed_time": startTime},
			sq.LtOrEq{"ct.crashed_time": endTime},
		})
	}
	if multipleMin > 0 {
		sqb = sqb.Where(sq.GtOrEq{"ct.multiple": int64(multipleMin * 100)})
	}
	if multipleMax > 0 {
		sqb = sqb.Where(sq.LtOrEq{"ct.multiple": int64(multipleMax * 100)})
	}

	sqb = sqb.OrderBy("ct.crashed_time DESC").
		Limit(uint64(pageSize)).
		Offset(uint64(start))

	sqlStr, args, err := sqb.ToSql()
	if err != nil {
		return nil, err
	}

	var result []*TermResult
	err = m.conn.QueryRowsCtx(ctx, &result, sqlStr, args...)
	if err == sqlx.ErrNotFound {
		return nil, nil
	}
	return result, err
}

// GetTermResultList 获取期数结果列表
// 索引建议：
// - crash_term表：
//   - idx_crashed_time(crashed_time)
//   - idx_channel_id(channel_id)
//   - idx_term_hash(term_hash)
//
// - channel表：
//   - idx_client_game(client_id, game_name)
//func (m *defaultTermResultModel) GetTermResultList(ctx context.Context, clientId string, gameName string, termId int64, termHash string, startTime, endTime int64, page, pageSize int, multipleMin, multipleMax float64) ([]*TermResult, error) {
//	start := (page - 1) * pageSize
//
//	// 构建子查询以优先过滤crash_term表
//	subQuery := sq.Select("ct.id", "ct.channel_id", "ct.multiple", "ct.term_hash",
//		"ct.total_bet_amt", "ct.cashed_amt", "ct.crashed_time").
//		From(m.table + " ct").
//		Where(sq.Eq{"ct.is_crashed": servermodel.CrashTermIsCrashedYes})
//
//	if termId > 0 {
//		subQuery = subQuery.Where(sq.Eq{"ct.term_id": termId})
//	}
//	if termHash != "" {
//		subQuery = subQuery.Where(sq.Eq{"ct.term_hash": termHash})
//	}
//	if startTime != 0 && endTime != 0 {
//		subQuery = subQuery.Where(sq.GtOrEq{"ct.crashed_time": startTime}).
//			Where(sq.LtOrEq{"ct.crashed_time": endTime})
//	}
//	if multipleMin > 0 {
//		subQuery = subQuery.Where(sq.GtOrEq{"ct.multiple": multipleMin * 100})
//	}
//	if multipleMax > 0 {
//		subQuery = subQuery.Where(sq.LtOrEq{"ct.multiple": multipleMax * 100})
//	}
//
//	// 主查询使用子查询结果
//	sqb := sq.Select("t.id", "t.channel_id", "c.client_id", "c.client_name", "c.game_name",
//		"t.multiple", "t.term_hash", "t.total_bet_amt", "t.cashed_amt",
//		"(t.total_bet_amt - t.cashed_amt) as profit_amt", "t.crashed_time").
//		FromSelect(subQuery, "t").
//		Join("channel c FORCE INDEX(idx_client_game) ON t.channel_id = c.id")
//
//	// 添加channel表的筛选条件
//	if clientId != "" {
//		sqb = sqb.Where(sq.Eq{"c.client_id": clientId})
//	}
//	if gameName != "" {
//		sqb = sqb.Where(sq.Eq{"c.game_name": gameName})
//	}
//
//	// 默认按结束时间降序排列
//	sqb = sqb.OrderBy("t.crashed_time DESC").
//		Limit(uint64(pageSize)).
//		Offset(uint64(start))
//
//	sqlStr, args, err := sqb.ToSql()
//	if err != nil {
//		return nil, err
//	}
//
//	var result []*TermResult
//	err = m.conn.QueryRowsCtx(ctx, &result, sqlStr, args...)
//	if err == sqlx.ErrNotFound {
//		return nil, nil
//	}
//
//	return result, err
//}

// GetTermResultCount 获取期数结果总数
func (m *defaultTermResultModel) GetTermResultCount(ctx context.Context, clientId string, gameName string, termId int64, termHash string, startTime, endTime int64, multipleMin, multipleMax float64) (int, error) {
	// 构建子查询以优先过滤crash_term表
	subQuery := sq.Select("ct.channel_id").
		From(m.table + " ct").
		Where(sq.Eq{"ct.is_crashed": servermodel.CrashTermIsCrashedYes})

	if termId > 0 {
		subQuery = subQuery.Where(sq.Eq{"ct.id": termId})
	}
	if termHash != "" {
		subQuery = subQuery.Where(sq.Eq{"ct.term_hash": termHash})
	}
	if startTime != 0 && endTime != 0 {
		subQuery = subQuery.Where(sq.GtOrEq{"ct.crashed_time": startTime}).
			Where(sq.LtOrEq{"ct.crashed_time": endTime})
	}
	if multipleMin > 0 {
		subQuery = subQuery.Where(sq.GtOrEq{"ct.multiple": multipleMin * 100})
	}
	if multipleMax > 0 {
		subQuery = subQuery.Where(sq.LtOrEq{"ct.multiple": multipleMax * 100})
	}

	// 主查询使用子查询结果
	sqb := sq.Select("COUNT(*)").
		FromSelect(subQuery, "t").
		Join("channel c FORCE INDEX(idx_client_game) ON t.channel_id = c.id")

	if clientId != "" {
		sqb = sqb.Where(sq.Eq{"c.client_id": clientId})
	}
	if gameName != "" {
		sqb = sqb.Where(sq.Eq{"c.game_name": gameName})
	}

	sqlStr, args, err := sqb.ToSql()
	if err != nil {
		return 0, err
	}

	var count int
	err = m.conn.QueryRowCtx(ctx, &count, sqlStr, args...)
	if err == sqlx.ErrNotFound {
		return 0, nil
	}

	return count, err
}

type TermResultStat struct {
	TotalNum int `db:"total_num"`

	EqualOne int `db:"equal_one"`
	Lt1_5    int `db:"lt_1_5"`
	Lt2      int `db:"lt_2"`
	Lt5      int `db:"lt_5"`
	Lt10     int `db:"lt_10"`
	Lt20     int `db:"lt_20"`
	Lt50     int `db:"lt_50"`
	Lt100    int `db:"lt_100"`
	Ge100    int `db:"ge_100"`
}

func (m *defaultTermResultModel) GetTermResultStat(
	ctx context.Context,
	clientId string,
	gameName string,
	termId int64,
	termHash string,
	startTime, endTime int64,
) (*TermResultStat, error) {

	sqb := sq.Select(
		"COUNT(*) AS total_num",
		"COALESCE(SUM(ct.multiple = 100), 0) AS equal_one",
		"COALESCE(SUM(ct.multiple < 150), 0) AS lt_1_5",
		"COALESCE(SUM(ct.multiple < 200), 0) AS lt_2",
		"COALESCE(SUM(ct.multiple < 500), 0) AS lt_5",
		"COALESCE(SUM(ct.multiple < 1000), 0) AS lt_10",
		"COALESCE(SUM(ct.multiple < 2000), 0) AS lt_20",
		"COALESCE(SUM(ct.multiple < 5000), 0) AS lt_50",
		"COALESCE(SUM(ct.multiple < 10000), 0) AS lt_100",
		"COALESCE(SUM(ct.multiple >= 10000), 0) AS ge_100",
	).
		From(m.table + " ct").
		Join("channel c ON ct.channel_id = c.id").
		Where(sq.Eq{"ct.is_crashed": servermodel.CrashTermIsCrashedYes})

	if clientId != "" {
		sqb = sqb.Where(sq.Eq{"c.client_id": clientId})
	}
	if gameName != "" {
		sqb = sqb.Where(sq.Eq{"c.game_name": gameName})
	}

	if termId > 0 {
		sqb = sqb.Where(sq.Eq{"ct.term_id": termId}) // 重点：确认字段
	}
	if termHash != "" {
		sqb = sqb.Where(sq.Eq{"ct.term_hash": termHash})
	}
	if startTime > 0 && endTime > 0 {
		sqb = sqb.Where(sq.And{
			sq.GtOrEq{"ct.crashed_time": startTime},
			sq.LtOrEq{"ct.crashed_time": endTime},
		})
	}

	sqlStr, args, err := sqb.ToSql()
	if err != nil {
		return nil, err
	}

	var stat TermResultStat
	err = m.conn.QueryRowCtx(ctx, &stat, sqlStr, args...)
	if err == sqlx.ErrNotFound {
		return &stat, nil
	}
	return &stat, err
}

//func (m *defaultTermResultModel) GetTermResultStat(
//	ctx context.Context,
//	clientId string,
//	gameName string,
//	termId int64,
//	termHash string,
//	startTime, endTime int64,
//) (*TermResultStat, error) {
//
//	// ===== 子查询：与 GetTermResultList 完全一致 =====
//	subQuery := sq.Select(
//		"ct.id",
//		"ct.channel_id",
//		"ct.multiple",
//	).
//		From(m.table + " ct").
//		Where(sq.Eq{"ct.is_crashed": servermodel.CrashTermIsCrashedYes})
//
//	if termId > 0 {
//		subQuery = subQuery.Where(sq.Eq{"ct.id": termId})
//	}
//	if termHash != "" {
//		subQuery = subQuery.Where(sq.Eq{"ct.term_hash": termHash})
//	}
//	if startTime > 0 && endTime > 0 {
//		subQuery = subQuery.
//			Where(sq.GtOrEq{"ct.crashed_time": startTime}).
//			Where(sq.LtOrEq{"ct.crashed_time": endTime})
//	}
//
//	// ===== 主统计查询（关键：COALESCE）=====
//	sqb := sq.Select(
//		"COUNT(*) AS total_num",
//
//		"COALESCE(SUM(t.multiple = 100), 0) AS equal_one",
//		"COALESCE(SUM(t.multiple < 150), 0) AS lt_1_5",
//		"COALESCE(SUM(t.multiple < 200), 0) AS lt_2",
//		"COALESCE(SUM(t.multiple < 500), 0) AS lt_5",
//		"COALESCE(SUM(t.multiple < 1000), 0) AS lt_10",
//		"COALESCE(SUM(t.multiple < 2000), 0) AS lt_20",
//		"COALESCE(SUM(t.multiple < 5000), 0) AS lt_50",
//		"COALESCE(SUM(t.multiple < 10000), 0) AS lt_100",
//		"COALESCE(SUM(t.multiple >= 10000), 0) AS ge_100",
//	).
//		FromSelect(subQuery, "t").
//		Join("channel c FORCE INDEX(idx_client_game) ON t.channel_id = c.id")
//
//	if clientId != "" {
//		sqb = sqb.Where(sq.Eq{"c.client_id": clientId})
//	}
//	if gameName != "" {
//		sqb = sqb.Where(sq.Eq{"c.game_name": gameName})
//	}
//
//	sqlStr, args, err := sqb.ToSql()
//	if err != nil {
//		return nil, err
//	}
//
//	var stat TermResultStat
//	err = m.conn.QueryRowCtx(ctx, &stat, sqlStr, args...)
//	if err == sqlx.ErrNotFound {
//		// 极端情况兜底：直接返回 0 结构
//		return &stat, nil
//	}
//	return &stat, err
//}
