package servermodel

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"strings"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/errgroup"
)

const (
	Bet_F_id                            = "id"
	Bet_F_api_order_no                  = "api_order_no"
	Bet_F_channel_id                    = "channel_id"
	Bet_F_real_channel_id               = "real_channel_id"
	Bet_F_runtime_channel_id            = "runtime_channel_id"
	Bet_F_term_id                       = "term_id"
	Bet_F_user_id                       = "user_id"
	Bet_F_user_name                     = "user_name"
	Bet_F_bet_type                      = "bet_type"
	Bet_F_amount                        = "amount"
	Bet_F_currency                      = "currency"
	Bet_F_auto_cashout_multiple         = "auto_cashout_multiple"
	Bet_F_manual_cashout_multiple       = "manual_cashout_multiple"
	Bet_F_bet_at_multiple               = "bet_at_multiple"
	Bet_F_service_fee                   = "service_fee"
	Bet_F_rake                          = "rake"
	Bet_F_rake_amt                      = "rake_amt"
	Bet_F_cashed_out_amount             = "cashed_out_amount"
	Bet_F_in_retry                      = "in_retry"
	Bet_F_order_status                  = "order_status"
	Bet_F_max_cashedout_bet_id          = "max_cashedout_bet_id"
	Bet_F_max_multiple                  = "max_multiple"
	Bet_F_create_time                   = "create_time"
	Bet_F_update_time                   = "update_time"
	Bet_F_ctime                         = "ctime"
	Bet_F_game_play                     = "game_play"
	Bet_F_bonus_bet_id                  = "bonus_bet_id"
	Bet_F_game_name                     = "game_name"
	Bet_F_First_Cashout_Amount          = "first_cashout_amount"
	Bet_F_First_Manual_Cashout_Multiple = "first_manual_cashout_multiple"
	Bet_F_Cashout_Times                 = "manual_cashout_times"
)

const (
	// 【注单状态】
	// 创建时状态
	OrderStatusCreating       = 1000  // 创建中
	OrderStatusCreated        = 2000  // 已创建
	OrderStatusCreationFailed = 10100 // 创建失败

	// 兑现时状态
	OrderStatusCashingOut = 3000 // 兑现中
	OrderStatusCashedout  = 4000 // 已兑现

	// 退款时状态
	OrderStatusRefunding = 5000 // 退款中
	OrderStatusRefunded  = 6000 // 已退款

	// 订单状态结束重试中
	OrderStatusOverRetry = 9999

	//注单状态 1=游戏前 2=游戏中
	BET_TYPE_pre = 1
	BET_TYPE_ing = 2

	// 兑现状态 1=已兑现 2=未逃生
	CASHOUT_STATUS_suc  = 1
	CASHOUT_STATUS_fail = 2

	// 是否进入重试任务 1=是 2=否
	BET_IN_RETRY_yes = 1
	BET_IN_RETRY_no  = 2
)

const Multiple_Tail = 10000

type UserBetStatis struct {
	TotalAmt        int64 `db:"totalAmt"`
	TotalCashoutAmt int64 `db:"totalCashoutAmt"`
}

var _ BetModel = (*customBetModel)(nil)

// 后台注单列表入参
type GetAdminPageArgs struct {
	ChannelId     []int64 `json:"channel_id,optional"`                               //渠道ids
	TermId        int64   `json:"term_id,optional"`                                  //局id
	UserId        int64   `json:"user_id,optional"`                                  //用户id
	GameName      string  `json:"game_name,optional"`                                //游戏名称
	BetType       int64   `json:"bet_type,optional,default=0,options=[0,1,2,3]"`     // 订单类型 1=赛前 2=滚盘 3=奖金
	CashoutStatus int64   `json:"cashout_status,optional,default=0,options=[0,1,2]"` //兑现状态 1=已兑现 2=未逃生
	StartDate     int64   `json:"start_date,optional"`                               //开始时间
	EndDate       int64   `json:"end_date,optional"`                                 //结束时间
	OrderId       string  `json:"order_id,optional"`                                 // 注单号，精确查询
	Currency      string  `json:"currency,optional"`                                 // 币种，默认选择全部
	BetAmtMin     int64   `json:"bet_amt_min,optional"`                              // 投注金额最小值
	BetAmtMax     int64   `json:"bet_amt_max,optional"`                              // 投注金额最大值
	MultipleMin   float64 `json:"multiple_min,optional"`                             // 奖金倍数最小值
	MultipleMax   float64 `json:"multiple_max,optional"`                             // 奖金倍数最大值
}

type GetListOfWagersReq struct {
	UserIds       string `form:"user_ids,optional"`                                       // 会员用户ID，逗号分隔：1,2,3
	Currency      string `form:"currency,optional"`                                       // 币种
	ParentWagerNo string `form:"parent_wager_no,optional"`                                // 总订单号
	WagerNo       string `form:"wager_no,optional"`                                       // 子订单号
	TicketNo      string `form:"ticket_no,optional"`                                      // 游戏商订单号
	Status        *int64 `form:"status,optional"`                                         // 子订单状态
	DateType      string `form:"date_type,optional,options=[wager_time,settlement_time]"` // 日期筛选类型
	FromDate      int64  `form:"from_date,optional"`                                      // 起始时间(UTC秒)
	ToDate        int64  `form:"to_date,optional"`                                        // 截止时间(UTC秒)
	Page          int64  `form:"page,optional,default=1"`                                 // 页码，默认1
	Size          int64  `form:"size,optional,default=100"`                               // 每页大小，默认100
}

type (
	// BetModel is an interface to be customized, add more methods here,
	// and implement the added methods in customBetModel.
	BetModel interface {
		betModel
		withSession(session sqlx.Session) BetModel
		GetPageById(ctx context.Context, orderId int64, pageSize int) ([]*Bet, error)
		GetByApiOrderNo(ctx context.Context, orderNo string) (*Bet, error)
		UpdateById(ctx context.Context, id int64, data map[string]interface{}) error
		UpdateByIds(ctx context.Context, ids []int64, data map[string]interface{}) error
		GetByApiOrderNos(ctx context.Context, orderNos []string) ([]*Bet, error)
		GetUncashedBetsByTerm(ctx context.Context, term *CrashTerm) (bets []*Bet, err error)
		GetNormalByTermID(ctx context.Context, channelID int64, termID int64) (bets []*Bet, err error)
		GetByRuntimeChannelAndTerm(ctx context.Context, runtimeChannelID int64, termID int64) (bets []*Bet, err error)
		GetById(ctx context.Context, id int64) (*Bet, error)
		GetUnrefundBetsByTerm(ctx context.Context, term *CrashTerm) (bets []*Bet, err error)
		GetByIds(ctx context.Context, ids []int64) ([]*Bet, error)
		GetBetsTodayBest(ctx context.Context, channelId, userId int64, startTime, endTime time.Time) (resp []*Bet, err error)
		GetByUserCurrencyCreateTime(ctx context.Context, currency string, userId int64, startTime time.Time) ([]*Bet, error)
		GetBetsByTime(ctx context.Context, currency string, channelId, userId int64, startTime, endTime time.Time) ([]*Bet, error)
		GetBetsByTimePage(ctx context.Context, currency string, userId int64, page, pageSize uint64, startTime, endTime time.Time) ([]*Bet, error)
		GetBetsByTimeTotal(ctx context.Context, currency string, userId int64, startTime, endTime time.Time) (int64, error)
		GetUserBetsById(ctx context.Context, userId, channelId int64, posId int64, startTime, endTime time.Time, pageSize int, currency string) ([]*Bet, error)
		GetUserBets(ctx context.Context, userId, channelId int64, posId int64, startTime, endTime time.Time, pageSize int, currency string) ([]*Bet, error)
		GetBetsByItemid(ctx context.Context, channelID int64, itemId int64, posId int64, pageSize int64) ([]*Bet, error)
		GetBetsByItemIDChannelIDs(ctx context.Context, channelID []int64, itemId int64, posId int64, pageSize int64) (resp []*Bet, err error)
		GetAllBetsByItemId(ctx context.Context, channelID int64, itemId int64) ([]*Bet, error)
		GetTermCashoutBets(ctx context.Context, channelID []int64, itemId int64) (resp []*Bet, err error)
		GetMaxAutoMultiple(ctx context.Context, channelId int64, startTime time.Time, pageSize int) ([]*Bet, error)
		GetMaxManualMultiple(ctx context.Context, channelId int64, startTime time.Time, pageSize int) ([]*Bet, error)
		GetUserBetStatis(ctx context.Context, userId, channelId int64, startTime, endTime time.Time) (*UserBetStatis, error)
		GetAdminPage(ctx context.Context, args *GetAdminPageArgs, start, pageSize int) ([]*Bet, error)
		GetAdminBetRows(ctx context.Context, args *GetAdminPageArgs, page, pageSize int64) (rows []*Row, totalRows int64, err error)
		CountAdminBetRows(ctx context.Context, args *GetAdminPageArgs) (totalRows int64, err error)
		GetAdminPageNum(ctx context.Context, args *GetAdminPageArgs) (int, error)
		GetBetsByTimePageGroupByOrderNo(ctx context.Context, channelId int64, currency string, userId int64, page, pageSize uint64, startTime, endTime time.Time) (resp []*Bet, err error)
		GetBetsByTimeTotalGroupByOrderNo(ctx context.Context, channelId int64, currency string, userId int64, startTime, endTime time.Time) (resp int64, err error)
		GetAdminPageUnion(ctx context.Context, args *GetAdminPageArgs, offset, limit int) ([]*BetUnion, int64, error)
		ListBetsByChannelIDs(ctx context.Context, channelIDs []int64, req *GetListOfWagersReq, page, size int64) (list []*Bet, total int64, err error)
		GroupSumsByChannelRows(
			ctx context.Context,
			channelIDs []int64,
			userIDs []int64,
			page, size int64,
			groupBy string,
			dateField string,
			fromTS, toTS int64,
			lastMaxID *int64,
		) (rows []RowAgg, totalGroups int64, err error)
		ListCurrencies(ctx context.Context) (resp []string, err error)
		Table() string
		Conn() sqlx.SqlConn
	}

	customBetModel struct {
		*defaultBetModel
	}
)

