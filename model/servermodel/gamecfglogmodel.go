package servermodel

import "github.com/zeromicro/go-zero/core/stores/sqlx"

var _ GameCfgLogModel = (*customGameCfgLogModel)(nil)

const (
	GameCfgLog_F_id                = "id"
	GameCfgLog_F_client_id         = "client_id"
	GameCfgLog_F_user_id           = "user_id"
	GameCfgLog_F_rake              = "rake"
	GameCfgLog_F_ctrl_trigger_rate = "ctrl_trigger_rate"
	GameCfgLog_F_ctrl_put_rate     = "ctrl_put_rate"
	GameCfgLog_F_create_time       = "create_time"
)

type (
	// GameCfgLogModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGameCfgLogModel.
	GameCfgLogModel interface {
		gameCfgLogModel
		withSession(session sqlx.Session) GameCfgLogModel
	}

	customGameCfgLogModel struct {
		*defaultGameCfgLogModel
	}
)

// NewGameCfgLogModel returns a model for the database table.
func NewGameCfgLogModel(conn sqlx.SqlConn) GameCfgLogModel {
	return &customGameCfgLogModel{
		defaultGameCfgLogModel: newGameCfgLogModel(conn),
	}
}

func (m *customGameCfgLogModel) withSession(session sqlx.Session) GameCfgLogModel {
	return NewGameCfgLogModel(sqlx.NewSqlConnFromSession(session))
}
