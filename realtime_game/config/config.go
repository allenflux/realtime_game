package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Config 是实时服务的统一配置。
type Config struct {
	Name string `json:",optional" yaml:",optional"`

	API struct {
		Host string `json:",optional" yaml:",optional"`
		Port int    `json:",optional" yaml:",optional"`
	} `json:",optional" yaml:",optional"`

	Mysql struct {
		DataSource   string `json:",optional" yaml:",optional"`
		GmDataSource string `json:",optional" yaml:",optional"`
	} `json:",optional" yaml:",optional"`

	Redis redis.RedisConf `json:",optional" yaml:",optional"`

	ApiSys struct {
		Host  string `json:",optional" yaml:",optional"`
		Token string `json:",optional" yaml:",optional"`
		Lang  string `json:",optional" yaml:",optional"`
	} `json:",optional" yaml:",optional"`

	Runtime struct {
		PreStartMs   int64 `json:",optional" yaml:",optional"`
		StartingMs   int64 `json:",optional" yaml:",optional"`
		CloseDelayMs int64 `json:",optional" yaml:",optional"`

		TickMs                 int64 `json:",optional" yaml:",optional"`
		LeaseTTLSeconds        int64 `json:",optional" yaml:",optional"`
		LeaseRenewSeconds      int64 `json:",optional" yaml:",optional"`
		RetryIntervalSeconds   int64 `json:",optional" yaml:",optional"`
		RetryPageSize          int64 `json:",optional" yaml:",optional"`
		AutoCashoutBatchSize   int64 `json:",optional" yaml:",optional"`
		FinalizeSettlementPage int64 `json:",optional" yaml:",optional"`
	} `json:",optional" yaml:",optional"`

	Features struct {
		Bonus struct {
			Enabled bool      `json:",optional" yaml:",optional"`
			Ratios  []float64 `json:",optional" yaml:",optional"`
		} `json:",optional" yaml:",optional"`

		Jackpot struct {
			Enabled bool `json:",optional" yaml:",optional"`
		} `json:",optional" yaml:",optional"`

		Leaderboard struct {
			Enabled bool  `json:",optional" yaml:",optional"`
			Size    int64 `json:",optional" yaml:",optional"`
		} `json:",optional" yaml:",optional"`

		Robot struct {
			Enabled             bool    `json:",optional" yaml:",optional"`
			CountPerRound       int64   `json:",optional" yaml:",optional"`
			MinAmount           float64 `json:",optional" yaml:",optional"`
			MaxAmount           float64 `json:",optional" yaml:",optional"`
			MinAutoCashMultiple float64 `json:",optional" yaml:",optional"`
			MaxAutoCashMultiple float64 `json:",optional" yaml:",optional"`
		} `json:",optional" yaml:",optional"`

		Risk struct {
			Enabled            bool  `json:",optional" yaml:",optional"`
			LossThreshold      int64 `json:",optional" yaml:",optional"`
			ForceControlWindow int64 `json:",optional" yaml:",optional"`
		} `json:",optional" yaml:",optional"`
	} `json:",optional" yaml:",optional"`
}

// FillDefault 为缺省值补默认配置。
func (c *Config) FillDefault() {
	if c.API.Host == "" {
		c.API.Host = "0.0.0.0"
	}
	if c.API.Port == 0 {
		c.API.Port = 18080
	}
	if c.ApiSys.Lang == "" {
		c.ApiSys.Lang = "en"
	}
	if c.Runtime.PreStartMs <= 0 {
		c.Runtime.PreStartMs = 8000
	}
	if c.Runtime.StartingMs <= 0 {
		c.Runtime.StartingMs = 2000
	}
	if c.Runtime.CloseDelayMs <= 0 {
		c.Runtime.CloseDelayMs = 5000
	}
	if c.Runtime.TickMs <= 0 {
		c.Runtime.TickMs = 100
	}
	if c.Runtime.LeaseTTLSeconds <= 0 {
		c.Runtime.LeaseTTLSeconds = 15
	}
	if c.Runtime.LeaseRenewSeconds <= 0 {
		c.Runtime.LeaseRenewSeconds = 5
	}
	if c.Runtime.RetryIntervalSeconds <= 0 {
		c.Runtime.RetryIntervalSeconds = 5
	}
	if c.Runtime.RetryPageSize <= 0 {
		c.Runtime.RetryPageSize = 100
	}
	if c.Runtime.AutoCashoutBatchSize <= 0 {
		c.Runtime.AutoCashoutBatchSize = 200
	}
	if c.Runtime.FinalizeSettlementPage <= 0 {
		c.Runtime.FinalizeSettlementPage = 500
	}
	if len(c.Features.Bonus.Ratios) == 0 {
		c.Features.Bonus.Ratios = []float64{0.10, 0.05, 0.02}
	}
	if c.Features.Leaderboard.Size <= 0 {
		c.Features.Leaderboard.Size = 20
	}
	if c.Features.Robot.CountPerRound <= 0 {
		c.Features.Robot.CountPerRound = 3
	}
	if c.Features.Robot.MinAmount <= 0 {
		c.Features.Robot.MinAmount = 5
	}
	if c.Features.Robot.MaxAmount <= 0 {
		c.Features.Robot.MaxAmount = 50
	}
	if c.Features.Robot.MinAutoCashMultiple <= 0 {
		c.Features.Robot.MinAutoCashMultiple = 1.2
	}
	if c.Features.Robot.MaxAutoCashMultiple <= 0 {
		c.Features.Robot.MaxAutoCashMultiple = 3.5
	}
	if c.Features.Risk.LossThreshold <= 0 {
		c.Features.Risk.LossThreshold = 500000
	}
	if c.Features.Risk.ForceControlWindow <= 0 {
		c.Features.Risk.ForceControlWindow = 3
	}
}