// NewBetModel returns a model for the database table.
func NewBetModel(conn sqlx.SqlConn) BetModel {
	return &customBetModel{
		defaultBetModel: newBetModel(conn),
	}
}

func (m *customBetModel) withSession(session sqlx.Session) BetModel {
	return NewBetModel(sqlx.NewSqlConnFromSession(session))
}

// 获取大于id的一页数据
func (m *customBetModel) GetPageById(ctx context.Context, orderId int64, pageSize int) (resp []*Bet, err error) {
	sqlStr, sqlParams, _ := sq.Select(betFieldNames...).From(m.table).
		Where(sq.Gt{Bet_F_id: orderId}).OrderBy(Bet_F_id + " asc ").
		Limit(uint64(pageSize)).ToSql()
	resp = make([]*Bet, 0)
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 根据订单号获取一条数据
func (m *customBetModel) GetByApiOrderNo(ctx context.Context, orderNo string) (resp *Bet, err error) {
	resp = &Bet{}
	if len(orderNo) == 0 {
		return
	}
	sqlStr, sqlParams, _ := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_api_order_no: orderNo,
	}).ToSql()
	err = m.conn.QueryRowCtx(ctx, resp, sqlStr, sqlParams...)
	return
}

// 根据id，修改数据
func (m *customBetModel) UpdateById(ctx context.Context, id int64, data map[string]interface{}) (err error) {
	if len(data) == 0 || id <= 0 {
		return
	}
	sqlStr, sqlParams, err := sq.Update(m.table).Where(sq.Eq{
		Bet_F_id: id,
	}).SetMap(data).ToSql()
	if err != nil {
		return err
	}
	_, err = m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	return err
}

// 获取多条订单信息
func (m *customBetModel) GetByApiOrderNos(ctx context.Context, orderNos []string) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	if len(orderNos) == 0 {
		return
	}
	sqlStr, sqlParams, err := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_api_order_no: orderNos,
	}).ToSql()
	if err != nil {
		return
	}
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err != nil && (err == sql.ErrNoRows || err == sqlx.ErrNotFound) {
		err = nil
	}
	return
}

// 获取指定期所有未兑现注单
func (m *customBetModel) GetUncashedBetsByTerm(ctx context.Context, term *CrashTerm) (bets []*Bet, err error) {
	bets = make([]*Bet, 0)
	sqlStr, sqlArgs, _ := sq.
		Select(betFieldNames...).
		From(m.table).
		Where("term_id = ?", term.TermId).
		Where("runtime_channel_id = ? OR (runtime_channel_id = 0 AND channel_id = ?)", term.ChannelId, term.ChannelId).
		Where("(auto_cashout_multiple <= ? OR (manual_cashout_multiple > 0 AND manual_cashout_multiple <= ?))",
			term.Multiple, term.Multiple).
		Where(sq.Eq{"order_status": []int64{OrderStatusCreated, OrderStatusCashingOut}}).
		ToSql()

	err = m.conn.QueryRowsCtx(ctx, &bets, sqlStr, sqlArgs...)
	if err != nil && (err == sql.ErrNoRows || err == sqlx.ErrNotFound) {
		err = nil
	}
	return
}

// 获取指定期， 所有退款中的注单
func (m *customBetModel) GetUnrefundBetsByTerm(ctx context.Context, term *CrashTerm) (bets []*Bet, err error) {
	bets = make([]*Bet, 0)
	sqlStr, sqlArgs, _ := sq.Select(betFieldNames...).From(m.table).
		Where(sq.Eq{
			Bet_F_term_id:      term.Id,
			Bet_F_order_status: OrderStatusRefunding,
		}).
		Where("runtime_channel_id = ? OR (runtime_channel_id = 0 AND channel_id = ?)", term.ChannelId, term.ChannelId).
		ToSql()
	err = m.conn.QueryRowsCtx(ctx, &bets, sqlStr, sqlArgs...)
	if err != nil && (err == sql.ErrNoRows || err == sqlx.ErrNotFound) {
		err = nil
	}
	return
}

// 按期获取所有正常注单
func (m *customBetModel) GetNormalByTermID(ctx context.Context, channelID int64, termID int64) (bets []*Bet, err error) {
	bets = make([]*Bet, 0)
	sqlStr, sqlArgs, _ := sq.Select(betFieldNames...).From(m.table).Where("term_id=?", termID).
		Where("runtime_channel_id = ? OR (runtime_channel_id = 0 AND channel_id = ?)", channelID, channelID).
		Where(sq.Eq{"order_status": []int64{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout}}).ToSql()
	bets = make([]*Bet, 0)
	err = m.conn.QueryRowsCtx(ctx, &bets, sqlStr, sqlArgs...)
	if err != nil && (err == sql.ErrNoRows || err == sqlx.ErrNotFound) {
		err = nil
	}
	return
}

func (m *customBetModel) GetByRuntimeChannelAndTerm(ctx context.Context, runtimeChannelID int64, termID int64) (bets []*Bet, err error) {
	bets = make([]*Bet, 0)
	sqlStr, sqlArgs, _ := sq.Select(betFieldNames...).From(m.table).
		Where(sq.Eq{Bet_F_term_id: termID}).
		Where("runtime_channel_id = ? OR (runtime_channel_id = 0 AND channel_id = ?)", runtimeChannelID, runtimeChannelID).
		Where(sq.Eq{"order_status": []int64{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout}}).
		ToSql()
	err = m.conn.QueryRowsCtx(ctx, &bets, sqlStr, sqlArgs...)
	if err != nil && (err == sql.ErrNoRows || err == sqlx.ErrNotFound) {
		err = nil
	}
	return
}

// 根据id获取一条数据
func (m *customBetModel) GetById(ctx context.Context, id int64) (resp *Bet, err error) {
	resp = &Bet{}
	if id <= 0 {
		return
	}
	sqlStr, sqlParams, _ := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_id: id,
	}).ToSql()
	m.conn.QueryRowCtx(ctx, resp, sqlStr, sqlParams...)
	return
}

// 批量修改
func (m *customBetModel) UpdateByIds(ctx context.Context, ids []int64, data map[string]interface{}) error {
	if len(ids) == 0 || len(data) == 0 {
		return nil
	}
	sqlStr, sqlParams, err := sq.Update(m.table).Where(sq.Eq{
		Bet_F_id: ids,
	}).SetMap(data).ToSql()
	if err != nil {
		return err
	}
	_, err = m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	return err

}

// 根据ids， 获取多条数据
func (m *customBetModel) GetByIds(ctx context.Context, ids []int64) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	if len(ids) == 0 {
		return
	}
	sqlStr, sqlParams, err := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_id: ids,
	}).ToSql()
	if err != nil {
		return
	}
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

func (m *customBetModel) GetBetsTodayBest(ctx context.Context, channelId, userId int64, startTime, endTime time.Time) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)

	conj := sq.And{
		sq.Gt{Bet_F_cashed_out_amount: 0},
		sq.GtOrEq{Bet_F_ctime: startTime.Unix()},
		sq.Lt{Bet_F_ctime: endTime.Unix()},
		sq.Eq{
			Bet_F_user_id:      userId,
			Bet_F_order_status: []int{OrderStatusCashingOut, OrderStatusCashedout},
		},
	}

	if channelId > 0 {
		conj = append(conj, sq.Eq{Bet_F_channel_id: channelId})
	}

	sqlStr, sqlParams, err := sq.Select(betFieldNames...).
		From(m.table).
		Where(conj).
		ToSql()
	if err != nil {
		return
	}
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 通过币种、用户，拉取对应已经创建以及兑现的订单
func (m *customBetModel) GetByUserCurrencyCreateTime(ctx context.Context, currency string, userId int64, startTime time.Time) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	sqlStr, sqlParams, err := sq.Select(betFieldNames...).From(m.table).Where(sq.Gt{
		Bet_F_ctime: startTime.Unix(),
	}).Where(sq.Eq{
		Bet_F_user_id:      userId,
		Bet_F_currency:     currency,
		Bet_F_order_status: []int{OrderStatusCreated, OrderStatusCashedout},
	}).ToSql()
	if err != nil {
		return
	}
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

