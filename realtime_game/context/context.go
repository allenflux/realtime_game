package context

import (
	rtconfig "crash/realtime_game/config"
	"crash/realtime_game/settlement"
	"crash/realtime_game/store"

	"crash/model/gmmodel"
	"crash/model/servermodel"

	goredis "github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

// AppContext 装配新项目所需的依赖。
type AppContext struct {
	Config rtconfig.Config

	// MySQL 主库
	DB sqlx.SqlConn
	// MySQL GM 库
	GMDB sqlx.SqlConn

	Redis *goredis.Redis

	ChannelModel          servermodel.ChannelModel
	CrashTermModel        servermodel.CrashTermModel
	BetModel              servermodel.BetModel
	RetryCashoutTaskModel servermodel.RetryCashoutTaskModel
	RetryRefundTaskModel  servermodel.RetryRefundTaskModel

	CurrencyLimitModel      gmmodel.CurrencyLimitModel
	GameChannelMappingModel gmmodel.GameChannelMappingModel

	SnapshotStore *store.SnapshotStore
	LeaseStore    *store.LeaseStore

	Settlement settlement.Adapter
}

// New 创建应用上下文。
func New(c rtconfig.Config) *AppContext {
	c.FillDefault()

	db := sqlx.NewMysql(c.Mysql.DataSource)
	gmdb := sqlx.NewMysql(c.Mysql.GmDataSource)
	rds := goredis.MustNewRedis(c.Redis)

	ctx := &AppContext{
		Config: c,

		DB:   db,
		GMDB: gmdb,

		Redis: rds,

		ChannelModel:          servermodel.NewChannelModel(db),
		CrashTermModel:        servermodel.NewCrashTermModel(db),
		BetModel:              servermodel.NewBetModel(db),
		RetryCashoutTaskModel: servermodel.NewRetryCashoutTaskModel(db),
		RetryRefundTaskModel:  servermodel.NewRetryRefundTaskModel(db),

		CurrencyLimitModel:      gmmodel.NewCurrencyLimitModel(gmdb),
		GameChannelMappingModel: gmmodel.NewGameChannelMappingModel(gmdb),
	}

	ctx.SnapshotStore = store.NewSnapshotStore(rds)
	ctx.LeaseStore = store.NewLeaseStore(rds)
	ctx.Settlement = settlement.NewApiSysAdapter(c.ApiSys.Host, c.ApiSys.Token, c.ApiSys.Lang, ctx.GameChannelMappingModel)
	return ctx
}
