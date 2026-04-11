package servermodel

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ SeedFileHashMgrModel = (*customSeedFileHashMgrModel)(nil)

type (
	// SeedFileHashMgrModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSeedFileHashMgrModel.
	SeedFileHashMgrModel interface {
		seedFileHashMgrModel
		withSession(session sqlx.Session) SeedFileHashMgrModel
		InsertBatch(ctx context.Context, args []*SeedFileHashMgr) (sql.Result, error)
		FindByTermRange(ctx context.Context, termStart, termEnd int64) ([]*SeedFileHashMgr, error)
		FindByFileHash(ctx context.Context, fileHash string) (*SeedFileHashMgr, error)
	}

	customSeedFileHashMgrModel struct {
		*defaultSeedFileHashMgrModel
	}
)

// NewSeedFileHashMgrModel returns a model for the database table.
func NewSeedFileHashMgrModel(conn sqlx.SqlConn) SeedFileHashMgrModel {
	return &customSeedFileHashMgrModel{
		defaultSeedFileHashMgrModel: newSeedFileHashMgrModel(conn),
	}
}

func (m *customSeedFileHashMgrModel) withSession(session sqlx.Session) SeedFileHashMgrModel {
	return NewSeedFileHashMgrModel(sqlx.NewSqlConnFromSession(session))
}

// InsertBatch 批量插入种子文件哈希记录
func (m *customSeedFileHashMgrModel) InsertBatch(ctx context.Context, args []*SeedFileHashMgr) (sql.Result, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("args is empty")
	}

	// 构建批量插入SQL
	// 注意：手动指定字段列表，包含 create_time
	var values []string
	var sqlArgs []interface{}

	for _, arg := range args {
		values = append(values, "(?, ?, ?, ?, ?, ?, ?)")
		sqlArgs = append(sqlArgs, arg.ChannelId, arg.TermStart, arg.TermEnd, arg.FileName, arg.FileHash, arg.FileUrl, arg.CreateTime)
	}

	// 手动指定字段列表（不使用 seedFileHashMgrRowsExpectAutoSet，因为需要包含 create_time）
	query := fmt.Sprintf("insert into %s (`channel_id`,`term_start`,`term_end`,`file_name`,`file_hash`,`file_url`,`create_time`) values %s",
		m.table, strings.Join(values, ","))

	return m.conn.ExecCtx(ctx, query, sqlArgs...)
}

// FindByTermRange 根据期数范围查询
func (m *customSeedFileHashMgrModel) FindByTermRange(ctx context.Context, termStart, termEnd int64) ([]*SeedFileHashMgr, error) {
	query := fmt.Sprintf("select %s from %s where `term_start` >= ? and `term_end` <= ? order by `term_start` asc", seedFileHashMgrRows, m.table)
	var resp []*SeedFileHashMgr
	err := m.conn.QueryRowsCtx(ctx, &resp, query, termStart, termEnd)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByFileHash 根据文件哈希查询
func (m *customSeedFileHashMgrModel) FindByFileHash(ctx context.Context, fileHash string) (*SeedFileHashMgr, error) {
	query := fmt.Sprintf("select %s from %s where `file_hash` = ? limit 1", seedFileHashMgrRows, m.table)
	var resp SeedFileHashMgr
	err := m.conn.QueryRowCtx(ctx, &resp, query, fileHash)
	switch err {
	case nil:
		return &resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
