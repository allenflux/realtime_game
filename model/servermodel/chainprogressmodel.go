package servermodel

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ChainProgressModel = (*customChainProgressModel)(nil)

type (
	// ChainProgressModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChainProgressModel.
	ChainProgressModel interface {
		chainProgressModel
		withSession(session sqlx.Session) ChainProgressModel
		GetByChannelID(ctx context.Context, id int64) (resp *ChainProgress, err error)
	}

	customChainProgressModel struct {
		*defaultChainProgressModel
	}
)

// NewChainProgressModel returns a model for the database table.
func NewChainProgressModel(conn sqlx.SqlConn) ChainProgressModel {
	return &customChainProgressModel{
		defaultChainProgressModel: newChainProgressModel(conn),
	}
}

func (m *customChainProgressModel) GetByChannelID(ctx context.Context, id int64) (*ChainProgress, error) {
	if id == 0 {
		return nil, nil
	}

	sqlStr, sqlParams, _ := sq.
		Select(chainProgressFieldNames...).
		From(m.table).
		Where(sq.Eq{"channel_id": id}).
		OrderBy("id DESC").
		Limit(1).
		ToSql()

	resp := new(ChainProgress)
	if err := m.conn.QueryRowCtx(ctx, resp, sqlStr, sqlParams...); err != nil {
		if err == sqlx.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}
	return resp, nil
}

func (m *customChainProgressModel) withSession(session sqlx.Session) ChainProgressModel {
	return NewChainProgressModel(sqlx.NewSqlConnFromSession(session))
}
