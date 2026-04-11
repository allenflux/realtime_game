package gmmodel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CurrencyLimitModel = (*customCurrencyLimitModel)(nil)

type (
	// CurrencyLimitModel 是currency_limit表的操作接口
	CurrencyLimitModel interface {
		Insert(ctx context.Context, data *CurrencyLimit) (sql.Result, error)
		FindOne(ctx context.Context, id int64) (*CurrencyLimit, error)
		FindByCidAndCurrency(ctx context.Context, channelId int64, currency string) (*CurrencyLimit, error)
		FindByCidClientIdAndCurrency(ctx context.Context, channelId int64, clientId string, currency string) (*CurrencyLimit, error)
		Update(ctx context.Context, data *CurrencyLimit) error
		Delete(ctx context.Context, id int64) error
		GetPageByCidAndClientId(ctx context.Context, channelId int64, clientId string, pageSize, offset int) ([]*CurrencyLimit, error)
		GetNumByCidAndClientId(ctx context.Context, channelId int64, clientId string) (int64, error)
		DeleteByCidAndCurrency(ctx context.Context, channelId int64, currency string) error
		GetDistinctCurrencyAndPrecision(ctx context.Context) ([]*CurrencyLimit, error)
		GetCurrencyPrecisionByCurrencies(ctx context.Context, currencies []string) ([]*CurrencyLimit, error)
		GetCurrencyPrecisionByCurrency(ctx context.Context, currency string) (*CurrencyLimit, error)
		FindByChannelIdAndClientId(ctx context.Context, channelId int64, clientId string) ([]*CurrencyLimit, error)
		ListByChanIds(ctx context.Context, channelIds []int64) ([]*CurrencyLimit, error)
	}

	// CurrencyLimit 是数据库currency_limit表的映射
	CurrencyLimit struct {
		Id                int64     `db:"id"`
		ChannelId         int64     `db:"channel_id"`
		ClientId          string    `db:"client_id"`
		Currency          string    `db:"currency"`
		CurrencyPrecision float64   `db:"currency_precision"`
		MinBet            int64     `db:"min_bet"`
		MaxBet            int64     `db:"max_bet"`
		MaxProfit         int64     `db:"max_profit"`
		CreateTime        time.Time `db:"create_time"`
		UpdateTime        time.Time `db:"update_time"`
		IsActive          int64     `db:"is_active"`
	}

	customCurrencyLimitModel struct {
		conn  sqlx.SqlConn
		table string
	}
)

// NewCurrencyLimitModel 返回一个操作CurrencyLimit表的模型
func NewCurrencyLimitModel(conn sqlx.SqlConn) CurrencyLimitModel {
	return &customCurrencyLimitModel{
		conn:  conn,
		table: "`currency_limit`",
	}
}

// Insert 插入一条CurrencyLimit记录
func (m *customCurrencyLimitModel) Insert(ctx context.Context, data *CurrencyLimit) (sql.Result, error) {
	var fields []string
	var placeholders []string
	var args []interface{}

	// 必填字段
	fields = append(fields, "channel_id", "client_id", "currency")
	placeholders = append(placeholders, "?", "?", "?")
	args = append(args, data.ChannelId, data.ClientId, data.Currency)

	// 可选字段，有值才添加
	if data.CurrencyPrecision != 0 {
		fields = append(fields, "currency_precision")
		placeholders = append(placeholders, "?")
		args = append(args, data.CurrencyPrecision)
	}

	if data.MinBet == 0 {
		fields = append(fields, "min_bet")
		placeholders = append(placeholders, "?")
		args = append(args, data.MinBet)
	}

	if data.MaxBet == 0 {
		fields = append(fields, "max_bet")
		placeholders = append(placeholders, "?")
		args = append(args, data.MaxBet)
	}

	if data.MaxProfit == 0 {
		fields = append(fields, "max_profit")
		placeholders = append(placeholders, "?")
		args = append(args, data.MaxProfit)
	}

	// 添加时间字段
	now := time.Now()
	fields = append(fields, "create_time", "update_time")
	placeholders = append(placeholders, "?", "?")
	args = append(args, now, now)

	// 处理is_active字段
	if data.IsActive != 0 {
		fields = append(fields, "is_active")
		placeholders = append(placeholders, "?")
		args = append(args, data.IsActive)
	}

	query := fmt.Sprintf("insert into %s (%s) values (%s)",
		m.table,
		strings.Join(fields, ", "),
		strings.Join(placeholders, ", "))

	return m.conn.ExecCtx(ctx, query, args...)
}

