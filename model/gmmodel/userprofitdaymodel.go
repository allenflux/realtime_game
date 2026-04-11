package gmmodel

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	UserProfitDay_F_id          = "id"
	UserProfitDay_F_user_id     = "user_id"
	UserProfitDay_F_channel_id  = "channel_id"
	UserProfitDay_F_bet_amt     = "bet_amt"
	UserProfitDay_F_pupm_amt    = "pupm_amt"
	UserProfitDay_F_cashout_amt = "cashout_amt"
	UserProfitDay_F_profit_amt  = "profit_amt"
	UserProfitDay_F_rate        = "rate"
	UserProfitDay_F_record_date = "record_date"
	UserProfitDay_F_create_time = "create_time"
)

var _ UserProfitDayModel = (*customUserProfitDayModel)(nil)

type (
	// UserProfitDayModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserProfitDayModel.
	UserProfitDayModel interface {
		userProfitDayModel
		withSession(session sqlx.Session) UserProfitDayModel
		GetByUidAndRecordDates(ctx context.Context, uid int64, channelId int64, recordDates []int64) ([]*UserProfitDay, error)
	}

	customUserProfitDayModel struct {
		*defaultUserProfitDayModel
	}
)

// NewUserProfitDayModel returns a model for the database table.
func NewUserProfitDayModel(conn sqlx.SqlConn) UserProfitDayModel {
	return &customUserProfitDayModel{
		defaultUserProfitDayModel: newUserProfitDayModel(conn),
	}
}

func (m *customUserProfitDayModel) withSession(session sqlx.Session) UserProfitDayModel {
	return NewUserProfitDayModel(sqlx.NewSqlConnFromSession(session))
}

// 获取指定日期的数据
func (m *customUserProfitDayModel) GetByUidAndRecordDates(ctx context.Context, uid int64, channelId int64, recordDates []int64) (resp []*UserProfitDay, err error) {
	resp = make([]*UserProfitDay, 0)
	if uid <= 0 || len(recordDates) == 0 {
		return
	}
	sqlStr, sqlParams, err := sq.Select(userProfitDayFieldNames...).From(m.table).Where(sq.Eq{
		UserProfitDay_F_user_id:     uid,
		UserProfitDay_F_channel_id:  channelId,
		UserProfitDay_F_record_date: recordDates,
	}).ToSql()
	if err != nil {
		return
	}

	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound || err == sql.ErrNoRows {
		err = nil
	}

	return
}