func (m *customBetModel) GetBetsByTime(ctx context.Context, currency string, channelId, userId int64, startTime, endTime time.Time) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)

	conj := sq.And{
		sq.Gt{Bet_F_ctime: startTime.Unix()},
		sq.Lt{Bet_F_ctime: endTime.Unix()},
		sq.Eq{
			Bet_F_user_id:      userId,
			Bet_F_currency:     currency,
			Bet_F_order_status: []int{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
		},
	}

	if channelId > 0 {
		conj = append(conj, sq.Eq{Bet_F_channel_id: channelId})
	}

	sqlStr, sqlParams, err := sq.Select(betFieldNames...).
		From(m.table).
		Where(conj).
		OrderBy(Bet_F_ctime + " desc").
		ToSql()

	if err != nil {
		return
	}
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 通过币种、用户，拉取用户下注订单，分页
func (m *customBetModel) GetBetsByTimePage(ctx context.Context, currency string, userId int64, page, pageSize uint64, startTime, endTime time.Time) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	offset := (page - 1) * pageSize
	sqlStr, sqlParams, err := sq.Select(betFieldNames...).From(m.table).Where(sq.Gt{
		Bet_F_ctime: startTime.Unix(),
	}).Where(sq.Lt{
		Bet_F_ctime: endTime.Unix(),
	}).Where(sq.Eq{
		Bet_F_user_id:      userId,
		Bet_F_currency:     currency,
		Bet_F_order_status: []int{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
	}).OrderBy(Bet_F_ctime + " desc").
		Limit(pageSize).
		Offset(offset).ToSql()

	if err != nil {
		return
	}
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 通过币种、用户，拉取用户下注订单，总数
func (m *customBetModel) GetBetsByTimeTotal(ctx context.Context, currency string, userId int64, startTime, endTime time.Time) (resp int64, err error) {
	sqb := sq.Select("count(*)").From(m.table).Where(sq.Gt{
		Bet_F_ctime: startTime.Unix(),
	}).Where(sq.Lt{
		Bet_F_ctime: endTime.Unix(),
	}).Where(sq.Eq{
		Bet_F_user_id:      userId,
		Bet_F_currency:     currency,
		Bet_F_order_status: []int{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
	})
	sqlStr, sqlParams, err := sqb.ToSql()
	_ = m.conn.QueryRowCtx(ctx, &resp, sqlStr, sqlParams...)

	if err != nil {
		return
	}
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 获取用户指定时间开始的一页注单数据
func (m *customBetModel) GetUserBetsById(ctx context.Context, userId, channelId int64, posId int64, startTime, endTime time.Time, pageSize int, currency string) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	if userId <= 0 || posId < 0 || pageSize < 0 {
		return
	}

	sqb := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_channel_id:   channelId,
		Bet_F_user_id:      userId,
		Bet_F_order_status: []int64{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
	})

	if currency != "" {
		sqb = sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
			Bet_F_channel_id:   channelId,
			Bet_F_user_id:      userId,
			Bet_F_order_status: []int64{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
			Bet_F_currency:     currency,
		})
	}

	if posId > 0 {
		sqb = sqb.Where(sq.Lt{
			Bet_F_id: posId,
		})
	}
	if startTime.Unix() > 0 {
		sqb = sqb.Where(sq.GtOrEq{
			Bet_F_ctime: startTime.Unix(),
		})
	}
	if endTime.Unix() > 0 {
		sqb = sqb.Where(sq.Lt{
			Bet_F_ctime: endTime.Unix(),
		})
	}

	sqlStr, sqlParams, err := sqb.GroupBy(Bet_F_api_order_no).OrderBy(Bet_F_id + " desc ").Limit(uint64(pageSize)).ToSql()
	if pageSize == 0 {
		sqlStr, sqlParams, err = sqb.GroupBy(Bet_F_api_order_no).OrderBy(Bet_F_id + " desc ").ToSql()
	}
	if err != nil {
		return
	}

	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

func (m *customBetModel) GetUserBets(ctx context.Context, userId, channelId int64, posId int64, startTime, endTime time.Time, pageSize int, currency string) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	if userId <= 0 || posId < 0 || pageSize <= 0 {
		return
	}

	sqb := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_channel_id:   channelId,
		Bet_F_user_id:      userId,
		Bet_F_order_status: []int64{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
	})

	if currency != "" {
		sqb = sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
			Bet_F_channel_id:   channelId,
			Bet_F_user_id:      userId,
			Bet_F_order_status: []int64{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
			Bet_F_currency:     currency,
		})
	}

	if posId > 0 {
		sqb = sqb.Where(sq.Lt{
			Bet_F_id: posId,
		})
	}
	if startTime.Unix() > 0 {
		sqb = sqb.Where(sq.GtOrEq{
			Bet_F_ctime: startTime.Unix(),
		})
	}
	if endTime.Unix() > 0 {
		sqb = sqb.Where(sq.Lt{
			Bet_F_ctime: endTime.Unix(),
		})
	}

	sqlStr, sqlParams, err := sqb.OrderBy(Bet_F_id + " desc ").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return
	}

	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 获取指定局下的一页有效注单数据
func (m *customBetModel) GetBetsByItemid(ctx context.Context, channelID int64, itemId int64, posId int64, pageSize int64) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	if itemId <= 0 || posId < 0 {
		return
	}

	if pageSize == 0 {
		pageSize = 50
	}

	sqb := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_term_id:      itemId,
		Bet_F_channel_id:   channelID,
		Bet_F_order_status: []int64{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
	})
	if posId > 0 {
		sqb = sqb.Where(sq.Lt{
			Bet_F_id: posId,
		})
	}
	sqlStr, sqlParams, err := sqb.OrderBy(Bet_F_id + " desc ").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return
	}

	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 获取指定局下的一页有效注单数据
func (m *customBetModel) GetBetsByItemIDChannelIDs(ctx context.Context, channelID []int64, itemId int64, posId int64, pageSize int64) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	if itemId <= 0 || posId < 0 {
		return
	}

	if pageSize == 0 {
		pageSize = 50
	}

	sqb := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_term_id:      itemId,
		Bet_F_channel_id:   channelID,
		Bet_F_order_status: []int64{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
	})
	if posId > 0 {
		sqb = sqb.Where(sq.Lt{
			Bet_F_id: posId,
		})
	}
	sqlStr, sqlParams, err := sqb.OrderBy(Bet_F_id + " desc ").Limit(uint64(pageSize)).ToSql()
	if err != nil {
		return
	}

	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 获取指定局下的一页有效注单数据
func (m *customBetModel) GetAllBetsByItemId(ctx context.Context, channelID int64, itemId int64) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	sqb := sq.Select(betFieldNames...).From(m.table).
		Where(sq.Eq{
			Bet_F_term_id:      itemId,
			Bet_F_order_status: []int64{OrderStatusCreated, OrderStatusCashedout},
		}).
		Where("runtime_channel_id = ? OR (runtime_channel_id = 0 AND channel_id = ?)", channelID, channelID)
	sqlStr, sqlParams, err := sqb.ToSql()
	if err != nil {
		return
	}

	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

