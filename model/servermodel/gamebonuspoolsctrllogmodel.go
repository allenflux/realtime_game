package servermodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ GameBonusPoolsCtrlLogModel = (*customGameBonusPoolsCtrlLogModel)(nil)

const (
	GameBonusPoolsCtrlLog_F_id            = "id"
	GameBonusPoolsCtrlLog_F_client_id     = "client_id"
	GameBonusPoolsCtrlLog_F_user_id       = "user_id"
	GameBonusPoolsCtrlLog_F_in_pools_amt  = "in_pools_amt"
	GameBonusPoolsCtrlLog_F_out_pools_amt = "out_pools_amt"
	GameBonusPoolsCtrlLog_F_create_time   = "create_time"
)

type (
	// GameBonusPoolsCtrlLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGameBonusPoolsCtrlLogModel.
	GameBonusPoolsCtrlLogModel interface {
		gameBonusPoolsCtrlLogModel
		withSession(session sqlx.Session) GameBonusPoolsCtrlLogModel
	}

	customGameBonusPoolsCtrlLogModel struct {
		*defaultGameBonusPoolsCtrlLogModel
	}
)

// NewGameBonusPoolsCtrlLogModel returns a model for the database table.
func NewGameBonusPoolsCtrlLogModel(conn sqlx.SqlConn) GameBonusPoolsCtrlLogModel {
	return &customGameBonusPoolsCtrlLogModel{
		defaultGameBonusPoolsCtrlLogModel: newGameBonusPoolsCtrlLogModel(conn),
	}
}

func (m *customGameBonusPoolsCtrlLogModel) withSession(session sqlx.Session) GameBonusPoolsCtrlLogModel {
	return NewGameBonusPoolsCtrlLogModel(sqlx.NewSqlConnFromSession(session))
}
