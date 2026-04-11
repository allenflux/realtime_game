package servermodel

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GameChannelMappingModel = (*customGameChannelMappingModel)(nil)

type (
	// GameChannelMappingModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGameChannelMappingModel.
	GameChannelMappingModel interface {
		gameChannelMappingModel
		withSession(session sqlx.Session) GameChannelMappingModel
		FindByClientId(ctx context.Context, clientId int64) ([]*GameChannelMapping, error)
	}

	customGameChannelMappingModel struct {
		*defaultGameChannelMappingModel
	}
)

// NewGameChannelMappingModel returns a model for the database table.
func NewGameChannelMappingModel(conn sqlx.SqlConn) GameChannelMappingModel {
	return &customGameChannelMappingModel{
		defaultGameChannelMappingModel: newGameChannelMappingModel(conn),
	}
}

func (m *customGameChannelMappingModel) withSession(session sqlx.Session) GameChannelMappingModel {
	return NewGameChannelMappingModel(sqlx.NewSqlConnFromSession(session))
}

// FindByClientId 根据客户端ID查询游戏渠道映射列表
func (m *customGameChannelMappingModel) FindByClientId(ctx context.Context, clientId int64) ([]*GameChannelMapping, error) {
	query := "select * from " + m.table + " where client_id = ?"
	var resp []*GameChannelMapping
	err := m.conn.QueryRowsCtx(ctx, &resp, query, clientId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
