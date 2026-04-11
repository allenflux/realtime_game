package servermodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ GameBlewLogModel = (*customGameBlewLogModel)(nil)

const (
	GameBlewLog_F_id          = "id"
	GameBlewLog_F_client_id   = "client_id"
	GameBlewLog_F_term_id     = "term_id"
	GameBlewLog_F_user_id     = "user_id"
	GameBlewLog_F_ctrl_result = "ctrl_result"
	GameBlewLog_F_fail_msg    = "fail_msg"
	GameBlewLog_F_create_time = "create_time"
)

type (
	// GameBlewLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGameBlewLogModel.
	GameBlewLogModel interface {
		gameBlewLogModel
		withSession(session sqlx.Session) GameBlewLogModel
	}

	customGameBlewLogModel struct {
		*defaultGameBlewLogModel
	}
)

// NewGameBlewLogModel returns a model for the database table.
func NewGameBlewLogModel(conn sqlx.SqlConn) GameBlewLogModel {
	return &customGameBlewLogModel{
		defaultGameBlewLogModel: newGameBlewLogModel(conn),
	}
}

func (m *customGameBlewLogModel) withSession(session sqlx.Session) GameBlewLogModel {
	return NewGameBlewLogModel(sqlx.NewSqlConnFromSession(session))
}
