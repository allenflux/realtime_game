package servermodel

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	UserPageConfig_F_id          = "id"
	UserPageConfig_F_user_id     = "user_id"
	UserPageConfig_F_config_json = "config_json"
	UserPageConfig_F_create_time = "create_time"
	UserPageConfig_F_update_time = "update_time"
	UserPageConfig_F_game_id     = "game_id"
)

var _ UserPageConfigModel = (*customUserPageConfigModel)(nil)

type (
	// UserPageConfigModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserPageConfigModel.
	UserPageConfigModel interface {
		userPageConfigModel
		withSession(session sqlx.Session) UserPageConfigModel
		GetByUserId(ctx context.Context, userId int64) ([]*UserPageConfig, error)
		GetByUserIdAndGameId(ctx context.Context, userId int64, gameId int64) ([]*UserPageConfig, error)
		UpdateById(ctx context.Context, id int64, data map[string]interface{}) error
		AddNew(ctx context.Context, data map[string]interface{}) (sql.Result, error)
	}

	customUserPageConfigModel struct {
		*defaultUserPageConfigModel
	}
)

// NewUserPageConfigModel returns a model for the database table.
func NewUserPageConfigModel(conn sqlx.SqlConn) UserPageConfigModel {
	return &customUserPageConfigModel{
		defaultUserPageConfigModel: newUserPageConfigModel(conn),
	}
}

func (m *customUserPageConfigModel) withSession(session sqlx.Session) UserPageConfigModel {
	return NewUserPageConfigModel(sqlx.NewSqlConnFromSession(session))
}

// 根据用户ID获取配置
func (m *customUserPageConfigModel) GetByUserId(ctx context.Context, userId int64) ([]*UserPageConfig, error) {
	if userId <= 0 {
		return nil, errors.New("userId is empty")
	}

	sqlStr, sqlParams, _ := sq.Select(userPageConfigFieldNames...).From(m.table).Where(sq.Eq{
		UserPageConfig_F_user_id: userId,
	}).ToSql()

	var configs []*UserPageConfig
	err := m.conn.QueryRowsCtx(ctx, &configs, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}

	return configs, err
}

// 根据用户ID和游戏ID获取配置
func (m *customUserPageConfigModel) GetByUserIdAndGameId(ctx context.Context, userId int64, gameId int64) ([]*UserPageConfig, error) {
	if userId <= 0 {
		return nil, errors.New("userId is empty")
	}

	if gameId <= 0 {
		return nil, errors.New("gameId is empty")
	}

	sqlStr, sqlParams, _ := sq.Select(userPageConfigFieldNames...).From(m.table).Where(sq.Eq{
		UserPageConfig_F_user_id: userId,
		UserPageConfig_F_game_id: gameId,
	}).ToSql()

	var configs []*UserPageConfig
	err := m.conn.QueryRowsCtx(ctx, &configs, sqlStr, sqlParams...)
	if err == sqlx.ErrNotFound {
		err = nil
	}

	return configs, err
}

// 根据ID更新配置
func (m *customUserPageConfigModel) UpdateById(ctx context.Context, id int64, data map[string]interface{}) error {
	if id == 0 || len(data) == 0 {
		return nil
	}

	sqlStr, sqlParams, err := sq.Update(m.table).Where(sq.Eq{
		UserPageConfig_F_id: id,
	}).SetMap(data).ToSql()

	if err != nil {
		return err
	}

	_, err = m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
	return err
}

// 添加新配置
func (m *customUserPageConfigModel) AddNew(ctx context.Context, data map[string]interface{}) (sql.Result, error) {
	if data == nil || len(data) == 0 {
		return nil, errors.New("data is nil")
	}

	sqlStr, sqlParams, _ := sq.Insert(m.table).SetMap(data).ToSql()
	return m.conn.ExecCtx(ctx, sqlStr, sqlParams...)
}
