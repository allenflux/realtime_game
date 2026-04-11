package gmmodel

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/shopspring/decimal"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	UserProfit_F_id                = "id"
	UserProfit_F_user_id           = "user_id"
	UserProfit_F_channel_id        = "channel_id"
	UserProfit_F_game_name         = "game_name"
	UserProfit_F_bet_amt_3days     = "bet_amt_3days"
	UserProfit_F_pump_amt_3days    = "pump_amt_3days"
	UserProfit_F_cashout_amt_3days = "cashout_amt_3days"
	UserProfit_F_profit_amt_3days  = "profit_amt_3days"
	UserProfit_F_rate_3days        = "rate_3days"
	UserProfit_F_bet_amt_7days     = "bet_amt_7days"
	UserProfit_F_pump_amt_7days    = "pump_amt_7days"
	UserProfit_F_cashout_amt_7days = "cashout_amt_7days"
	UserProfit_F_profit_amt_7days  = "profit_amt_7days"
	UserProfit_F_rate_7days        = "rate_7days"
	UserProfit_F_total_bet_amt     = "total_bet_amt"
	UserProfit_F_total_pupm_amt    = "total_pupm_amt"
	UserProfit_F_total_cashout_amt = "total_cashout_amt"
	UserProfit_F_total_profit_amt  = "total_profit_amt"
	UserProfit_F_total_rate        = "total_rate"
	UserProfit_F_status            = "status"
	UserProfit_F_create_time       = "create_time"
)

var _ UserProfitModel = (*customUserProfitModel)(nil)

type UserProfitGetPageArgs struct {
	ChannelIds     []int64 //渠道ids
	GameName       string  //游戏名称
	UserId         int64   //用户id
	Status         int64   //状态
	StartBetAmt    float64 //投注总额范围，起始值
	EndBetAmt      float64 //投注总额范围，结束值
	StartProfitAmt float64 //盈亏总额范围，起始值
	EndProfitAmt   float64 //盈亏总额范围，结束值
	StartRate      float64 //总杀率范围，起始值
	EndRate        float64 //总杀率范围，结束值
	Start          int     //页码
	PageSize       int     //页数
}

type (
	// UserProfitModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserProfitModel.
	UserProfitModel interface {
		userProfitModel
		withSession(session sqlx.Session) UserProfitModel
		GetPage(ctx context.Context, args *UserProfitGetPageArgs) ([]*UserProfit, error)
		GetPageNum(ctx context.Context, args *UserProfitGetPageArgs) (int, error)
		GetByUseridAndChannelId(ctx context.Context, userid int64, channelId int64, gameName string) (*UserProfit, error)
		UpdateById(ctx context.Context, id int64, data map[string]interface{}) error
		GetPageById(ctx context.Context, id int64, pageSize int64) ([]*UserProfit, error)
		GetTotalProfitAmt(ctx context.Context, userids []int64, channelId int64) (decimal.Decimal, error)
	}

	customUserProfitModel struct {
		*defaultUserProfitModel
	}
)

// NewUserProfitModel returns a model for the database table.
func NewUserProfitModel(conn sqlx.SqlConn) UserProfitModel {
	return &customUserProfitModel{
		defaultUserProfitModel: newUserProfitModel(conn),
	}
}

func (m *customUserProfitModel) withSession(session sqlx.Session) UserProfitModel {
	return NewUserProfitModel(sqlx.NewSqlConnFromSession(session))
}

