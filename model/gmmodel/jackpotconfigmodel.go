package gmmodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ JackpotConfigModel = (*customJackpotConfigModel)(nil)

type (
	// JackpotConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customJackpotConfigModel.
	JackpotConfigModel interface {
		jackpotConfigModel
		withSession(session sqlx.Session) JackpotConfigModel
	}

	customJackpotConfigModel struct {
		*defaultJackpotConfigModel
	}
)

// NewJackpotConfigModel returns a model for the database table.
func NewJackpotConfigModel(conn sqlx.SqlConn) JackpotConfigModel {
	return &customJackpotConfigModel{
		defaultJackpotConfigModel: newJackpotConfigModel(conn),
	}
}

func (m *customJackpotConfigModel) withSession(session sqlx.Session) JackpotConfigModel {
	return NewJackpotConfigModel(sqlx.NewSqlConnFromSession(session))
}
