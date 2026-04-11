package servermodel

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ GameTermIdModel = (*customGameTermIdModel)(nil)

const (
	GameTermId_F_id           = "id"
	GameTermId_F_channel_id   = "channel_id"
	GameTermId_F_term_id      = "term_id"
	GameTermId_F_game_term_id = "game_term_id"
)

type (
	// GameTermIdModel is an interface to be customized, add more methods here,
	// and implement the added methods in customGameTermIdModel.
	GameTermIdModel interface {
		gameTermIdModel
		withSession(session sqlx.Session) GameTermIdModel
		//GetLastOneByChannelId(ctx context.Context, channelId int64) (resp *GameTermId, err error)
		//GetByTermIds(ctx context.Context, termIds []int64) (resp []*GameTermId, err error)
	}

	customGameTermIdModel struct {
		*defaultGameTermIdModel
	}
)

// NewGameTermIdModel returns a model for the database table.
func NewGameTermIdModel(conn sqlx.SqlConn) GameTermIdModel {
	return &customGameTermIdModel{
		defaultGameTermIdModel: newGameTermIdModel(conn),
	}
}

func (m *customGameTermIdModel) withSession(session sqlx.Session) GameTermIdModel {
	return NewGameTermIdModel(sqlx.NewSqlConnFromSession(session))
}

//// 获取游戏下最大的局id
//func (m *customGameTermIdModel) GetLastOneByChannelId(ctx context.Context, channelId int64) (resp *GameTermId, err error) {
//	if channelId <= 0 {
//		err = errors.New("empty channel id")
//		return
//	}
//	sqlStr, sqlParams, err := sq.Select("*").From(m.table).Where(sq.Eq{
//		GameTermId_F_channel_id: channelId,
//	}).OrderBy(GameTermId_F_id + " DESC").Limit(1).ToSql()
//	if err != nil {
//		return
//	}
//	resp = &GameTermId{}
//	err = m.conn.QueryRowCtx(ctx, resp, sqlStr, sqlParams...)
//	if err == ErrNotFound {
//		err = nil
//	}
//	return
//}
//
//// 获取局id对应的子局id
//func (m *customGameTermIdModel) GetByTermIds(ctx context.Context, termIds []int64) (resp []*GameTermId, err error) {
//	if len(termIds) == 0 {
//		err = errors.New("empty termIds")
//		return
//	}
//	sqlStr, sqlParams, err := sq.Select("*").From(m.table).Where(sq.Eq{
//		GameTermId_F_term_id: termIds,
//	}).ToSql()
//	if err != nil {
//		return
//	}
//	resp = make([]*GameTermId, 0)
//	err = m.conn.QueryRowsCtx(ctx, &resp, sqlStr, sqlParams...)
//	if err == ErrNotFound {
//		err = nil
//	}
//	return
//}