func (m *customBetModel) GetTermCashoutBets(ctx context.Context, channelID []int64, itemId int64) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	if len(channelID) == 0 {
		return
	}
	conj := sq.And{
		sq.Gt{Bet_F_cashed_out_amount: 0},
		sq.Eq{
			Bet_F_order_status: []int{OrderStatusCashingOut, OrderStatusCashedout},
		},
		sq.Eq{
			Bet_F_term_id: itemId,
		},
	}

	sqlStr, sqlParams, err := sq.Select(betFieldNames...).
		From(m.table).
		Where(conj).
		Where(sq.Or{
			sq.Eq{Bet_F_runtime_channel_id: channelID},
			sq.And{
				sq.Eq{Bet_F_runtime_channel_id: int64(0)},
				sq.Eq{Bet_F_channel_id: channelID},
			},
		}).
		ToSql()
	if err != nil {
		return
	}
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 获取当日自动兑现倍数最大的10条注单数据
func (m *customBetModel) GetMaxAutoMultiple(ctx context.Context, channelId int64, startTime time.Time, pageSize int) (resp []*Bet, err error) {
	if channelId <= 0 || startTime.Unix() <= 0 || pageSize <= 0 {
		return
	}

	sqlStr, sqlParams, _ := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_channel_id:              channelId,
		Bet_F_manual_cashout_multiple: 0,
		Bet_F_order_status:            []int64{OrderStatusCashingOut, OrderStatusCashedout},
	}).Where(sq.GtOrEq{
		Bet_F_ctime: startTime.Unix(),
	}).Where(sq.Gt{
		Bet_F_cashed_out_amount: 0,
	}).OrderBy(Bet_F_auto_cashout_multiple + " desc ").Limit(uint64(pageSize)).ToSql()
	resp = make([]*Bet, 0)
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 获取当日手动兑现倍数最大的10条注单数据
func (m *customBetModel) GetMaxManualMultiple(ctx context.Context, channelId int64, startTime time.Time, pageSize int) (resp []*Bet, err error) {
	if channelId <= 0 || startTime.Unix() <= 0 || pageSize <= 0 {
		return
	}

	sqlStr, sqlParams, _ := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
		Bet_F_channel_id:   channelId,
		Bet_F_order_status: []int64{OrderStatusCashingOut, OrderStatusCashedout},
	}).Where(sq.Gt{
		Bet_F_manual_cashout_multiple: 0,
	}).Where(sq.GtOrEq{
		Bet_F_ctime: startTime.Unix(),
	}).OrderBy(Bet_F_manual_cashout_multiple + " desc ").Limit(uint64(pageSize)).ToSql()
	resp = make([]*Bet, 0)
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 返回用户总投注额和总兑现金额
func (m *customBetModel) GetUserBetStatis(ctx context.Context, userId, channelId int64, startTime, endTime time.Time) (resp *UserBetStatis, err error) {
	if userId <= 0 || channelId <= 0 {
		return
	}

	resp = &UserBetStatis{}
	sqb := sq.Select("sum(amount) as totalAmt, sum(cashed_out_amount) as totalCashoutAmt").From(m.table).Where(sq.Eq{
		Bet_F_channel_id:   channelId,
		Bet_F_user_id:      userId,
		Bet_F_order_status: []int64{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
	})

	if startTime.Unix() > 0 {
		sqb = sqb.Where(sq.GtOrEq{
			Bet_F_ctime: startTime.Unix(),
		})
	}
	if endTime.Unix() > 0 {
		sqb = sqb.Where(sq.Lt{
			Bet_F_ctime: endTime.Unix(),
		})
	}

	sqlStr, sqlParams, err := sqb.ToSql()
	if err != nil {
		return
	}
	m.conn.QueryRowCtx(ctx, resp, sqlStr, sqlParams...)
	return
}

// 获取后台需要的一页注单列表
func (m *customBetModel) GetAdminPage(ctx context.Context, args *GetAdminPageArgs, start, pageSize int) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	if start < 0 || pageSize < 0 {
		return
	}
	// 如果有游戏名称过滤
	if len(args.GameName) > 0 {
		// 查询对应游戏名称的渠道IDs
		channelSqb := sq.Select("id").From("channel").Where(sq.Eq{
			Channel_F_game_name: args.GameName,
		})
		if len(args.ChannelId) > 0 {
			// 如果已经有渠道ID限制，需要同时满足两个条件
			channelSqb = channelSqb.Where(sq.Eq{
				Channel_F_id: args.ChannelId,
			})
		}

		// 获取符合游戏名称的渠道IDs
		var gameChannelIds []int64
		channelSqlStr, channelSqlParams, _ := channelSqb.ToSql()
		err = m.conn.QueryRowsCtx(ctx, &gameChannelIds, channelSqlStr, channelSqlParams...)
		if err == sqlx.ErrNotFound {
			err = nil
		}
		if err != nil {
			return
		}

		// 如果没有符合条件的渠道，返回空结果
		if len(gameChannelIds) == 0 {
			return
		}

		// 用找到的渠道IDs替换原有的渠道IDs
		args.ChannelId = gameChannelIds
	}

	sqb := sq.Select(betFieldNames...).From(m.table)
	if len(args.ChannelId) > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_channel_id: args.ChannelId,
		})
	}
	if args.TermId > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_term_id: args.TermId,
		})
	}
	if args.UserId > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_user_id: args.UserId,
		})
	}
	if args.BetType > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_bet_type: args.BetType,
		})
	}
	if args.CashoutStatus == CASHOUT_STATUS_suc {
		sqb = sqb.Where(sq.Eq{
			Bet_F_order_status: []int64{
				OrderStatusCashingOut,
				OrderStatusCashedout,
			},
		})
	} else if args.CashoutStatus == CASHOUT_STATUS_fail {
		sqb = sqb.Where(sq.Eq{
			Bet_F_order_status: OrderStatusCreated,
		})
	} else {
		sqb = sqb.Where(sq.Eq{
			Bet_F_order_status: []int64{
				OrderStatusCashingOut,
				OrderStatusCashedout,
				OrderStatusCreated,
			},
		})
	}

	if args.StartDate > 0 {
		sqb = sqb.Where(sq.GtOrEq{
			Bet_F_ctime: args.StartDate,
		})
	}
	if args.EndDate > 0 {
		sqb = sqb.Where(sq.Lt{
			Bet_F_ctime: args.EndDate,
		})
	}
	// 新增查询条件
	if len(args.OrderId) > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_api_order_no: args.OrderId,
		})
	}
	if len(args.Currency) > 0 && args.Currency != "0" {
		sqb = sqb.Where(sq.Eq{
			Bet_F_currency: args.Currency,
		})
	}
	if args.BetAmtMin > 0 {
		sqb = sqb.Where(sq.GtOrEq{
			"(amount + service_fee)": args.BetAmtMin * 10000, // 转换为存储的单位
		})
	}
	if args.BetAmtMax > 0 {
		sqb = sqb.Where(sq.LtOrEq{
			"(amount + service_fee)": args.BetAmtMax * 10000, // 转换为存储的单位
		})
	}
	// 奖金倍数查询条件
	if args.MultipleMin > 0 || args.MultipleMax > 0 {
		// 计算奖金倍数：兑现时倍数/投注时倍数
		var multipleExpr string
		if args.MultipleMin > 0 && args.MultipleMax > 0 {
			multipleExpr = "((CASE WHEN manual_cashout_multiple > 0 THEN manual_cashout_multiple ELSE auto_cashout_multiple END) / bet_at_multiple) >= ? AND ((CASE WHEN manual_cashout_multiple > 0 THEN manual_cashout_multiple ELSE auto_cashout_multiple END) / bet_at_multiple) <= ?"
			sqb = sqb.Where(multipleExpr, args.MultipleMin*100, args.MultipleMax*100) // 转换为存储的单位
		} else if args.MultipleMin > 0 {
			multipleExpr = "((CASE WHEN manual_cashout_multiple > 0 THEN manual_cashout_multiple ELSE auto_cashout_multiple END) / bet_at_multiple) >= ?"
			sqb = sqb.Where(multipleExpr, args.MultipleMin*100)
		} else if args.MultipleMax > 0 {
			multipleExpr = "((CASE WHEN manual_cashout_multiple > 0 THEN manual_cashout_multiple ELSE auto_cashout_multiple END) / bet_at_multiple) <= ?"
			sqb = sqb.Where(multipleExpr, args.MultipleMax*100)
		}
		// 只查询已兑现的订单，因为未兑现的订单没有奖金倍数
		sqb = sqb.Where(sq.Gt{
			Bet_F_cashed_out_amount: 0,
		})
	}
	sqlStr, sqlParams, _ := sqb.OrderBy(Bet_F_id + " desc").Offset(uint64(start)).Limit(uint64(pageSize)).ToSql()
	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound || err == sql.ErrNoRows {
		err = nil
	}
	return
}

// 计算按“行数”分页的 1-based 起止行号
func pageToRowRange(page, pageSize int64) (startRow, endRow int64) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	startRow = (page-1)*pageSize + 1
	endRow = page * pageSize
	return
}

// 将 GameName 映射为 channel_id 列表（若 args.ChannelId 也传了，则取交集）
func resolveChannelsByGameName(ctx context.Context, conn sqlx.SqlConn, args *GetAdminPageArgs) error {
	if len(args.GameName) == 0 {
		return nil
	}
	sqlCh := "SELECT id FROM channel WHERE game_name = ?"
	params := []any{args.GameName}

	// 若已带 ChannelId，则 AND id IN (...)
	if n := len(args.ChannelId); n > 0 {
		sqlCh += " AND id IN (" + placeholders(n) + ")"
		for _, v := range args.ChannelId {
			params = append(params, v)
		}
	}

	var ids []int64
	if err := conn.QueryRowsCtx(ctx, &ids, sqlCh, params...); err != nil && err != sql.ErrNoRows {
		return err
	}
	// 无匹配则置空（上层据此直接返回空结果）
	args.ChannelId = ids
	return nil
}

// 拼接 WHERE 子句与参数（与分页/总量共享）
func buildWhere(args GetAdminPageArgs) (where string, wargs []any) {
	parts := make([]string, 0, 12)

	// 兑现状态
	switch args.CashoutStatus {
	case 1: // 已兑现（含进行中）
		parts = append(parts, "b.order_status IN (3000,4000)")
	case 2: // 未兑现（仅已创建）
		parts = append(parts, "b.order_status = 2000")
	default:
		parts = append(parts, "b.order_status IN (2000,3000,4000)")
	}

	// 时间
	if args.StartDate > 0 {
		parts = append(parts, "b.ctime >= ?")
		wargs = append(wargs, args.StartDate)
	}
	if args.EndDate > 0 {
		parts = append(parts, "b.ctime < ?")
		wargs = append(wargs, args.EndDate)
	}

	// 维度过滤
	if args.TermId > 0 {
		parts = append(parts, "b.term_id = ?")
		wargs = append(wargs, args.TermId)
	}
	if args.UserId > 0 {
		parts = append(parts, "b.user_id = ?")
		wargs = append(wargs, args.UserId)
	}
	if n := len(args.ChannelId); n > 0 {
		parts = append(parts, "b.channel_id IN ("+placeholders(n)+")")
		for _, v := range args.ChannelId {
			wargs = append(wargs, v)
		}
	}
	if args.BetType > 0 {
		// 1=赛前 2=滚盘 3=奖金（如果“奖金单独查询”是 business 逻辑，且来自 bonus_bet，可在外部处理）
		parts = append(parts, "b.bet_type = ?")
		wargs = append(wargs, args.BetType)
	}
	if len(args.OrderId) > 0 {
		parts = append(parts, "b.api_order_no = ?")
		wargs = append(wargs, args.OrderId)
	}
	if len(args.Currency) > 0 && args.Currency != "0" {
		parts = append(parts, "b.currency = ?")
		wargs = append(wargs, args.Currency)
	}

	// 金额（元 → *10000 存储）
	if args.BetAmtMin > 0 {
		parts = append(parts, "(b.amount + b.service_fee) >= ?")
		wargs = append(wargs, args.BetAmtMin*10000)
	}
	if args.BetAmtMax > 0 {
		parts = append(parts, "(b.amount + b.service_fee) <= ?")
		wargs = append(wargs, args.BetAmtMax*10000)
	}

	// 倍数（float → *100 的整数比较），并限定已兑付才有意义
	if args.MultipleMin > 0 || args.MultipleMax > 0 {
		parts = append(parts, "b.cashed_out_amount > 0")
		expr := "( (CASE WHEN b.manual_cashout_multiple>0 THEN b.manual_cashout_multiple ELSE b.auto_cashout_multiple END) / b.bet_at_multiple )"
		switch {
		case args.MultipleMin > 0 && args.MultipleMax > 0:
			parts = append(parts, expr+" BETWEEN ? AND ?")
			wargs = append(wargs, int64(args.MultipleMin*100.0), int64(args.MultipleMax*100.0))
		case args.MultipleMin > 0:
			parts = append(parts, expr+" >= ?")
			wargs = append(wargs, int64(args.MultipleMin*100.0))
		case args.MultipleMax > 0:
			parts = append(parts, expr+" <= ?")
			wargs = append(wargs, int64(args.MultipleMax*100.0))
		}
	}

	if len(parts) > 0 {
		where = "WHERE " + joinAnd(parts)
	}
	return
}

