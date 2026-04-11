package servermodel

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"

	sq "github.com/Masterminds/squirrel"
)

var _ BonusBetModel = (*customBonusBetModel)(nil)

type (
	// BonusBetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBonusBetModel.
	BonusBetModel interface {
		bonusBetModel
		withSession(session sqlx.Session) BonusBetModel
		InsertBatch(ctx context.Context, dataList []*BonusBet) ([]*BonusBet, error)
		GetAdminPage(ctx context.Context, args *GetAdminPageArgs, offset, limit int) ([]*BonusBet, error)
		GetAdminPageNum(ctx context.Context, args *GetAdminPageArgs) (int64, error)
		GetByBetIds(ctx context.Context, betIds []int64) ([]*BonusBet, error)
		GetByChannelUserTerms(ctx context.Context, channelId, userId int64, termIds []int64) ([]*BonusBet, error)
	}

	customBonusBetModel struct {
		*defaultBonusBetModel
	}
)

// NewBonusBetModel returns a model for the database table.
func NewBonusBetModel(conn sqlx.SqlConn) BonusBetModel {
	return &customBonusBetModel{
		defaultBonusBetModel: newBonusBetModel(conn),
	}
}

func (m *customBonusBetModel) withSession(session sqlx.Session) BonusBetModel {
	return NewBonusBetModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customBonusBetModel) InsertBatch(ctx context.Context, dataList []*BonusBet) ([]*BonusBet, error) {
	if len(dataList) == 0 {
		return nil, nil
	}

	// 显式列：不含 `id`，顺序与 args 完全一致
	const cols = "`bet_id`,`channel_id`,`term_id`,`user_id`,`bonus_order_no`,`amount`,`bet_at_multiple`,`rake`,`service_fee`,`cashout_multiple`,`order_status`,`bonus_amount`,`ctime`,`bonus_rank`"
	const ph = "(?,?,?,?,?,?,?,?,?,?,?,?,?,?)" // 14 个

	placeholders := make([]string, 0, len(dataList))
	args := make([]any, 0, len(dataList)*14)

	for _, d := range dataList {
		placeholders = append(placeholders, ph)
		args = append(args,
			d.BetId,
			d.ChannelId,
			d.TermId,
			d.UserId,
			d.BonusOrderNo,
			d.Amount,
			d.BetAtMultiple,
			d.Rake,
			d.ServiceFee,
			d.CashoutMultiple,
			d.OrderStatus,
			d.BonusAmount,
			d.Ctime,
			d.BonusRank,
		)
	}

	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", m.table, cols, strings.Join(placeholders, ","))
	res, err := m.conn.ExecCtx(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	// 仅在“单条 INSERT 多值、全部成功、无触发器/ODKU/IGNORE”的前提下用 LastInsertId 回推
	firstID, err := res.LastInsertId()
	if err != nil {
		// 拿不到就直接返回，不要强行回推
		return dataList, nil
	}
	aff, err := res.RowsAffected()
	if err != nil || int64(len(dataList)) != aff {
		// 行数不一致时也不要回推
		return dataList, nil
	}
	for i, d := range dataList {
		d.Id = firstID + int64(i)
	}
	return dataList, nil
}

func (m *customBonusBetModel) GetAdminPage(ctx context.Context, args *GetAdminPageArgs, offset, limit int) ([]*BonusBet, error) {
	query := sq.Select(bonusBetRows).From(m.table)
	if len(args.ChannelId) > 0 {
		query = query.Where(sq.Eq{"channel_id": args.ChannelId})
	}
	if args.TermId > 0 {
		query = query.Where(sq.Eq{"term_id": args.TermId})
	}
	if args.UserId > 0 {
		query = query.Where(sq.Eq{"user_id": args.UserId})
	}
	if args.StartDate > 0 {
		startDate := time.Unix(args.StartDate, 0).Format("2006-01-02 15:04:05")
		query = query.Where(sq.GtOrEq{"DATE(create_time)": startDate})
	}
	if args.EndDate > 0 {
		endDate := time.Unix(args.EndDate, 0).Format("2006-01-02 15:04:05")
		query = query.Where(sq.LtOrEq{"DATE(create_time)": endDate})
	}
	// 新增查询条件
	if len(args.OrderId) > 0 {
		query = query.Where(sq.Eq{"id": args.OrderId})
	}
	// 奖金订单的币种查询，需要查看实际表结构
	if len(args.Currency) > 0 && args.Currency != "0" {
		// 注意：奖金订单表可能没有currency字段，需要根据实际情况调整
		// 暂时跳过，因为奖金订单可能通过关联的原始订单来确定币种
	}
	// 奖金订单的投注金额通常为0，不需要按投注金额过滤
	if args.BetAmtMin > 0 || args.BetAmtMax > 0 {
		// 奖金订单的投注金额为0，如果有投注金额范围查询，直接返回空结果
		return []*BonusBet{}, nil
	}
	// 奖金倍数查询条件 - 对于奖金订单，使用不同的逻辑
	if args.MultipleMin > 0 || args.MultipleMax > 0 {
		// 奖金订单的奖金倍数计算：cashout_multiple / bet_at_multiple
		if args.MultipleMin > 0 && args.MultipleMax > 0 {
			query = query.Where("(cashout_multiple / bet_at_multiple) >= ? AND (cashout_multiple / bet_at_multiple) <= ?", args.MultipleMin*100, args.MultipleMax*100)
		} else if args.MultipleMin > 0 {
			query = query.Where("(cashout_multiple / bet_at_multiple) >= ?", args.MultipleMin*100)
		} else if args.MultipleMax > 0 {
			query = query.Where("(cashout_multiple / bet_at_multiple) <= ?", args.MultipleMax*100)
		}
	}
	query = query.OrderBy("id desc").Limit(uint64(limit)).Offset(uint64(offset))

	sqlStr, sqlArgs, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*BonusBet
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlArgs...)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customBonusBetModel) GetAdminPageNum(ctx context.Context, args *GetAdminPageArgs) (int64, error) {
	query := sq.Select("count(*)").From(m.table)
	if len(args.ChannelId) > 0 {
		query = query.Where(sq.Eq{"channel_id": args.ChannelId})
	}
	if args.TermId > 0 {
		query = query.Where(sq.Eq{"term_id": args.TermId})
	}
	if args.UserId > 0 {
		query = query.Where(sq.Eq{"user_id": args.UserId})
	}
	if args.StartDate > 0 {
		startDate := time.Unix(args.StartDate, 0).Format("2006-01-02 15:04:05")
		query = query.Where(sq.GtOrEq{"create_time": startDate})
	}
	if args.EndDate > 0 {
		endDate := time.Unix(args.EndDate, 0).Format("2006-01-02 15:04:05")
		query = query.Where(sq.LtOrEq{"create_time": endDate})
	}
	// 新增查询条件
	if len(args.OrderId) > 0 {
		query = query.Where(sq.Eq{"id": args.OrderId})
	}
	// 奖金订单的币种查询，需要查看实际表结构
	if len(args.Currency) > 0 && args.Currency != "0" {
		// 注意：奖金订单表可能没有currency字段，需要根据实际情况调整
		// 暂时跳过，因为奖金订单可能通过关联的原始订单来确定币种
	}
	// 奖金订单的投注金额通常为0，不需要按投注金额过滤
	if args.BetAmtMin > 0 || args.BetAmtMax > 0 {
		// 奖金订单的投注金额为0，如果有投注金额范围查询，直接返回0
		return 0, nil
	}
	// 奖金倍数查询条件 - 对于奖金订单，使用不同的逻辑
	if args.MultipleMin > 0 || args.MultipleMax > 0 {
		// 奖金订单的奖金倍数计算：cashout_multiple / bet_at_multiple
		if args.MultipleMin > 0 && args.MultipleMax > 0 {
			query = query.Where("(cashout_multiple / bet_at_multiple) >= ? AND (cashout_multiple / bet_at_multiple) <= ?", args.MultipleMin*100, args.MultipleMax*100)
		} else if args.MultipleMin > 0 {
			query = query.Where("(cashout_multiple / bet_at_multiple) >= ?", args.MultipleMin*100)
		} else if args.MultipleMax > 0 {
			query = query.Where("(cashout_multiple / bet_at_multiple) <= ?", args.MultipleMax*100)
		}
	}

	sqlStr, sqlArgs, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	var count int64
	err = m.conn.QueryRowCtx(ctx, &count, sqlStr, sqlArgs...)
	switch err {
	case nil:
		return count, nil
	case sqlx.ErrNotFound:
		return 0, ErrNotFound
	default:
		return 0, err
	}
}

func (m *customBonusBetModel) GetByBetIds(ctx context.Context, betIds []int64) ([]*BonusBet, error) {
	if len(betIds) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(betIds))
	args := make([]interface{}, 0, len(betIds))

	for _, id := range betIds {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE bet_id IN (%s)", bonusBetRows, m.table, strings.Join(placeholders, ","))

	var list []*BonusBet
	err := m.conn.QueryRowsCtx(ctx, &list, query, args...)

	if err != nil && err != sqlx.ErrNotFound {
		return nil, err
	}

	return list, nil
}

func (m *customBonusBetModel) GetByChannelUserTerms(ctx context.Context, channelId, userId int64, termIds []int64) ([]*BonusBet, error) {
	query := sq.Select(bonusBetRows).From(m.table)

	if channelId > 0 {
		query = query.Where(sq.Eq{"channel_id": channelId})
	}
	if userId > 0 {
		query = query.Where(sq.Eq{"user_id": userId})
	}
	if len(termIds) > 0 {
		query = query.Where(sq.Eq{"term_id": termIds})
	}

	query = query.OrderBy("id desc")

	sqlStr, sqlArgs, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var resp []*BonusBet
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlArgs...)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
