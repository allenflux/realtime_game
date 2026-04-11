package servermodel

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	Channel_F_id                      = "id"
	Channel_F_client_id               = "client_id"
	Channel_F_client_name             = "client_name"
	Channel_F_channel_type            = "channel_type"
	Channel_F_game_name               = "game_name"
	Channel_F_is_active               = "is_active"
	Channel_F_tz                      = "tz"
	Channel_F_ctrl_hours              = "ctrl_hours"
	Channel_F_service_fee             = "service_fee"
	Channel_F_rake                    = "rake"
	Channel_F_ctrl_trigger_rate       = "ctrl_trigger_rate"
	Channel_F_ctrl_put_rate           = "ctrl_put_rate"
	Channel_F_total_profit            = "total_profit"
	Channel_F_ctrl_profit             = "ctrl_profit"
	Channel_F_ctrl_coef               = "ctrl_coef"
	Channel_F_divisor                 = "divisor"
	Channel_F_inc_num                 = "inc_num"
	Channel_F_min_bet_num             = "min_bet_num"
	Channel_F_max_bet_num             = "max_bet_num"
	Channel_F_max_cashout_multiple    = "max_cashout_multiple"
	Channel_F_max_cashout_per_bet     = "max_cashout_per_bet"
	Channel_F_next_rand_multiple      = "next_rand_multiple"
	Channel_F_next_rand_multiple_used = "next_rand_multiple_used"
	Channel_F_break_payout_rate       = "break_payout_rate"
	Channel_F_user_profit_correct     = "user_profit_correct"
	Channel_F_next_is_ctrl            = "next_is_ctrl"
	Channel_F_total_bet_amt           = "total_bet_amt"
	Channel_F_total_cashed_amt        = "total_cashed_amt"
	Channel_F_bet_amt_decimal         = "bet_amt_decimal"
	Channel_F_create_time             = "create_time"
	Channel_F_update_time             = "update_time"
	Channel_F_stat_time_offset        = "stat_time_offset"
	Channel_F_agent_time_zone         = "agent_time_zone"
	Channel_F_ui_layout               = "ui_layout"
	Channel_F_default_line_status     = "default_line_status"
)

const (
	//下局随机爆点结果是否被使用 1=否 2=是
	CHANNEL_next_rand_multiple_used_no  = 1
	CHANNEL_next_rand_multiple_used_yes = 2

	//是否运行 1=正常 2=禁止
	CHANNEL_is_active_yes = 1
	CHANNEL_is_active_no  = 2

	//下局是否控制 1=否 2=是
	CHANNEL_next_is_ctrl_no  = 1
	CHANNEL_next_is_ctrl_yes = 2
)

var _ ChannelModel = (*customChannelModel)(nil)

type (
	// ChannelModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChannelModel.
	ChannelModel interface {
		channelModel
		withSession(session sqlx.Session) ChannelModel
		GetAllActiveGames(ctx context.Context) ([]*Channel, error)
		GetByIds(ctx context.Context, ids []int64) ([]*Channel, error)
		GetPageByIdAndStatus(ctx context.Context, ids []int64, isActivity int, gameName string, pageSize int, start int) ([]Channel, error)
		GetNumByIdAndStatus(ctx context.Context, ids []int64, isActivity int, gameName string) (int, error)
		UpdateById(ctx context.Context, id int64, data map[string]interface{}) error
		GetByClientId(ctx context.Context, clientId string) ([]*Channel, error)
		AddNew(ctx context.Context, data map[string]interface{}) (sql.Result, error)
		QueryChannelByClientIdAndGameName(ctx context.Context, clientId int64, gameName string) ([]*Channel, error)
		QueryChannels(ctx context.Context, clientIds []int64) ([]*Channel, error)
		QueryAllClientIds(ctx context.Context) ([]int64, error)
	}

	customChannelModel struct {
		*defaultChannelModel
	}
)

// NewChannelModel returns a model for the database table.
func NewChannelModel(conn sqlx.SqlConn) ChannelModel {
	return &customChannelModel{
		defaultChannelModel: newChannelModel(conn),
	}
}

func (m *customChannelModel) withSession(session sqlx.Session) ChannelModel {
	return NewChannelModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customChannelModel) GetAllActiveGames(ctx context.Context) ([]*Channel, error) {
	sqlStr, _, _ := sq.Select(channelFieldNames...).From(m.table).Where("is_active = 1").ToSql()
	var channels []*Channel
	if err := m.conn.QueryRowsCtx(ctx, &channels, sqlStr); err != nil {
		return nil, err
	}
	return channels, nil
}

// 根据ids 获取数据
func (m *customChannelModel) GetByIds(ctx context.Context, ids []int64) ([]*Channel, error) {
	sqlStr, sqlParams, _ := sq.Select(channelFieldNames...).From(m.table).Where(sq.Eq{
		Channel_F_id: ids,
	}).ToSql()
	var channels []*Channel
	err := m.conn.QueryRowsCtx(ctx, &channels, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	if err != nil {
		return nil, err
	}
	return channels, nil
}

// 根据id和状态获取一页数据
func (m *defaultChannelModel) GetPageByIdAndStatus(ctx context.Context, ids []int64, isActivity int, gameName string, pageSize int, start int) ([]Channel, error) {
	sqb := sq.Select("*").From(m.table)
	if len(ids) > 0 {
		sqb = sqb.Where(sq.Eq{Channel_F_id: ids})
	}
	if isActivity > 0 {
		sqb = sqb.Where(sq.Eq{Channel_F_is_active: isActivity})
	}

	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{Channel_F_game_name: gameName})
	}
	sqlStr, sqlParams, _ := sqb.OrderBy(Channel_F_is_active+" asc ", Channel_F_client_id+" desc ").Offset(uint64(start)).Limit(uint64(pageSize)).ToSql()
	resp := make([]Channel, 0)
	err := m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

