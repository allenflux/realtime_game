package servermodel

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"time"
)

// 字段常量
const (
	RetryRefundTask_F_id              = "id"
	RetryRefundTask_F_bet_id          = "bet_id"
	RetryRefundTask_F_status          = "status"
	RetryRefundTask_F_retry_num       = "retry_num"
	RetryRefundTask_F_next_retry_time = "next_retry_time"
	RetryRefundTask_F_update_time     = "update_time"
	RetryRefundTask_F_create_time     = "create_time"
)

// 枚举值常量
const (
	// 重试状态 1=待重试 2=重试中 3=已兑现 4=需要人工干预
	RetryRefundTask_Status_need  = 1
	RetryRefundTask_Status_ing   = 2
	RetryRefundTask_Status_suc   = 3
	RetryRefundTask_Status_close = 4
)

var _ RetryRefundTaskModel = (*customRetryRefundTaskModel)(nil)

type (
	// RetryRefundTaskModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRetryRefundTaskModel.
	RetryRefundTaskModel interface {
		retryRefundTaskModel
		withSession(session sqlx.Session) RetryRefundTaskModel
		GetPageNeedRetry(ctx context.Context, pageSize int64) ([]*RetryRefundTask, error)
		UpdateByIds(ctx context.Context, ids []int64, data map[string]interface{}) error
	}

	customRetryRefundTaskModel struct {
		*defaultRetryRefundTaskModel
	}
)

// NewRetryRefundTaskModel returns a model for the database table.
func NewRetryRefundTaskModel(conn sqlx.SqlConn) RetryRefundTaskModel {
	return &customRetryRefundTaskModel{
		defaultRetryRefundTaskModel: newRetryRefundTaskModel(conn),
	}
}

func (m *customRetryRefundTaskModel) withSession(session sqlx.Session) RetryRefundTaskModel {
	return NewRetryRefundTaskModel(sqlx.NewSqlConnFromSession(session))
}

// 获取一页待处理的数据
func (m *customRetryRefundTaskModel) GetPageNeedRetry(ctx context.Context, pageSize int64) ([]*RetryRefundTask, error) {
	sqlStr, sqlParams, _ := sq.Select(retryRefundTaskFieldNames...).From(m.table).
		Where(sq.Eq{
			RetryRefundTask_F_status: RetryRefundTask_Status_need,
		}).Where(sq.LtOrEq{
		RetryRefundTask_F_next_retry_time: time.Now().Unix(),
	}).OrderBy(RetryRefundTask_F_id + " asc ").Limit(uint64(pageSize)).ToSql()

	resp := make([]*RetryRefundTask, 0)
	err := m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

// 修改数据
func (m *customRetryRefundTaskModel) UpdateByIds(ctx context.Context, ids []int64, data map[string]interface{}) error {
	if len(ids) == 0 || len(data) == 0 {
		return nil
	}

	sqlStr, sqlParams, _ := sq.Update(m.table).Where(sq.Eq{
		RetryRefundTask_F_id: ids,
	}).SetMap(data).ToSql()
	_, err := m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	if err != nil {
		return err
	}

	return nil
}
