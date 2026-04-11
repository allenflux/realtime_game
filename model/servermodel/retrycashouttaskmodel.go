package servermodel

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"strings"
	"time"
)

// 字段常量
const (
	RetryCashoutTask_F_id              = "id"
	RetryCashoutTask_F_bet_id          = "bet_id"
	RetryCashoutTask_F_status          = "status"
	RetryCashoutTask_F_retry_num       = "retry_num"
	RetryCashoutTask_F_next_retry_time = "next_retry_time"
	RetryCashoutTask_F_update_time     = "update_time"
	RetryCashoutTask_F_create_time     = "create_time"
)

// 枚举常量
const (
	// 重试状态 1=待重试 2=重试中 3=已兑现 4=关闭重试 9=需要人工干预
	RetryCashoutTask_Status_need  = 1
	RetryCashoutTask_Status_ing   = 2
	RetryCashoutTask_Status_suc   = 3
	RetryCashoutTask_Status_close = 4
)

var _ RetryCashoutTaskModel = (*customRetryCashoutTaskModel)(nil)

type (
	// RetryCashoutTaskModel is an interface to be customized, add more methods here,
	// and implement the added methods in customRetryCashoutTaskModel.
	RetryCashoutTaskModel interface {
		retryCashoutTaskModel
		withSession(session sqlx.Session) RetryCashoutTaskModel
		GetPageNeedRetry(ctx context.Context, pageSize int64) ([]*RetryCashoutTask, error)
		GetPageNeedOnlineRetry(ctx context.Context, pageSize int64) ([]*RetryCashoutTask, error)
		UpdateByIds(ctx context.Context, ids []int64, data map[string]interface{}) error
		InsertBatch(ctx context.Context, tasks []*RetryCashoutTask) (sql.Result, error)
	}

	customRetryCashoutTaskModel struct {
		*defaultRetryCashoutTaskModel
	}
)

// NewRetryCashoutTaskModel returns a model for the database table.
func NewRetryCashoutTaskModel(conn sqlx.SqlConn) RetryCashoutTaskModel {
	return &customRetryCashoutTaskModel{
		defaultRetryCashoutTaskModel: newRetryCashoutTaskModel(conn),
	}
}

func (m *customRetryCashoutTaskModel) withSession(session sqlx.Session) RetryCashoutTaskModel {
	return NewRetryCashoutTaskModel(sqlx.NewSqlConnFromSession(session))
}

// 获取一页待处理的数据
func (m *customRetryCashoutTaskModel) GetPageNeedRetry(ctx context.Context, pageSize int64) ([]*RetryCashoutTask, error) {
	sqlStr, sqlParams, _ := sq.Select(retryCashoutTaskFieldNames...).From(m.table).
		Where(sq.Eq{
			RetryCashoutTask_F_status: RetryCashoutTask_Status_need,
		}).Where(sq.LtOrEq{
		RetryCashoutTask_F_next_retry_time: time.Now().Unix(),
	}).OrderBy(RetryCashoutTask_F_id + " asc ").Limit(uint64(pageSize)).ToSql()

	resp := make([]*RetryCashoutTask, 0)
	err := m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

// 获取一页CLOSE数据，兜底处理
func (m *customRetryCashoutTaskModel) GetPageNeedOnlineRetry(ctx context.Context, pageSize int64) ([]*RetryCashoutTask, error) {
	sqlStr, sqlParams, _ := sq.Select(retryCashoutTaskFieldNames...).From(m.table).
		Where(sq.Eq{
			RetryCashoutTask_F_status: RetryCashoutTask_Status_close,
		}).OrderBy(RetryCashoutTask_F_id + " desc ").Limit(uint64(pageSize)).ToSql()

	resp := make([]*RetryCashoutTask, 0)
	err := m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}
	return resp, err
}

// 修改数据
func (m *customRetryCashoutTaskModel) UpdateByIds(ctx context.Context, ids []int64, data map[string]interface{}) error {
	if len(ids) == 0 || len(data) == 0 {
		return nil
	}

	sqlStr, sqlParams, _ := sq.Update(m.table).Where(sq.Eq{
		RetryCashoutTask_F_id: ids,
	}).SetMap(data).ToSql()
	_, err := m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	if err != nil {
		return err
	}

	return nil
}

// 批量插入
func (m *customRetryCashoutTaskModel) InsertBatch(ctx context.Context, tasks []*RetryCashoutTask) (sql.Result, error) {
	if len(tasks) == 0 {
		return nil, nil
	}

	// 构造字段名字符串
	columns := retryCashoutTaskRowsExpectAutoSet
	columnCount := len(strings.Split(columns, ","))
	valuePlaceholders := "(" + strings.TrimRight(strings.Repeat("?,", columnCount), ",") + ")"

	// 构造多个占位符与参数
	var placeholders []string
	var args []interface{}

	for _, task := range tasks {
		placeholders = append(placeholders, valuePlaceholders)
		args = append(args, task.BetId, task.Status, task.RetryNum, task.NextRetryTime)
	}

	// 构造最终 SQL
	query := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s", m.table, columns, strings.Join(placeholders, ","))

	// 执行 SQL
	return m.conn.ExecCtx(ctx, query, args...)
}
