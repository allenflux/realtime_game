package servermodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ GameCtrlLogModel = (*customGameCtrlLogModel)(nil)

const (
	GameCtrlLog_F_id                  = "id"
	GameCtrlLog_F_client_id           = "client_id"
	GameCtrlLog_F_term_id             = "term_id"
	GameCtrlLog_F_user_id             = "user_id"
	GameCtrlLog_F_is_control          = "is_control"
	GameCtrlLog_F_break_payout_rate   = "break_payout_rate"
	GameCtrlLog_F_user_profit_correct = "user_profit_correct"
	GameCtrlLog_F_next_rand_multiple  = "next_rand_multiple"
	GameCtrlLog_F_create_time         = "create_time"
)

type (
	// GameCtrlLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGameCtrlLogModel.
	GameCtrlLogModel interface {
		gameCtrlLogModel
		withSession(session sqlx.Session) GameCtrlLogModel
	}

	customGameCtrlLogModel struct {
		*defaultGameCtrlLogModel
	}
)

// NewGameCtrlLogModel returns a model for the database table.
func NewGameCtrlLogModel(conn sqlx.SqlConn) GameCtrlLogModel {
	return &customGameCtrlLogModel{
		defaultGameCtrlLogModel: newGameCtrlLogModel(conn),
	}
}

func (m *customGameCtrlLogModel) withSession(session sqlx.Session) GameCtrlLogModel {
	return NewGameCtrlLogModel(sqlx.NewSqlConnFromSession(session))
}
