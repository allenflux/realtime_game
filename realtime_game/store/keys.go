package store

import "fmt"

func keySnapshot(channelID int64) string     { return fmt.Sprintf("rt:v3:round:%d:snapshot", channelID) }
func keyLease(channelID int64) string        { return fmt.Sprintf("rt:v3:round:%d:lease", channelID) }
func keyBetHot(channelID int64) string       { return fmt.Sprintf("rt:v3:round:%d:bets", channelID) }
func keyAutoZSet(channelID int64) string     { return fmt.Sprintf("rt:v3:round:%d:auto", channelID) }
func keyOpLock(name string) string           { return fmt.Sprintf("rt:v3:lock:%s", name) }
func keyIdempotent(name string) string       { return fmt.Sprintf("rt:v3:idempotent:%s", name) }
func keyWorkerHealth(workerID string) string { return fmt.Sprintf("rt:v3:worker:%s:health", workerID) }