// --- 小工具 ---
func placeholders(n int) string {
	if n <= 0 {
		return ""
	}
	s := "?"
	for i := 1; i < n; i++ {
		s += ",?"
	}
	return s
}
func joinAnd(ss []string) string {
	out := ss[0]
	for i := 1; i < len(ss); i++ {
		out += " AND " + ss[i]
	}
	return out
}

// Row：展开后的行；bet 一行 +（如有奖）bonus 一行
type Row struct {
	RowID       int64  `db:"row_id"`
	RowType     string `db:"row_type"`      // "bet" | "bonus"
	RowSortTime int64  `db:"row_sort_time"` // 使用 bet.ctime（时间戳，便于排序）
	Ctime       int64  `db:"ctime"`         // 同上，原始时间戳

	BetID           int64  `db:"bet_id"`
	ApiOrderNo      string `db:"api_order_no"`
	UserID          int64  `db:"user_id"`
	ChannelID       int64  `db:"channel_id"`
	TermID          int64  `db:"term_id"`
	Amount          int64  `db:"amount"`
	ServiceFee      int64  `db:"service_fee"`
	CashedOutAmount int64  `db:"cashed_out_amount"`
	OrderStatus     int64  `db:"order_status"`
	GamePlay        int64  `db:"game_play"`

	ManualCashoutMultiple int64  `db:"manual_cashout_multiple"`
	AutoCashoutMultiple   int64  `db:"auto_cashout_multiple"`
	Id                    int64  `db:"id"`
	BetType               int64  `db:"bet_type"`
	Currency              string `db:"currency"`
	BetAtMultiple         int64  `db:"bet_at_multiple"`
	RakeAmt               int64  `db:"rake_amt"`

	// 新增字段
	ManualCashoutTimes         int64 `db:"manual_cashout_times"`          // 手动兑现次数
	FirstCashoutAmount         int64 `db:"first_cashout_amount"`          // 第一次分步兑现金额
	FirstManualCashoutMultiple int64 `db:"first_manual_cashout_multiple"` // 第一次分步兑现倍数
	IsCashoutAmountMerged      int64 `db:"is_cashout_amount_merged"`      // 是否已累加 first_cashout_amount

	// 奖金（可能为 NULL）
	BonusID          sql.NullInt64  `db:"bonus_id"`
	BonusOrderNo     sql.NullString `db:"bonus_order_no"`
	BonusAmount      sql.NullInt64  `db:"bonus_amount"`
	BonusCashoutMult sql.NullInt64  `db:"bonus_cashout_multiple"`
	BonusRank        sql.NullInt64  `db:"bonus_rank"`
	BonusCreateTime  sql.NullString `db:"bonus_create_time"`
}

// 不拆对地按“行数”分页。
// pageIndex 从 1 开始；pageSizeRows 是每页的“行数配额”（有奖=2，无奖=1）。
// 不拆对分页查询：返回 rows + totalRows（展开后的总行数）
func (m *customBetModel) GetAdminBetRows(ctx context.Context, in *GetAdminPageArgs, page, pageSize int64) (rows []*Row, totalRows int64, err error) {
	rows = []*Row{}

	// game_name → channel_id（若传了）
	if len(in.GameName) > 0 {
		if err = resolveChannelsByGameName(ctx, m.conn, in); err != nil {
			return
		}
		// 没有匹配渠道，直接空
		if len(in.ChannelId) == 0 {
			return rows, 0, nil
		}
	}

	// where & params（分页/总量共用）
	where, wargs := buildWhere(*in)

	// 总量（展开后的行数：每单至少 1 行 + 有奖金再 +1）
	{
		sqlTotal := `SELECT COUNT(*) + COUNT(bonus_bet_id) AS total_rows FROM bet b ` + where
		if err = m.conn.QueryRowCtx(ctx, &totalRows, sqlTotal, wargs...); err != nil && err != sql.ErrNoRows {
			return
		}
		if totalRows == 0 {
			return rows, 0, nil
		}
	}

	// 分页（按“行”计；bet 行 + 可选 bonus 行）：换算到 1-based 起止行
	startRow, endRow := pageToRowRange(page, pageSize)

	// 分页主查：瘦列做窗口 → 命中键回表(b.*) → 展开(bet+bonus 两行)
	finalSQL := `
WITH
ordered_keys AS (
  SELECT b.id, b.ctime, b.bonus_bet_id
  FROM bet b
  ` + where + `
  ORDER BY b.ctime DESC, b.id DESC
  LIMIT ?
),
key_blocks AS (
  SELECT
    ok.*,
    CASE WHEN ok.bonus_bet_id IS NULL THEN 0 ELSE 1 END AS has_bonus,
    (1 + CASE WHEN ok.bonus_bet_id IS NULL THEN 0 ELSE 1 END) AS rows_per_order,
    SUM(1 + CASE WHEN ok.bonus_bet_id IS NULL THEN 0 ELSE 1 END)
      OVER (ORDER BY ok.ctime DESC, ok.id DESC) AS running_rows
  FROM ordered_keys ok
),
page_keys AS (
  SELECT id, ctime, bonus_bet_id
  FROM (
    SELECT
      kb.*,
      (running_rows - rows_per_order + 1) AS block_start_row,
      running_rows                        AS block_end_row
    FROM key_blocks kb
  ) z
  WHERE z.block_start_row >= ? AND z.block_end_row <= ?
),
page_bets AS (
  SELECT b.*
  FROM bet b
  JOIN page_keys pk ON pk.id = b.id
)

-- bet 行
SELECT
  pb.id                AS row_id,
  'bet'                AS row_type,
  pb.ctime             AS row_sort_time,
  pb.ctime             AS ctime,

  pb.id                AS bet_id,
  pb.api_order_no      AS api_order_no,
  pb.user_id           AS user_id,
  pb.channel_id        AS channel_id,
  pb.term_id           AS term_id,
  pb.amount,
  pb.service_fee,
  pb.cashed_out_amount,
  pb.order_status,
  pb.game_play,

  pb.manual_cashout_multiple,
  pb.auto_cashout_multiple,
  pb.id                AS id,
  pb.bet_type,
  pb.currency,
  pb.bet_at_multiple,
  pb.rake_amt,

  pb.manual_cashout_times,
  pb.first_cashout_amount,
  pb.first_manual_cashout_multiple,
  pb.is_cashout_amount_merged,

  NULL                 AS bonus_id,
  NULL                 AS bonus_order_no,
  NULL                 AS bonus_amount,
  NULL                 AS bonus_cashout_multiple,
  NULL                 AS bonus_rank,
  NULL                 AS bonus_create_time
FROM page_bets pb

UNION ALL

-- bonus 行
SELECT
  pb.id                AS row_id,
  'bonus'              AS row_type,
  pb.ctime             AS row_sort_time,
  pb.ctime             AS ctime,

  pb.id                AS bet_id,
  pb.api_order_no      AS api_order_no,
  pb.user_id           AS user_id,
  pb.channel_id        AS channel_id,
  pb.term_id           AS term_id,
  pb.amount,
  pb.service_fee,
  pb.cashed_out_amount,
  pb.order_status,
  pb.game_play,

  pb.manual_cashout_multiple,
  pb.auto_cashout_multiple,
  pb.id                AS id,
  pb.bet_type,
  pb.currency,
  pb.bet_at_multiple,
  pb.rake_amt,

  pb.manual_cashout_times,
  pb.first_cashout_amount,
  pb.first_manual_cashout_multiple,
  pb.is_cashout_amount_merged,

  bb.id                AS bonus_id,
  bb.bonus_order_no,
  bb.bonus_amount,
  bb.cashout_multiple  AS bonus_cashout_multiple,
  bb.bonus_rank,
  bb.create_time       AS bonus_create_time
FROM page_bets pb
JOIN bonus_bet bb ON bb.id = pb.bonus_bet_id

ORDER BY
  row_sort_time DESC,
  row_id DESC,
  CASE row_type WHEN 'bet' THEN 0 ELSE 1 END
`
	// 参数顺序：where 的 wargs... → LIMIT endRow → 窗口筛选起止 [startRow, endRow]
	finalArgs := append([]any{}, wargs...)
	finalArgs = append(finalArgs, endRow, startRow, endRow)

	if err = m.conn.QueryRowsCtx(ctx, &rows, finalSQL, finalArgs...); err != nil && err != sql.ErrNoRows {
		return
	}
	return
}

// 可选：仅查询总量（若某些场景只需要数量）
func (m *customBetModel) CountAdminBetRows(ctx context.Context, in *GetAdminPageArgs) (totalRows int64, err error) {
	if len(in.GameName) > 0 {
		if err = resolveChannelsByGameName(ctx, m.conn, in); err != nil {
			return
		}
		if len(in.ChannelId) == 0 {
			return 0, nil
		}
	}
	where, wargs := buildWhere(*in)
	sqlTotal := `SELECT COUNT(*) + COUNT(bonus_bet_id) AS total_rows FROM bet b ` + where
	if err = m.conn.QueryRowCtx(ctx, &totalRows, sqlTotal, wargs...); err == sql.ErrNoRows {
		err = nil
	}
	return
}

