package gmmodel

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	TaskPosition_F_id          = "id"
	TaskPosition_F_task_name   = "task_name"
	TaskPosition_F_position_id = "position_id"
)

var _ TaskPositionModel = (*customTaskPositionModel)(nil)

type (
	// TaskPositionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTaskPositionModel.
	TaskPositionModel interface {
		taskPositionModel
		withSession(session sqlx.Session) TaskPositionModel
		GetByTaskName(taskName string) (TaskPosition, error)
	}

	customTaskPositionModel struct {
		*defaultTaskPositionModel
	}
)

// NewTaskPositionModel returns a model for the database table.
func NewTaskPositionModel(conn sqlx.SqlConn) TaskPositionModel {
	return &customTaskPositionModel{
		defaultTaskPositionModel: newTaskPositionModel(conn),
	}
}

func (m *customTaskPositionModel) withSession(session sqlx.Session) TaskPositionModel {
	return NewTaskPositionModel(sqlx.NewSqlConnFromSession(session))
}

func (m *customTaskPositionModel) GetByTaskName(taskName string) (resp TaskPosition, err error) {
	resp = TaskPosition{}
	if len(taskName) == 0 {
		return
	}

	sqlStr, sqlParams, _ := sq.Select(taskPositionFieldNames...).From(m.table).Where(sq.Eq{
		TaskPosition_F_task_name: taskName,
	}).ToSql()
	m.conn.QueryRow(&resp, sqlStr, sqlParams...)
	return
}