// 根据部分条件，获取一页数据
func (m *customUserProfitModel) GetPage(ctx context.Context, args *UserProfitGetPageArgs) (resp []*UserProfit, err error) {
	resp = make([]*UserProfit, 0)
	sqb := sq.Select("*").From(m.table)
	if len(args.ChannelIds) > 0 {
		sqb = sqb.Where(sq.Eq{UserProfit_F_channel_id: args.ChannelIds})
	}
	if args.GameName != "" {
		sqb = sqb.Where(sq.Eq{UserProfit_F_game_name: args.GameName})
	}
	if args.UserId != 0 {
		sqb = sqb.Where(sq.Eq{UserProfit_F_user_id: args.UserId})
	}
	if args.Status != 0 {
		sqb = sqb.Where(sq.Eq{UserProfit_F_status: args.Status})
	}
	if args.StartBetAmt != 0 {
		sqb = sqb.Where(sq.GtOrEq{UserProfit_F_total_bet_amt: args.StartBetAmt})
	}
	if args.EndBetAmt != 0 {
		sqb = sqb.Where(sq.LtOrEq{UserProfit_F_total_bet_amt: args.EndBetAmt})
	}
	if args.StartProfitAmt > 0 {
		sqb = sqb.Where(sq.GtOrEq{UserProfit_F_total_profit_amt: args.StartProfitAmt})
	}
	if args.EndProfitAmt != 0 {
		sqb = sqb.Where(sq.LtOrEq{UserProfit_F_total_profit_amt: args.StartProfitAmt})
	}
	if args.StartRate != 0 {
		sqb = sqb.Where(sq.GtOrEq{UserProfit_F_total_rate: args.StartRate})
	}
	if args.EndRate != 0 {
		sqb = sqb.Where(sq.LtOrEq{UserProfit_F_total_rate: args.EndRate})
	}

	sqlStr, sqlParams, _ := sqb.Offset(uint64(args.Start)).Limit(uint64(args.PageSize)).ToSql()
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

func (m *customUserProfitModel) GetPageNum(ctx context.Context, args *UserProfitGetPageArgs) (int, error) {
	sqb := sq.Select(" count(*) as num").From(m.table)
	if len(args.ChannelIds) > 0 {
		sqb = sqb.Where(sq.Eq{UserProfit_F_channel_id: args.ChannelIds})
	}
	if args.GameName != "" {
		sqb = sqb.Where(sq.Eq{UserProfit_F_game_name: args.GameName})
	}
	if args.UserId != 0 {
		sqb = sqb.Where(sq.Eq{UserProfit_F_user_id: args.UserId})
	}
	if args.Status != 0 {
		sqb = sqb.Where(sq.Eq{UserProfit_F_status: args.Status})
	}
	if args.StartBetAmt != 0 {
		sqb = sqb.Where(sq.GtOrEq{UserProfit_F_total_bet_amt: args.StartBetAmt})
	}
	if args.EndBetAmt != 0 {
		sqb = sqb.Where(sq.LtOrEq{UserProfit_F_total_bet_amt: args.EndBetAmt})
	}
	if args.StartProfitAmt > 0 {
		sqb = sqb.Where(sq.GtOrEq{UserProfit_F_total_profit_amt: args.StartProfitAmt})
	}
	if args.EndProfitAmt != 0 {
		sqb = sqb.Where(sq.LtOrEq{UserProfit_F_total_profit_amt: args.StartProfitAmt})
	}
	if args.StartRate != 0 {
		sqb = sqb.Where(sq.GtOrEq{UserProfit_F_total_rate: args.StartRate})
	}
	if args.EndRate != 0 {
		sqb = sqb.Where(sq.LtOrEq{UserProfit_F_total_rate: args.EndRate})
	}

	resp := 0
	sqlStr, sqlParams, _ := sqb.ToSql()
	err := m.conn.QueryRowCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

// 根据userid和渠道id 获取一条用户信息
func (m *customUserProfitModel) GetByUseridAndChannelId(ctx context.Context, userid int64, channelId int64, gameName string) (resp *UserProfit, err error) {
	resp = &UserProfit{}
	if userid == 0 || channelId == 0 {
		return
	}
	sqlStr, sqlParams, _ := sq.Select("*").From(m.table).
		Where(sq.Eq{UserProfit_F_user_id: userid}).
		Where(sq.Eq{UserProfit_F_channel_id: channelId}).
		Where(sq.Eq{UserProfit_F_game_name: gameName}).ToSql()
	err = m.conn.QueryRowCtx(ctx, resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

// 修改数据
func (m *customUserProfitModel) UpdateById(ctx context.Context, id int64, data map[string]interface{}) error {
	if id == 0 || len(data) == 0 {
		return nil
	}
	sqlStr, sqlParams, _ := sq.Update(m.table).Where(sq.Eq{UserProfit_F_id: id}).SetMap(data).ToSql()
	_, err := m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	return err
}

// 获取一页数据
func (m *customUserProfitModel) GetPageById(ctx context.Context, id int64, pageSize int64) ([]*UserProfit, error) {
	resp := make([]*UserProfit, 0)
	if id < 0 || pageSize < 0 {
		return resp, nil
	}

	sqb := sq.Select("*").From(m.table)
	if id > 0 {
		sqb = sqb.Where(sq.Gt{UserProfit_F_id: id})
	}
	sqlStr, sqlParams, _ := sqb.OrderBy(UserProfit_F_id + " asc").Limit(uint64(pageSize)).ToSql()
	err := m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err != nil && err != sql.ErrNoRows {
		return resp, err
	}

	return resp, nil
}

func (m *customUserProfitModel) GetTotalProfitAmt(ctx context.Context, userids []int64, channelId int64) (resp decimal.Decimal, err error) {
	resp = decimal.NewFromInt(0)
	if len(userids) == 0 {
		return
	}
	sqlStr, sqlParams, _ := sq.Select(" COALESCE(sum(total_profit_amt), '0') as total_profit").From(m.table).
		Where(sq.Eq{
			UserProfit_F_user_id:    userids,
			UserProfit_F_channel_id: channelId,
		}).ToSql()
	totalProfit := "0"
	err = m.conn.QueryRowCtx(ctx, &totalProfit, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	resp, _ = decimal.NewFromString(totalProfit)
	return
}
