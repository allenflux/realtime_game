package gmmodel

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
)

var _ JackpotRecordModel = (*customJackpotRecordModel)(nil)

type (
	// JackpotRecordModel is an interface to be customized, add more methods here,
	// and implement the added methods in customJackpotRecordModel.
	JackpotRecordModel interface {
		jackpotRecordModel
		withSession(session sqlx.Session) JackpotRecordModel
		GetPage(ctx context.Context, args *JackpotRecordListArgs) ([]*JackpotRecord, int, error)
		FindLatestJackpotRecord(ctx context.Context, channelId, termId int64, gameName, currency string) (*JackpotRecord, error)
		GetStatsForTerm(ctx context.Context, channelId int64, gameName, currency string, termId int64) (*JackpotStats, error)
		FindTopByTermIdDesc(ctx context.Context, channelId int64, gameName, currency string, limit int) ([]*JackpotRecord, error)
		GetJackpotRecords(ctx context.Context, args *JackpotRecordListArgs) ([]*JackpotRecord, error)
		InsertBatch(ctx context.Context, dataList []*JackpotRecord) (sql.Result, error)
		FindLatest2PerCurrency(ctx context.Context, channelID int64) ([]*JackpotRecord, error)
	}

	customJackpotRecordModel struct {
		*defaultJackpotRecordModel
	}

	// 分页查询参数
	JackpotRecordListArgs struct {
		ChannelId int64
		ClientId  string
		GameName  string
		Currency  string
		TermId    int
		StartTime string
		EndTime   string
		Start     int
		PageSize  int
	}

	// 奖池统计数据结构体
	JackpotStats struct {
		JackpotInAmount   int64 `db:"jackpot_in_amount"`   // 计入奖池金额*10,000
		TotalJackpotPrize int64 `db:"total_jackpot_prize"` // 总派奖金额*10,000
	}
)

// NewJackpotRecordModel returns a model for the database table.
func NewJackpotRecordModel(conn sqlx.SqlConn) JackpotRecordModel {
	return &customJackpotRecordModel{
		defaultJackpotRecordModel: newJackpotRecordModel(conn),
	}
}

func (m *customJackpotRecordModel) withSession(session sqlx.Session) JackpotRecordModel {
	return NewJackpotRecordModel(sqlx.NewSqlConnFromSession(session))
}

// 分页查询方法
func (m *customJackpotRecordModel) GetPage(ctx context.Context, args *JackpotRecordListArgs) ([]*JackpotRecord, int, error) {
	return m.defaultJackpotRecordModel.getPage(ctx, args)
}

func (m *customJackpotRecordModel) GetJackpotRecords(ctx context.Context, args *JackpotRecordListArgs) ([]*JackpotRecord, error) {
	subQ := sq.
		Select("MAX(term_id)").
		From(m.table).
		Where(sq.Eq{
			"channel_id": args.ChannelId,
			"game_name":  args.GameName,
		})

	sqlStr, params, _ := sq.
		Select(jackpotRecordFieldNames...).
		From(m.table + " AS jr").
		Where(sq.Eq{
			"jr.channel_id": args.ChannelId,
			"jr.game_name":  args.GameName,
		}).
		Where(sq.Expr("jr.term_id = (?)", subQ)).
		ToSql()

	var rows []*JackpotRecord
	if err := m.conn.QueryRowsCtx(ctx, &rows, sqlStr, params...); err != nil {
		// QueryRowsCtx 通常不会返回 sqlx.ErrNotFound，直接返回 err 即可
		return nil, err
	}
	return rows, nil
}

func (m *defaultJackpotRecordModel) InsertBatch(ctx context.Context, dataList []*JackpotRecord) (sql.Result, error) {
	if len(dataList) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(dataList))
	args := make([]interface{}, 0, len(dataList)*13)

	for _, data := range dataList {
		placeholders = append(placeholders, "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
		args = append(args,
			data.ChannelId,
			data.ClientId,
			data.GameName,
			data.TermId,
			data.Ctime,
			data.EndTime,
			data.Currency,
			data.ValidBetAmount,
			data.JackpotInAmount,
			data.JackpotPrize1,
			data.JackpotPrize2,
			data.JackpotPrize3,
			data.JackpotBalance,
		)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES %s",
		m.table,
		jackpotRecordRowsExpectAutoSet,
		strings.Join(placeholders, ","),
	)

	return m.conn.ExecCtx(ctx, query, args...)
}