// FindOne 根据id查询CurrencyLimit记录
func (m *customCurrencyLimitModel) FindOne(ctx context.Context, id int64) (*CurrencyLimit, error) {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", currencyLimitRows, m.table)
	var resp CurrencyLimit
	err := m.conn.QueryRowCtx(ctx, &resp, query, id)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByCidAndCurrency 根据channel_id和currency查询CurrencyLimit记录
func (m *customCurrencyLimitModel) FindByCidAndCurrency(ctx context.Context, channelId int64, currency string) (*CurrencyLimit, error) {
	query := fmt.Sprintf("select %s from %s where `channel_id` = ? and `currency` = ? limit 1", currencyLimitRows, m.table)
	var resp CurrencyLimit
	err := m.conn.QueryRowCtx(ctx, &resp, query, channelId, currency)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByCidClientIdAndCurrency 根据channel_id、client_id和currency查询CurrencyLimit记录
func (m *customCurrencyLimitModel) FindByCidClientIdAndCurrency(ctx context.Context, channelId int64, clientId string, currency string) (*CurrencyLimit, error) {
	query := fmt.Sprintf("select %s from %s where `channel_id` = ? and `client_id` = ? and `currency` = ? limit 1", currencyLimitRows, m.table)
	var resp CurrencyLimit
	err := m.conn.QueryRowCtx(ctx, &resp, query, channelId, clientId, currency)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// Update 更新CurrencyLimit记录
func (m *customCurrencyLimitModel) Update(ctx context.Context, data *CurrencyLimit) error {
	query := fmt.Sprintf("update %s set min_bet = ?, max_bet = ?, max_profit = ?, update_time = ?, is_active = ? where id = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, data.MinBet, data.MaxBet, data.MaxProfit, time.Now(), data.IsActive, data.Id)
	return err
}

// Delete 删除CurrencyLimit记录
func (m *customCurrencyLimitModel) Delete(ctx context.Context, id int64) error {
	query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, id)
	return err
}

// DeleteByCidAndCurrency 根据channel_id和currency删除CurrencyLimit记录
func (m *customCurrencyLimitModel) DeleteByCidAndCurrency(ctx context.Context, channelId int64, currency string) error {
	query := fmt.Sprintf("delete from %s where `channel_id` = ? and `currency` = ?", m.table)
	_, err := m.conn.ExecCtx(ctx, query, channelId, currency)
	return err
}

// GetPageByCidAndClientId 分页获取货币限额配置
func (m *customCurrencyLimitModel) GetPageByCidAndClientId(ctx context.Context, channelId int64, clientId string, pageSize, offset int) ([]*CurrencyLimit, error) {
	var condition []string
	var args []interface{}

	if channelId > 0 {
		condition = append(condition, "`channel_id` = ?")
		args = append(args, channelId)
	}

	if len(clientId) > 0 {
		condition = append(condition, "`client_id` = ?")
		args = append(args, clientId)
	}

	condStr := ""
	if len(condition) > 0 {
		condStr = "where " + strings.Join(condition, " and ")
	}

	query := fmt.Sprintf("select %s from %s %s order by currency limit ? offset ?", currencyLimitRows, m.table, condStr)
	args = append(args, pageSize, offset)

	var resp []*CurrencyLimit
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// GetNumByCidAndClientId 获取符合条件的记录总数
func (m *customCurrencyLimitModel) GetNumByCidAndClientId(ctx context.Context, channelId int64, clientId string) (int64, error) {
	var condition []string
	var args []interface{}

	if channelId > 0 {
		condition = append(condition, "`channel_id` = ?")
		args = append(args, channelId)
	}

	if len(clientId) > 0 {
		condition = append(condition, "`client_id` = ?")
		args = append(args, clientId)
	}

	condStr := ""
	if len(condition) > 0 {
		condStr = "where " + strings.Join(condition, " and ")
	}

	query := fmt.Sprintf("select count(1) from %s %s", m.table, condStr)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, args...)
	switch err {
	case nil:
		return count, nil
	default:
		return 0, err
	}
}

// GetDistinctCurrencyAndPrecision 获取去重后的货币和精度信息
func (m *customCurrencyLimitModel) GetDistinctCurrencyAndPrecision(ctx context.Context) ([]*CurrencyLimit, error) {
	query := fmt.Sprintf("select %s from %s GROUP BY currency order by currency", currencyLimitRows, m.table)
	var resp []*CurrencyLimit
	err := m.conn.QueryRowsCtx(ctx, &resp, query)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// GetCurrencyPrecisionByCurrencies 根据货币列表查询精度信息
func (m *customCurrencyLimitModel) GetCurrencyPrecisionByCurrencies(ctx context.Context, currencies []string) ([]*CurrencyLimit, error) {
	if len(currencies) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(currencies))
	args := make([]interface{}, len(currencies))
	for i, currency := range currencies {
		placeholders[i] = "?"
		args[i] = currency
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE `currency` IN (%s) GROUP BY `currency` order by currency", currencyLimitRows,
		m.table, strings.Join(placeholders, ","))

	var resp []*CurrencyLimit
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// GetCurrencyPrecisionByCurrency 根据货币查询精度信息
func (m *customCurrencyLimitModel) GetCurrencyPrecisionByCurrency(ctx context.Context, currency string) (*CurrencyLimit, error) {
	query := fmt.Sprintf("SELECT %s FROM %s WHERE `currency` = ? GROUP BY `currency` order by currency", currencyLimitRows, m.table)
	var resp CurrencyLimit
	err := m.conn.QueryRowCtx(ctx, &resp, query, currency)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customCurrencyLimitModel) FindByChannelIdAndClientId(ctx context.Context, channelId int64, clientId string) ([]*CurrencyLimit, error) {
	query := fmt.Sprintf("select %s from %s where `channel_id` = ? and `client_id` = ? and is_active = ? limit 100", currencyLimitRows, m.table)
	var resp []*CurrencyLimit
	err := m.conn.QueryRowsCtx(ctx, &resp, query, channelId, clientId, 1)
	switch err {
	case nil:
		return resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

const currencyLimitRows = "`id`, `channel_id`, `client_id`, `currency`, `currency_precision`, `min_bet`, `max_bet`, `max_profit`, `create_time`, `update_time`, `is_active`"

// ListByChanIds 根据渠道ID列表获取货币限额配置
func (m *customCurrencyLimitModel) ListByChanIds(ctx context.Context, channelIds []int64) ([]*CurrencyLimit, error) {
	if len(channelIds) == 0 {
		return nil, errors.New("channelIds cannot be empty")
	}

	// 构建 WHERE IN 查询
	placeholders := make([]string, len(channelIds))
	args := make([]interface{}, len(channelIds))

	for i, id := range channelIds {
		placeholders[i] = "?"
		args[i] = id
	}

	// 使用 WHERE IN 查询多个channel_id
	condition := fmt.Sprintf("channel_id IN (%s)", strings.Join(placeholders, ","))
	query := fmt.Sprintf("SELECT %s FROM %s WHERE %s", currencyLimitRows, m.table, condition)

	var resp []*CurrencyLimit
	err := m.conn.QueryRowsCtx(ctx, &resp, query, args...)
	switch {
	case err == nil:
		return resp, nil
	case errors.Is(err, sqlx.ErrNotFound): // 注意：这里应该是 sqlx.ErrNotFound
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
