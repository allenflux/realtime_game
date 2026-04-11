package servermodel

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ ChannelTermSeedModel = (*customChannelTermSeedModel)(nil)

type (
	// ChannelTermSeedModel is an interface to be customized, add more methods here,
	// and implement the added methods in customChannelTermSeedModel.
	ChannelTermSeedModel interface {
		channelTermSeedModel
		withSession(session sqlx.Session) ChannelTermSeedModel
		GetNum(ctx context.Context) (int64, error)
		GetSeeds(ctx context.Context, channelID int64, termID int64, num int64) (seed []*ChannelTermSeed, err error)
		GetSeedNum(ctx context.Context, channelID int64, termID int64) (int64, error)
		GetSeedMaxTerm(ctx context.Context, channelID int64) (*ChannelTermSeed, error)
		InsertBatch(ctx context.Context, args []*ChannelTermSeed) (sql.Result, error)
		UpdateByIds(ctx context.Context, ids []int64, data map[string]interface{}) error
	}

	customChannelTermSeedModel struct {
		*defaultChannelTermSeedModel
	}
)

// NewChannelTermSeedModel returns a model for the database table.
func NewChannelTermSeedModel(conn sqlx.SqlConn) ChannelTermSeedModel {
	return &customChannelTermSeedModel{
		defaultChannelTermSeedModel: newChannelTermSeedModel(conn),
	}
}

func (m *customChannelTermSeedModel) withSession(session sqlx.Session) ChannelTermSeedModel {
	return NewChannelTermSeedModel(sqlx.NewSqlConnFromSession(session))
}

const (
	SeedOk   = 0
	SeedOver = 1
)

func (m *customChannelTermSeedModel) GetSeedMaxTerm(ctx context.Context, channelID int64) (*ChannelTermSeed, error) {
	sqlStr, sqlArgs, err := sq.Select(channelTermSeedFieldNames...).
		From(m.table).
		Where(sq.Eq{
			"channel_id": channelID,
		}).OrderBy("term_id desc").Limit(1).ToSql()
	if err != nil {
		return nil, err
	}

	var channelTerm ChannelTermSeed
	err = m.conn.QueryRowCtx(ctx, &channelTerm, sqlStr, sqlArgs...)
	if err != nil && (err == sqlx.ErrNotFound || err == sql.ErrNoRows) {
		err = nil
	}
	return &channelTerm, err
}

func (m *customChannelTermSeedModel) GetSeeds(ctx context.Context, channelID int64, termID int64, num int64) ([]*ChannelTermSeed, error) {
	if num <= 0 {
		num = 1000
	}

	query := fmt.Sprintf(
		"select %s from %s where channel_id = ? and term_id >= ? and seed_status = ? order by term_id asc limit %d",
		channelTermSeedRows, m.table, num,
	)

	var resp []*ChannelTermSeed
	err := m.conn.QueryRowsCtx(ctx, &resp, query, channelID, termID, SeedOk)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	return resp, nil
}

func (m *customChannelTermSeedModel) GetSeedNum(ctx context.Context, channelID int64, termID int64) (int64, error) {
	query := fmt.Sprintf(
		"select count(*) from %s where channel_id = ? and term_id > ? and seed_status = ?",
		m.table,
	)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query, channelID, termID, SeedOk)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}

func (m *customChannelTermSeedModel) GetNum(ctx context.Context) (int64, error) {
	query := fmt.Sprintf(
		"select count(*) from %s",
		m.table,
	)

	var count int64
	err := m.conn.QueryRowCtx(ctx, &count, query)
	if err != nil {
		if err == sqlx.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}

	return count, nil
}

func (m *customChannelTermSeedModel) InsertBatch(ctx context.Context, args []*ChannelTermSeed) (sql.Result, error) {
	if len(args) == 0 {
		return nil, nil
	}

	// 显式列，务必与 valueArgs 顺序一致，且不包含 `id`
	const cols = "`channel_id`,`term_id`,`seed_id`,`seed_status`,`ctime`"

	valueStrings := make([]string, 0, len(args))
	valueArgs := make([]interface{}, 0, len(args)*5)

	for _, d := range args {
		valueStrings = append(valueStrings, "(?, ?, ?, ?, ?)")
		valueArgs = append(valueArgs,
			d.ChannelId,
			d.TermId,
			d.SeedId,
			d.SeedStatus,
			d.Ctime,
		)
	}

	query := fmt.Sprintf(
		"INSERT INTO %s (`channel_id`,`term_id`,`seed_id`,`seed_status`,`ctime`) VALUES %s",
		m.table, strings.Join(valueStrings, ","),
	)
	return m.conn.ExecCtx(ctx, query, valueArgs...)
}

// 批量修改
func (m *customChannelTermSeedModel) UpdateByIds(ctx context.Context, ids []int64, data map[string]interface{}) error {
	if len(ids) == 0 || len(data) == 0 {
		return nil
	}
	sqlStr, sqlParams, err := sq.Update(m.table).Where(sq.Eq{
		Bet_F_id: ids,
	}).SetMap(data).ToSql()
	if err != nil {
		return err
	}
	_, err = m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	return err

}