// 实现查找最新一条奖池记录的方法
func (m *customJackpotRecordModel) FindLatestJackpotRecord(ctx context.Context, channelId, termId int64, gameName, currency string) (*JackpotRecord, error) {
	var (
		where  []string
		params []interface{}
	)

	if channelId > 0 {
		where = append(where, "channel_id = ?")
		params = append(params, channelId)
	}
	if gameName != "" {
		where = append(where, "game_name = ?")
		params = append(params, gameName)
	}
	if currency != "" {
		where = append(where, "currency = ?")
		params = append(params, currency)
	}
	if termId > 0 {
		where = append(where, "term_id < ?")
		params = append(params, termId)
	}

	whereSql := ""
	if len(where) > 0 {
		whereSql = "WHERE " + strings.Join(where, " AND ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY end_time DESC LIMIT 1", jackpotRecordRows, m.table, whereSql)

	var resp JackpotRecord
	err := m.conn.QueryRowCtx(ctx, &resp, query, params...)

	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// 实现根据期数查询统计数据的方法
func (m *customJackpotRecordModel) GetStatsForTerm(ctx context.Context, channelId int64, gameName, currency string, termId int64) (*JackpotStats, error) {
	var (
		where  []string
		params []interface{}
	)

	if channelId > 0 {
		where = append(where, "channel_id = ?")
		params = append(params, channelId)
	}
	if gameName != "" {
		where = append(where, "game_name = ?")
		params = append(params, gameName)
	}
	if currency != "" {
		where = append(where, "currency = ?")
		params = append(params, currency)
	}
	if termId > 0 {
		where = append(where, "term_id = ?")
		params = append(params, termId)
	}

	whereSql := ""
	if len(where) > 0 {
		whereSql = "WHERE " + strings.Join(where, " AND ")
	}

	// 注意：这里假设 JackpotRecord 表中的 JackpotPrize1, JackpotPrize2, JackpotPrize3 存储的是正值派奖金额
	query := fmt.Sprintf("SELECT SUM(jackpot_in_amount) as jackpot_in_amount, SUM(jackpot_prize_1 + jackpot_prize_2 + jackpot_prize_3) as total_jackpot_prize FROM %s %s", m.table, whereSql)

	var stats JackpotStats
	err := m.conn.QueryRowCtx(ctx, &stats, query, params...)

	switch err {
	case nil:
		return &stats, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// 实现查找按 term_id 降序排序的前 N 条奖池记录的方法
func (m *customJackpotRecordModel) FindTopByTermIdDesc(ctx context.Context, channelId int64, gameName, currency string, limit int) ([]*JackpotRecord, error) {
	var (
		where  []string
		params []interface{}
	)

	if channelId > 0 {
		where = append(where, "channel_id = ?")
		params = append(params, channelId)
	}
	if gameName != "" {
		where = append(where, "game_name = ?")
		params = append(params, gameName)
	}
	if currency != "" {
		where = append(where, "currency = ?")
		params = append(params, currency)
	}

	whereSql := ""
	if len(where) > 0 {
		whereSql = "WHERE " + strings.Join(where, " AND ")
	}

	query := fmt.Sprintf("SELECT %s FROM %s %s ORDER BY term_id DESC LIMIT ?", jackpotRecordRows, m.table, whereSql)
	params = append(params, limit)

	var list []*JackpotRecord
	err := m.conn.QueryRowsCtx(ctx, &list, query, params...)

	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}

	return list, nil
}

func (m *customJackpotRecordModel) FindLatest2PerCurrency(ctx context.Context, channelID int64) ([]*JackpotRecord, error) {
	query := fmt.Sprintf(`
SELECT %s FROM (
  SELECT jr.*,
         ROW_NUMBER() OVER (
           PARTITION BY jr.currency
           ORDER BY jr.term_id DESC
         ) AS rn
  FROM %s jr
  WHERE jr.channel_id = ?
) x
WHERE x.rn <= 2
ORDER BY x.currency, x.term_id DESC
`, jackpotRecordRows, m.table)

	var resp []*JackpotRecord
	if err := m.conn.QueryRowsCtx(ctx, &resp, query, channelID); err != nil {
		return nil, err
	}
	return resp, nil
}
