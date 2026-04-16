package store

import "fmt"

func keySnapshot(channelID int64) string { return fmt.Sprintf("rt:v3:round:%d:snapshot", channelID) }
func keyLease(channelID int64) string    { return fmt.Sprintf("rt:v3:round:%d:lease", channelID) }
func keyBetHot(channelID int64) string   { return fmt.Sprintf("rt:v3:round:%d:bets", channelID) }
func keyAutoZSet(channelID int64) string { return fmt.Sprintf("rt:v3:round:%d:auto", channelID) }
func keyLeaderboard(channelID int64) string {
	return fmt.Sprintf("rt:v3:round:%d:leaderboard", channelID)
}
func keyBotMarker(termID int64) string       { return fmt.Sprintf("rt:v3:term:%d:bots", termID) }
func keyOpLock(name string) string           { return fmt.Sprintf("rt:v3:lock:%s", name) }
func keyIdempotent(name string) string       { return fmt.Sprintf("rt:v3:idempotent:%s", name) }
func keyWorkerHealth(workerID string) string { return fmt.Sprintf("rt:v3:worker:%s:health", workerID) }
func keyTokenUserInfo(token string) string   { return fmt.Sprintf("rt:v3:token:user:%s", token) }
