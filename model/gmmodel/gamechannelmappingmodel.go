package gmmodel

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GameChannelMappingModel = (*customGameChannelMappingModel)(nil)

type (
	GameChannelMappingModel interface {
		gameChannelMappingModel
		ExistMapping(ctx context.Context, clientId string, apisysGameId int64) (bool, error)
		FindByClientId(ctx context.Context, clientId string) ([]*GameChannelMapping, error)
	}

	customGameChannelMappingModel struct {
		*defaultGameChannelMappingModel
	}
)

func NewGameChannelMappingModel(conn sqlx.SqlConn) GameChannelMappingModel {
	return &customGameChannelMappingModel{
		defaultGameChannelMappingModel: newGameChannelMappingModel(conn),
	}
}

func (m *customGameChannelMappingModel) ExistMapping(ctx context.Context, clientId string, apisysGameId int64) (bool, error) {
	var count int64
	query := "SELECT COUNT(*) FROM game_channel_mapping WHERE client_id = ? AND apisys_game_id = ?"
	err := m.conn.QueryRowCtx(ctx, &count, query, clientId, apisysGameId)
	if err != nil {
		return false, fmt.Errorf("查询映射关系失败: %v", err)
	}
	return count > 0, nil
}

// FindByClientId 根据客户端ID查询游戏渠道映射列表
func (m *customGameChannelMappingModel) FindByClientId(ctx context.Context, clientId string) ([]*GameChannelMapping, error) {
	query := fmt.Sprintf("select %s from %s where client_id = ?", gameChannelMappingRows, m.table)
	var resp []*GameChannelMapping
	err := m.conn.QueryRowsCtx(ctx, &resp, query, clientId)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