func (m *defaultChannelModel) GetNumByIdAndStatus(ctx context.Context, ids []int64, isActivity int, gameName string) (int, error) {
	sqb := sq.Select("count(*) as num").From(m.table)
	if len(ids) > 0 {
		sqb = sqb.Where(sq.Eq{Channel_F_id: ids})
	}
	if isActivity > 0 {
		sqb = sqb.Where(sq.Eq{Channel_F_is_active: isActivity})
	}
	if len(gameName) > 0 {
		sqb = sqb.Where(sq.Eq{Channel_F_game_name: gameName})
	}

	sqlStr, sqlParams, _ := sqb.ToSql()
	resp := 0
	err := m.conn.QueryRowCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

func (m *defaultChannelModel) UpdateById(ctx context.Context, id int64, data map[string]interface{}) error {
	if id == 0 || len(data) == 0 {
		return nil
	}
	sqlStr, sqlParams, err := sq.Update(m.table).Where(sq.Eq{
		Channel_F_id: id,
	}).SetMap(data).ToSql()
	if err != nil {
		return err
	}
	_, err = m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	if err != nil {
		return err
	}
	return nil
}

// 根据apisys的clientId, 获取一条数据
func (m *defaultChannelModel) GetByClientId(ctx context.Context, clientId string) (resp []*Channel, err error) {
	resp = make([]*Channel, 0)
	if clientId == "" {
		return
	}
	sqlStr, sqlParams, _ := sq.Select(channelFieldNames...).From(m.table).Where(sq.Eq{
		Channel_F_client_id: clientId,
	}).ToSql()
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	if err != nil {
		return
	}

	return
}

// 不加锁，不考虑并发。
func (m *defaultChannelModel) AddNew(ctx context.Context, data map[string]interface{}) (sql.Result, error) {
	if len(data) == 0 {
		return nil, errors.New("data is nil")
	}

	var maxID int64
	query := fmt.Sprintf("SELECT IFNULL(MAX(id), 0) FROM %s WHERE id < 19000", m.table)
	if err := m.conn.QueryRowCtx(ctx, &maxID, query); err != nil {
		return nil, fmt.Errorf("query max id error: %w", err)
	}

	nextID := maxID + 1
	ins := make(map[string]interface{}, len(data)+1)
	for k, v := range data {
		if strings.EqualFold(k, "id") {
			continue
		}
		ins[k] = v
	}
	ins["id"] = nextID

	sqlStr, sqlArgs, _ := sq.Insert(m.table).SetMap(ins).ToSql()
	return m.conn.ExecCtx(ctx, sqlStr, sqlArgs...)
}

// QueryChannelByClientIdAndGameName 根据apisys的merchant_id -> clientId和game_type->game_name, 批量获取数据
func (m *defaultChannelModel) QueryChannelByClientIdAndGameName(ctx context.Context, clientId int64, gameName string) ([]*Channel, error) {
	// 参数校验
	if clientId == 0 && gameName == "" {
		return nil, errors.New("clientId and gameName cannot both be empty")
	}

	// 构建查询条件
	builder := sq.Select(channelFieldNames...).From(m.table)

	// 添加查询条件
	conditions := make(sq.Eq)
	if clientId != 0 {
		conditions[Channel_F_client_id] = clientId
	}
	if gameName != "" {
		conditions[Channel_F_game_name] = gameName
	}

	builder = builder.Where(conditions)

	// 生成SQL
	sqlStr, sqlParams, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	// 执行查询
	var resp []*Channel
	if err := m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*Channel{}, nil // 返回空切片而不是nil
		}
		return nil, fmt.Errorf("failed to query channels: %w", err)
	}

	return resp, nil
}

func (m *defaultChannelModel) QueryChannels(ctx context.Context, clientIds []int64) ([]*Channel, error) {
	// 参数校验
	if len(clientIds) == 0 {
		return nil, errors.New("clientIds cannot be empty")
	}

	// 将 []int64 转换为 []interface{}
	interfaceIds := make([]interface{}, len(clientIds))
	for i, id := range clientIds {
		interfaceIds[i] = id
	}

	// 使用 Squirrel 构建 WHERE IN 查询
	builder := sq.Select(channelFieldNames...).
		From(m.table).
		Where(sq.Eq{"client_id": interfaceIds})

	// 生成SQL
	sqlStr, sqlParams, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	// 执行查询
	var resp []*Channel
	if err := m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...); err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []*Channel{}, nil // 返回空切片而不是nil
		}
		return nil, fmt.Errorf("failed to query channels - 2: %w", err)
	}

	return resp, nil
}

func (m *defaultChannelModel) QueryAllClientIds(ctx context.Context) ([]int64, error) {
	// 明确排除的 client_id（不用 interface{}）
	excludeClientIDs := []int64{199998, 199999}

	builder := sq.
		Select(channelFieldNames...).
		From(m.table).
		Where(sq.NotEq{"client_id": excludeClientIDs})

	sqlStr, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build sql failed: %w", err)
	}

	var channels []*Channel
	err = m.conn.QueryRowsCtx(ctx, &channels, sqlStr, args...)
	if err != nil {
		if errors.Is(err, sqlx.ErrNotFound) {
			return []int64{}, nil
		}
		return nil, fmt.Errorf("query channels failed: %w", err)
	}

	ids := make([]int64, 0, len(channels))
	for _, c := range channels {
		i, err := strconv.ParseInt(c.ClientId, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, i)
	}

	return ids, nil
}
