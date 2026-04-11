package rediskey

import "fmt"

// 货币限额相关的Redis键定义
const (
	// 货币限额配置键前缀 - 格式: currency_limit:{channelId}:{currency}
	KeyCurrencyLimitPrefix = "currency_limit"

	// 货币限额配置版本键 - 格式: currency_limit_version:{channelId}
	KeyCurrencyLimitVersionPrefix = "currency_limit_version"

	// 货币限额更新通知键 - 格式: currency_limit_notify
	KeyCurrencyLimitNotify = "currency_limit_notify"

	// 下一局生效的货币限额配置键前缀 - 格式: currency_limit_next:{channelId}:{currency}
	KeyCurrencyLimitNextPrefix = "currency_limit_next"

	// 下一局生效的货币限额配置版本键 - 格式: currency_limit_next_version:{channelId}
	KeyCurrencyLimitNextVersionPrefix = "currency_limit_next_version"

	// 下一局生效的货币限额更新通知键 - 格式: currency_limit_next_notify
	KeyCurrencyLimitNextNotify = "currency_limit_next_notify"
)

// GetCurrencyLimitKey 获取货币限额配置的键名
// channelId: 渠道ID
// currency: 货币类型
func GetCurrencyLimitKey(channelId int64, currency string) string {
	return fmt.Sprintf("%s:%d:%s", KeyCurrencyLimitPrefix, channelId, currency)
}

// GetCurrencyLimitVersionKey 获取货币限额配置版本的键名
// channelId: 渠道ID
func GetCurrencyLimitVersionKey(channelId int64) string {
	return fmt.Sprintf("%s:%d", KeyCurrencyLimitVersionPrefix, channelId)
}

// GetCurrencyLimitNotifyKey 获取货币限额更新通知的键名
func GetCurrencyLimitNotifyKey() string {
	return KeyCurrencyLimitNotify
}

// GetCurrencyLimitNextKey 获取下一局生效的货币限额配置的键名
// channelId: 渠道ID
// currency: 货币类型
func GetCurrencyLimitNextKey(channelId int64, currency string) string {
	return fmt.Sprintf("%s:%d:%s", KeyCurrencyLimitNextPrefix, channelId, currency)
}

// GetCurrencyLimitNextVersionKey 获取下一局生效的货币限额配置版本的键名
// channelId: 渠道ID
func GetCurrencyLimitNextVersionKey(channelId int64) string {
	return fmt.Sprintf("%s:%d", KeyCurrencyLimitNextVersionPrefix, channelId)
}

// GetCurrencyLimitNextNotifyKey 获取下一局生效的货币限额更新通知的键名
func GetCurrencyLimitNextNotifyKey() string {
	return KeyCurrencyLimitNextNotify
}

// 货币限额配置的Hash字段名
const (
	FieldMinBet     = "minBet"     // 最小投注额
	FieldMaxBet     = "maxBet"     // 最大投注额
	FieldMaxProfit  = "maxProfit"  // 最大盈利
	FieldPrecision  = "precision"  // 货币精度
	FieldUpdateTime = "updateTime" // 更新时间
)

// 游戏状态相关常量
const (
	// 游戏状态键 - 格式: game_state:{gameId}
	KeyGameStatePrefix = "game_state"

	// 游戏配置版本字段
	FieldConfigVersion = "configVersion"

	// 游戏状态字段
	FieldGameStatus = "status"

	// 游戏状态值
	GameStatusRunning  = "running"  // 游戏进行中
	GameStatusFinished = "finished" // 游戏已结束
	GameStatusWaiting  = "waiting"  // 等待开始
)

// GetGameStateKey 获取游戏状态的键名
// gameId: 游戏ID
func GetGameStateKey(gameId int64) string {
	return fmt.Sprintf("%s:%d", KeyGameStatePrefix, gameId)
}
