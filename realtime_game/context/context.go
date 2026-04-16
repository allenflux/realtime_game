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
	BonusBetModel         servermodel.BonusBetModel
	RetryCashoutTaskModel servermodel.RetryCashoutTaskModel
	RetryRefundTaskModel  servermodel.RetryRefundTaskModel

	CurrencyLimitModel      gmmodel.CurrencyLimitModel
	GameChannelMappingModel gmmodel.GameChannelMappingModel
	JackpotConfigModel      gmmodel.JackpotConfigModel
	JackpotRecordModel      gmmodel.JackpotRecordModel
	GameStatisticsModel     gmmodel.GameStatisticsModel

	SnapshotStore  *store.SnapshotStore
	LeaseStore     *store.LeaseStore
	TokenUserStore *store.TokenUserStore

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
		BonusBetModel:         servermodel.NewBonusBetModel(db),
		RetryCashoutTaskModel: servermodel.NewRetryCashoutTaskModel(db),
		RetryRefundTaskModel:  servermodel.NewRetryRefundTaskModel(db),

		CurrencyLimitModel:      gmmodel.NewCurrencyLimitModel(gmdb),
		GameChannelMappingModel: gmmodel.NewGameChannelMappingModel(gmdb),
		JackpotConfigModel:      gmmodel.NewJackpotConfigModel(gmdb),
		JackpotRecordModel:      gmmodel.NewJackpotRecordModel(gmdb),
		GameStatisticsModel:     gmmodel.NewGameStatisticsModel(gmdb),
	}

	ctx.SnapshotStore = store.NewSnapshotStore(rds)
	ctx.LeaseStore = store.NewLeaseStore(rds)
	ctx.TokenUserStore = store.NewTokenUserStore(rds)
	ctx.Settlement = settlement.NewApiSysAdapter(c.ApiSys.Host, c.ApiSys.Token, c.ApiSys.Lang, ctx.GameChannelMappingModel)
	return ctx
}