// 获取后台需要的注单总数
func (m *customBetModel) GetAdminPageNum(ctx context.Context, args *GetAdminPageArgs) (int, error) {
	// 如果有游戏名称过滤
	if len(args.GameName) > 0 {
		// 查询对应游戏名称的渠道IDs
		channelSqb := sq.Select("id").From("channel").Where(sq.Eq{
			Channel_F_game_name: args.GameName,
		})
		if len(args.ChannelId) > 0 {
			// 如果已经有渠道ID限制，需要同时满足两个条件
			channelSqb = channelSqb.Where(sq.Eq{
				Channel_F_id: args.ChannelId,
			})
		}

		// 获取符合游戏名称的渠道IDs
		var gameChannelIds []int64
		channelSqlStr, channelSqlParams, _ := channelSqb.ToSql()
		err := m.conn.QueryRowsCtx(ctx, &gameChannelIds, channelSqlStr, channelSqlParams...)
		if err == sqlx.ErrNotFound {
			err = nil
		}
		if err != nil {
			return 0, err
		}

		// 如果没有符合条件的渠道，返回0
		if len(gameChannelIds) == 0 {
			return 0, nil
		}

		// 用找到的渠道IDs替换原有的渠道IDs
		args.ChannelId = gameChannelIds
	}

	sqb := sq.Select("count(*) as num").From(m.table)
	if len(args.ChannelId) > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_channel_id: args.ChannelId,
		})
	}
	if args.TermId > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_term_id: args.TermId,
		})
	}
	if args.UserId > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_user_id: args.UserId,
		})
	}
	if args.BetType > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_bet_type: args.BetType,
		})
	}
	if args.CashoutStatus == CASHOUT_STATUS_suc {
		sqb = sqb.Where(sq.Eq{
			Bet_F_order_status: []int64{
				OrderStatusCashingOut,
				OrderStatusCashedout,
			},
		})
	} else if args.CashoutStatus == CASHOUT_STATUS_fail {
		sqb = sqb.Where(sq.Eq{
			Bet_F_order_status: OrderStatusCreated,
		})
	} else {
		sqb = sqb.Where(sq.Eq{
			Bet_F_order_status: []int64{
				OrderStatusCashingOut,
				OrderStatusCashedout,
				OrderStatusCreated,
			},
		})
	}

	if args.StartDate > 0 {
		sqb = sqb.Where(sq.GtOrEq{
			Bet_F_ctime: args.StartDate,
		})
	}
	if args.EndDate > 0 {
		sqb = sqb.Where(sq.Lt{
			Bet_F_ctime: args.EndDate,
		})
	}
	// 新增查询条件
	if len(args.OrderId) > 0 {
		sqb = sqb.Where(sq.Eq{
			Bet_F_api_order_no: args.OrderId,
		})
	}
	if len(args.Currency) > 0 && args.Currency != "0" {
		sqb = sqb.Where(sq.Eq{
			Bet_F_currency: args.Currency,
		})
	}
	if args.BetAmtMin > 0 {
		sqb = sqb.Where(sq.GtOrEq{
			"(amount + service_fee)": args.BetAmtMin * 10000, // 转换为存储的单位
		})
	}
	if args.BetAmtMax > 0 {
		sqb = sqb.Where(sq.LtOrEq{
			"(amount + service_fee)": args.BetAmtMax * 10000, // 转换为存储的单位
		})
	}
	// 奖金倍数查询条件
	if args.MultipleMin > 0 || args.MultipleMax > 0 {
		// 计算奖金倍数：兑现时倍数/投注时倍数
		var multipleExpr string
		if args.MultipleMin > 0 && args.MultipleMax > 0 {
			multipleExpr = "((CASE WHEN manual_cashout_multiple > 0 THEN manual_cashout_multiple ELSE auto_cashout_multiple END) / bet_at_multiple) >= ? AND ((CASE WHEN manual_cashout_multiple > 0 THEN manual_cashout_multiple ELSE auto_cashout_multiple END) / bet_at_multiple) <= ?"
			sqb = sqb.Where(multipleExpr, args.MultipleMin*100, args.MultipleMax*100) // 转换为存储的单位
		} else if args.MultipleMin > 0 {
			multipleExpr = "((CASE WHEN manual_cashout_multiple > 0 THEN manual_cashout_multiple ELSE auto_cashout_multiple END) / bet_at_multiple) >= ?"
			sqb = sqb.Where(multipleExpr, args.MultipleMin*100)
		} else if args.MultipleMax > 0 {
			multipleExpr = "((CASE WHEN manual_cashout_multiple > 0 THEN manual_cashout_multiple ELSE auto_cashout_multiple END) / bet_at_multiple) <= ?"
			sqb = sqb.Where(multipleExpr, args.MultipleMax*100)
		}
		// 只查询已兑现的订单，因为未兑现的订单没有奖金倍数
		sqb = sqb.Where(sq.Gt{
			Bet_F_cashed_out_amount: 0,
		})
	}

	sqlStr, sqlParams, _ := sqb.ToSql()
	resp := 0
	_ = m.conn.QueryRowCtx(ctx, &resp, sqlStr, sqlParams...)

	return resp, nil
}

// 通过币种、用户，按api_order_no分组查询注单，分页
func (m *customBetModel) GetBetsByTimePageGroupByOrderNo(ctx context.Context, channelId int64, currency string, userId int64, page, pageSize uint64, startTime, endTime time.Time) (resp []*Bet, err error) {
	resp = make([]*Bet, 0)
	offset := (page - 1) * pageSize
	sqlStr, sqlParams, err := sq.Select(betRows).GroupBy(Bet_F_api_order_no).GroupBy(Bet_F_api_order_no).From(m.table).Where(sq.Gt{
		Bet_F_ctime: startTime.Unix(),
	}).Where(sq.Lt{
		Bet_F_ctime: endTime.Unix(),
	}).Where(sq.Eq{
		Bet_F_user_id:      userId,
		Bet_F_currency:     currency,
		Bet_F_order_status: []int{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
	}).OrderBy(Bet_F_ctime + " desc").
		Limit(pageSize).
		Offset(offset).ToSql()

	if channelId > 0 {
		sqlStr, sqlParams, err = sq.Select(betRows).GroupBy(Bet_F_api_order_no).GroupBy(Bet_F_api_order_no).From(m.table).Where(sq.Gt{
			Bet_F_ctime: startTime.Unix(),
		}).Where(sq.Lt{
			Bet_F_ctime: endTime.Unix(),
		}).Where(sq.Eq{
			Bet_F_user_id:      userId,
			Bet_F_channel_id:   channelId,
			Bet_F_currency:     currency,
			Bet_F_order_status: []int{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
		}).OrderBy(Bet_F_ctime + " desc").
			Limit(pageSize).
			Offset(offset).ToSql()
	}

	if err != nil {
		return
	}
	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

// 通过币种、用户，按api_order_no分组查询注单总数
func (m *customBetModel) GetBetsByTimeTotalGroupByOrderNo(ctx context.Context, channelId int64, currency string, userId int64, startTime, endTime time.Time) (resp int64, err error) {
	sqlStr, sqlParams, err := sq.Select("COUNT(DISTINCT api_order_no)").From(m.table).Where(sq.Gt{
		Bet_F_ctime: startTime.Unix(),
	}).Where(sq.Lt{
		Bet_F_ctime: endTime.Unix(),
	}).Where(sq.Eq{
		Bet_F_user_id:      userId,
		Bet_F_currency:     currency,
		Bet_F_order_status: []int{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
	}).ToSql()

	if channelId > 0 {
		sqlStr, sqlParams, err = sq.Select("COUNT(DISTINCT api_order_no)").From(m.table).Where(sq.Gt{
			Bet_F_ctime: startTime.Unix(),
		}).Where(sq.Lt{
			Bet_F_ctime: endTime.Unix(),
		}).Where(sq.Eq{
			Bet_F_user_id:      userId,
			Bet_F_channel_id:   channelId,
			Bet_F_currency:     currency,
			Bet_F_order_status: []int{OrderStatusCreated, OrderStatusCashingOut, OrderStatusCashedout},
		}).ToSql()
	}

	if err != nil {
		return
	}
	_ = m.conn.QueryRowCtx(ctx, &resp, sqlStr, sqlParams...)
	return
}

//
//// 根据主订单号查询子订单
//func (m *customBetModel) GetSubBetsByOrderNo(ctx context.Context, orderNo string) (resp []*Bet, err error) {
//	resp = make([]*Bet, 0)
//	sqlStr, sqlParams, err := sq.Select(betFieldNames...).From(m.table).Where(sq.Eq{
//		Bet_F_api_order_no: orderNo,
//	}).OrderBy(Bet_F_id + " asc").ToSql()
//
//	if err != nil {
//		return
//	}
//	m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
//	return
//}

// 联合分页查询结构体
// 用于bet和bonus_bet联合分页
type BetUnion struct {
	Id              int64  `db:"id"`
	ChannelId       int64  `db:"channel_id"`
	UserId          int64  `db:"user_id"`
	TermId          int64  `db:"term_id"`
	ApiOrderNo      string `db:"api_order_no"`
	BetType         int64  `db:"bet_type"`
	Amount          int64  `db:"amount"`
	BetAtMultiple   int64  `db:"bet_at_multiple"`
	RakeAmt         int64  `db:"rake_amt"`
	ServiceFee      int64  `db:"service_fee"`
	CashoutMultiple int64  `db:"cashout_multiple"`
	CashoutStatus   int64  `db:"cashout_status"`
	CashedOutAmount int64  `db:"cashed_out_amount"`
	ProfitAmt       int64  `db:"profit_amt"`
	BetTime         int64  `db:"bet_time"`
}

// 联合分页查询方法
func (m *customBetModel) GetAdminPageUnion(ctx context.Context, args *GetAdminPageArgs, offset, limit int) ([]*BetUnion, int64, error) {
	where := "1=1"
	where1 := ""
	params := make([]interface{}, 0)

	if len(args.ChannelId) > 0 {
		where += " AND channel_id IN (" + strings.TrimRight(strings.Repeat("?,", len(args.ChannelId)), ",") + ")"
		for _, v := range args.ChannelId {
			params = append(params, v)
		}
	}
	if args.TermId > 0 {
		where += " AND term_id=?"
		params = append(params, args.TermId)
	}
	if args.UserId > 0 {
		where += " AND user_id=?"
		params = append(params, args.UserId)
	}
	if args.StartDate > 0 {
		where1 = where + " AND UNIX_TIMESTAMP(create_time)>=?"
		where += " AND ctime>=?"
		params = append(params, args.StartDate)
	}
	if args.EndDate > 0 {
		where += " AND ctime<?"
		where1 += " AND UNIX_TIMESTAMP(create_time)<?"
		params = append(params, args.EndDate)
	}
	if args.BetType > 0 {
		where += " AND bet_type=?"
		where1 += " AND 3=?"
		params = append(params, args.BetType)
	}

	// bet表SQL
	betSql := `SELECT
		id, channel_id, user_id, term_id, api_order_no, bet_type,
		amount, bet_at_multiple, rake_amt, service_fee,
		CASE WHEN manual_cashout_multiple>0 THEN manual_cashout_multiple ELSE auto_cashout_multiple END AS cashout_multiple,
		CASE WHEN cashed_out_amount>0 THEN 1 ELSE 2 END AS cashout_status,
		cashed_out_amount,
		(cashed_out_amount-amount-service_fee) AS profit_amt,
		ctime AS bet_time
	FROM bet
	WHERE ` + where

	// bonus_bet表SQL
	bonusSql := `SELECT
		id, channel_id, user_id, term_id, '' AS api_order_no, 3 AS bet_type,
		0 AS amount, 0 AS bet_at_multiple, 0 AS rake_amt, 0 AS service_fee,
		cashout_multiple, 1 AS cashout_status,
		bonus_amount AS cashed_out_amount,
		bonus_amount AS profit_amt,
		UNIX_TIMESTAMP(create_time) AS bet_time
	FROM bonus_bet
	WHERE ` + where1

	// 合并SQL
	unionSql := `SELECT * FROM ((` + betSql + `) UNION ALL (` + bonusSql + `)) AS t ORDER BY term_id DESC, id DESC LIMIT ?, ?`
	allParams := append(params, params...)
	allParams = append(allParams, offset, limit)

	// 总数SQL
	countSql := `SELECT COUNT(*) FROM ((` + betSql + `) UNION ALL (` + bonusSql + `)) AS t`

	var total int64
	err := m.conn.QueryRowCtx(ctx, &total, countSql, append(params, params...)...)
	if err != nil {
		return nil, 0, err
	}

	var list []*BetUnion
	err = m.conn.QueryRowsCtx(ctx, &list, unionSql, allParams...)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// 根据渠道 + 查询条件分页查询 bet 表
func (m *customBetModel) ListBetsByChannelIDs(ctx context.Context, channelIDs []int64, req *GetListOfWagersReq, page, size int64) (list []*Bet, total int64, err error) {
	list = make([]*Bet, 0)

	if len(channelIDs) == 0 {
		return
	}
	if req == nil {
		req = &GetListOfWagersReq{}
	}

	// 解析 user_ids: "1,2,3" -> []int64
	var userIDs []int64
	if req.UserIds != "" {
		parts := strings.Split(req.UserIds, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			id, parseErr := strconv.ParseInt(p, 10, 64)
			if parseErr == nil {
				userIDs = append(userIDs, id)
			}
		}
	}

	// ========== 组装 WHERE ==========
	var whereBuilder strings.Builder
	var args []interface{}

	whereBuilder.WriteString(" WHERE ")

	// channel_id IN (...)
	whereBuilder.WriteString("channel_id IN (")
	whereBuilder.WriteString(buildPlaceholders(len(channelIDs)))
	whereBuilder.WriteString(")")
	for _, id := range channelIDs {
		args = append(args, id)
	}

	// user_id IN (...)
	if len(userIDs) > 0 {
		whereBuilder.WriteString(" AND user_id IN (")
		whereBuilder.WriteString(buildPlaceholders(len(userIDs)))
		whereBuilder.WriteString(")")
		for _, id := range userIDs {
			args = append(args, id)
		}
	}

	// 币种
	if req.Currency != "" {
		whereBuilder.WriteString(" AND currency = ?")
		args = append(args, req.Currency)
	}

	// ParentWagerNo / WagerNo 都用 api_order_no（Parent 优先，避免 AND 冲突）
	if req.ParentWagerNo != "" {
		whereBuilder.WriteString(" AND api_order_no = ?")
		args = append(args, req.ParentWagerNo)
	} else if req.WagerNo != "" {
		whereBuilder.WriteString(" AND api_order_no = ?")
		args = append(args, req.WagerNo)
	}

	// 子订单状态 -> order_status
	if statuses := mapWagerStatusToOrderStatus(req.Status); len(statuses) > 0 {
		whereBuilder.WriteString(" AND order_status IN (")
		whereBuilder.WriteString(buildPlaceholders(len(statuses)))
		whereBuilder.WriteString(")")
		for _, s := range statuses {
			args = append(args, s)
		}
	}

	// ========== 日期筛选 ==========
	// DateType: "wager_time" -> ctime
	//           "settlement_time" -> update_time
	orderByCol := ""

	if req.DateType == "settlement_time" {
		orderByCol = "update_time"

		if req.FromDate > 0 {
			whereBuilder.WriteString(" AND update_time >= ?")
			args = append(args, time.Unix(req.FromDate, 0).UTC())
		}
		if req.ToDate > 0 {
			whereBuilder.WriteString(" AND update_time <= ?")
			args = append(args, time.Unix(req.ToDate, 0).UTC())
		}
	} else {
		// 默认按下单时间（wager_time）处理 -> ctime
		orderByCol = "ctime"

		if req.FromDate > 0 {
			whereBuilder.WriteString(" AND ctime >= ?")
			args = append(args, req.FromDate)
		}
		if req.ToDate > 0 {
			whereBuilder.WriteString(" AND ctime <= ?")
			args = append(args, req.ToDate)
		}
	}

	// ========== 分页 ==========
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	offset := (page - 1) * size

	// ========== 统计总数 ==========
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", m.table) + whereBuilder.String()

	err = m.conn.QueryRowCtx(ctx, &total, countQuery, args...)
	if err != nil {
		return
	}
	if total == 0 {
		return
	}

	// ========== 查列表 ==========
	// betRows 是 go-zero 自动生成的列名常量
	query := fmt.Sprintf("SELECT %s FROM %s", betRows, m.table) +
		whereBuilder.String() +
		" ORDER BY " + orderByCol + " DESC LIMIT ? OFFSET ?"

	listArgs := make([]interface{}, 0, len(args)+2)
	listArgs = append(listArgs, args...)
	listArgs = append(listArgs, size, offset)

	err = m.conn.QueryRowsCtx(ctx, &list, query, listArgs...)
	if err != nil {
		return
	}

	return
}

// 生成 "?, ?, ?" 这种占位符串
func buildPlaceholders(n int) string {
	if n <= 0 {
		return ""
	}
	if n == 1 {
		return "?"
	}
	var b strings.Builder
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("?")
	}
	return b.String()
}

// 外部子订单状态 -> 内部 order_status
//
// 外部 status：
// -1 创建中
// 7  创建失败
// 13 创建失败（已取消）
// 0  待处理
// 9  已取消
// 12 已部分结算
// 1  未结算（下线中，将被移除）
// 2  已结算
// 6  已重新结算
// 11 已回滚
func mapWagerStatusToOrderStatus(status *int64) []int64 {
	if status == nil {
		return []int64{OrderStatusCreating, OrderStatusCreated, OrderStatusCreationFailed, OrderStatusCashingOut, OrderStatusCashedout, OrderStatusRefunding, OrderStatusRefunded, OrderStatusOverRetry}
	}

	switch *status {
	case -1: // 创建中
		return []int64{OrderStatusCreating}

	case 7, 13: // 创建失败 / 创建失败（已取消）
		return []int64{OrderStatusCreationFailed}

	case 0: // 待处理
		return []int64{OrderStatusCreated}

	case 12: // 已部分结算
		return []int64{OrderStatusCashedout}

	case 1: // 未结算（下线中，将被移除）
		// 视为仍在结算流程中
		return []int64{OrderStatusCreated, OrderStatusCashingOut}

	case 2: // 已结算
		return []int64{OrderStatusCashedout}

	case 6: // 已重新结算
		// 一般也是“已兑现”态，如需区分可扩展枚举
		return []int64{OrderStatusCashedout}

	case 9: // 已取消
		return []int64{OrderStatusRefunded}

	case 11: // 已回滚
		return []int64{OrderStatusRefunded}

	default:
		return []int64{OrderStatusCreating, OrderStatusCreated, OrderStatusCreationFailed, OrderStatusCashingOut, OrderStatusCashedout, OrderStatusRefunding, OrderStatusRefunded, OrderStatusOverRetry}
	}
}

// RowAgg：每个 channel 的聚合结果
type RowAgg struct {
	ChannelID int64  `db:"channel_id"`
	UserID    int64  `db:"user_id"`
	Currency  string `db:"currency"` // 取每组 max_id 对应行的 currency

	AmountSum      int64 `db:"amount_sum"`
	AutoSum        int64 `db:"auto_sum"`
	ManualSum      int64 `db:"manual_sum"`
	FeeSum         int64 `db:"fee_sum"`
	CashSum        int64 `db:"cash_sum"`
	TotalUserCount int64 `db:"total_user_count"` // 每组去重用户数
	TotalBetCount  int64 `db:"total_bet_count"`  // ✅ 新增：每组订单数量（COUNT(*)）

	MaxID int64 `db:"max_id"` // 该组最大 id（排序/调试）
}

// GroupSumsByChannelRows：按 groupBy 分组聚合 + 时间过滤 + 组分页 + currency + 用户去重数 + 每组条数
// - groupBy: "channel_id" | "user_id"
// - dateField: "wager_time" | "create_time"（其它值一律回退到 wager_time）
// - fromTS/toTS: unix 秒；会自动校正 from>to 的情况
// - lastMaxID: keyset 游标，nil/<=0 表示第一页；有值时将替代 offset 方式分页
func (m *customBetModel) GroupSumsByChannelRows(
	ctx context.Context,
	channelIDs []int64,
	userIDs []int64,
	page, size int64,
	groupBy string,
	dateField string,
	fromTS, toTS int64,
	lastMaxID *int64,
) (rows []RowAgg, totalGroups int64, err error) {

	rows = make([]RowAgg, 0, 16)

	// ---- 超时统一控制
	ctx, cancelAll := context.WithTimeout(ctx, 100*time.Second)
	defer cancelAll()

	// ---- 基础参数保护 ----
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	if size > 500 {
		size = 500
	}
	offset := (page - 1) * size

	// ---- 去重 ----
	channelIDs = uniqI64(channelIDs)
	userIDs = uniqI64(userIDs)

	// ---- 分组列选择 ----
	groupCol := Bet_F_channel_id
	switch groupBy {
	case "user_id":
		groupCol = Bet_F_user_id
	}

	// ---- 时间列选择 ----
	//dateCol := Bet_F_create_time
	dateCol := Bet_F_ctime

	if fromTS > 0 && toTS > 0 && fromTS > toTS {
		fromTS, toTS = toTS, fromTS
	}

	// ---- 公共 WHERE 条件 ----
	basePred := sq.And{
		sq.Eq{Bet_F_order_status: []int64{OrderStatusCreated, OrderStatusCashedout}},
		sq.NotEq{Bet_F_channel_id: []int64{ChannelBlockchainCrash, ChannelBlockchainRocket}},
	}
	if len(userIDs) > 0 {
		basePred = append(basePred, sq.Eq{Bet_F_user_id: userIDs})
	}
	if len(channelIDs) > 0 {
		basePred = append(basePred, sq.Eq{Bet_F_channel_id: channelIDs})
	}
	if fromTS > 0 {
		basePred = append(basePred, sq.GtOrEq{dateCol: fromTS})
		//basePred = append(basePred, sq.GtOrEq{dateCol: time.Unix(fromTS, 0).UTC()})
	}
	if toTS > 0 {
		basePred = append(basePred, sq.LtOrEq{dateCol: toTS})
		//basePred = append(basePred, sq.LtOrEq{dateCol: time.Unix(toTS, 0).UTC()})
	}
	if lastMaxID != nil && *lastMaxID > 0 {
		basePred = append(basePred, sq.Lt{Bet_F_id: *lastMaxID})
	}

	// ============= Step 0: Count (可选) =============
	{
		countSQL, countArgs, e := sq.
			Select("COUNT(DISTINCT " + groupCol + ")").
			From(m.table).
			Where(basePred).
			ToSql()
		if e != nil {
			return nil, 0, e
		}
		ctxCount, cancel := context.WithTimeout(ctx, 1*time.Second)
		defer cancel()
		_ = m.conn.QueryRowCtx(ctxCount, &totalGroups, countSQL, countArgs...) // 容忍超时
	}

	// ============= Step 1: 获取本页 group keys + max_id =============
	type keyRow struct {
		GroupID    int64 `db:"group_id"`
		MaxID      int64 `db:"max_id"`
		MerchantId int64 `db:"merchant_id"`
	}
	keySelect := []string{
		groupCol + " AS group_id",
		"MAX(id) AS max_id",
		Bet_F_channel_id + " AS merchant_id",
	}
	keyQB := sq.
		Select(keySelect...).
		From(m.table).
		Where(basePred).
		GroupBy(groupCol).
		OrderBy("MAX(id) DESC").
		Limit(uint64(size))
	if (lastMaxID == nil || *lastMaxID <= 0) && offset > 0 {
		keyQB = keyQB.Offset(uint64(offset))
	}

	keySQL, keyArgs, e := keyQB.ToSql()
	if e != nil {
		return nil, 0, e
	}

	keys := make([]keyRow, 0, size)
	if e = m.conn.QueryRowsCtx(ctx, &keys, keySQL, keyArgs...); e != nil {
		return nil, totalGroups, e
	}
	if len(keys) == 0 {
		return rows, totalGroups, nil
	}

	groupIDs := make([]int64, 0, len(keys))
	maxIDMap := make(map[int64]int64, len(keys))
	for _, kr := range keys {
		groupIDs = append(groupIDs, kr.GroupID)
		maxIDMap[kr.GroupID] = kr.MaxID
	}

	// ============= Step 2 & 3: 并发执行聚合 + 回表取 currency =============
	type aggRow struct {
		GroupID        int64 `db:"group_id"`
		AmountSum      int64 `db:"amount_sum"`
		AutoSum        int64 `db:"auto_sum"`
		ManualSum      int64 `db:"manual_sum"`
		FeeSum         int64 `db:"fee_sum"`
		CashSum        int64 `db:"cash_sum"`
		TotalUserCount int64 `db:"total_user_count"`
		TotalBetCount  int64 `db:"total_bet_count"`
		MerchantId     int64 `db:"merchant_id"`
	}
	aggPred := append(basePred, sq.Eq{groupCol: groupIDs})
	aggSelect := []string{
		Bet_F_channel_id + " AS merchant_id",
		groupCol + " AS group_id",
		"SUM(amount) AS amount_sum",
		"SUM(auto_cashout_multiple) AS auto_sum",
		"SUM(manual_cashout_multiple) AS manual_sum",
		"SUM(service_fee) AS fee_sum",
		"SUM(cashed_out_amount) AS cash_sum",
		"COUNT(DISTINCT " + Bet_F_user_id + ") AS total_user_count",
		"COUNT(*) AS total_bet_count",
	}
	aggQB := sq.Select(aggSelect...).From(m.table).Where(aggPred).GroupBy(groupCol)
	aggSQL, aggArgs, e := aggQB.ToSql()
	if e != nil {
		return nil, totalGroups, e
	}

	maxIDs := make([]int64, 0, len(keys))
	for _, gid := range groupIDs {
		if mid, ok := maxIDMap[gid]; ok {
			maxIDs = append(maxIDs, mid)
		}
	}
	curSQL, curArgs, e2 := sq.
		Select("id", "currency").
		From(m.table).
		Where(sq.Eq{Bet_F_id: maxIDs}).
		ToSql()
	if e2 != nil {
		return nil, totalGroups, e2
	}

	var aggs []aggRow
	type curRow struct {
		ID       int64  `db:"id"`
		Currency string `db:"currency"`
	}
	var curRows []curRow

	// 并发执行
	g, ctx2 := errgroup.WithContext(ctx)
	g.Go(func() error {
		return m.conn.QueryRowsCtx(ctx2, &aggs, aggSQL, aggArgs...)
	})
	g.Go(func() error {
		if len(maxIDs) == 0 {
			return nil
		}
		return m.conn.QueryRowsCtx(ctx2, &curRows, curSQL, curArgs...)
	})
	if err = g.Wait(); err != nil {
		return nil, totalGroups, err
	}

	aggMap := make(map[int64]aggRow, len(aggs))
	for _, ar := range aggs {
		aggMap[ar.GroupID] = ar
	}
	curMap := make(map[int64]string, len(curRows))
	for _, cr := range curRows {
		curMap[cr.ID] = cr.Currency
	}

	// ============= Step 4: 拼装结果 =============
	rows = make([]RowAgg, 0, len(keys))
	for _, kr := range keys {
		ar := aggMap[kr.GroupID]
		cur := curMap[kr.MaxID]
		item := RowAgg{
			AmountSum:      ar.AmountSum,
			AutoSum:        ar.AutoSum,
			ManualSum:      ar.ManualSum,
			FeeSum:         ar.FeeSum,
			CashSum:        ar.CashSum,
			TotalUserCount: ar.TotalUserCount,
			TotalBetCount:  ar.TotalBetCount,
			MaxID:          kr.MaxID,
			Currency:       cur,
		}
		if groupCol == Bet_F_channel_id {
			item.ChannelID = kr.GroupID
		} else {
			item.UserID = kr.GroupID
			item.ChannelID = kr.MerchantId
		}
		rows = append(rows, item)
	}

	return rows, totalGroups, nil
}

// 去重工具
func uniqI64(in []int64) []int64 {
	if len(in) <= 1 {
		return in
	}
	seen := make(map[int64]struct{}, len(in))
	out := make([]int64, 0, len(in))
	for _, v := range in {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func (m *customBetModel) Table() string {
	return m.table
}

func (m *customBetModel) Conn() sqlx.SqlConn {
	return m.conn
}

// 获取 bet 表中已存在的币种列表（去重）
func (m *customBetModel) ListCurrencies(ctx context.Context) (resp []string, err error) {
	resp = make([]string, 0)

	sqlStr, sqlParams, err := sq.
		Select(Bet_F_currency).
		From(m.table).
		GroupBy(Bet_F_currency).
		OrderBy(Bet_F_currency + " ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	if err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...); err != nil {
		return nil, err
	}
	return resp, nil
}
