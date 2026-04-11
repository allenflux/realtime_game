package store

import (
	"context"
	"crash/realtime_game/domain"
	"encoding/json"
	"errors"
	"strconv"

	goredis "github.com/zeromicro/go-zero/core/stores/redis"
)

// SnapshotStore 管理 Redis 热状态。
type SnapshotStore struct {
	rds *goredis.Redis
}

func NewSnapshotStore(rds *goredis.Redis) *SnapshotStore { return &SnapshotStore{rds: rds} }

// SaveSnapshot 保存当前局快照。
func (s *SnapshotStore) SaveSnapshot(ctx context.Context, snap *domain.RoundSnapshot) error {
	buf, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	return s.rds.SetCtx(ctx, keySnapshot(snap.ChannelID), string(buf))
}

// GetSnapshot 读取当前局快照。
func (s *SnapshotStore) GetSnapshot(ctx context.Context, channelID int64) (*domain.RoundSnapshot, error) {
	raw, err := s.rds.GetCtx(ctx, keySnapshot(channelID))
	if err != nil || raw == "" {
		return nil, nil
	}
	var snap domain.RoundSnapshot
	if err := json.Unmarshal([]byte(raw), &snap); err != nil {
		return nil, err
	}
	return &snap, nil
}

// SaveBetHot 保存订单热状态。
func (s *SnapshotStore) SaveBetHot(ctx context.Context, channelID int64, state *domain.BetHotState) error {
	buf, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return s.rds.HsetCtx(ctx, keyBetHot(channelID), state.OrderNo, string(buf))
}

// GetBetHot 读取订单热状态。
func (s *SnapshotStore) GetBetHot(ctx context.Context, channelID int64, orderNo string) (*domain.BetHotState, error) {
	raw, err := s.rds.HgetCtx(ctx, keyBetHot(channelID), orderNo)
	if err != nil || raw == "" {
		return nil, nil
	}
	var state domain.BetHotState
	if err := json.Unmarshal([]byte(raw), &state); err != nil {
		return nil, err
	}
	return &state, nil
}

// EnqueueAutoCashout 把订单加入自动兑现队列。
func (s *SnapshotStore) EnqueueAutoCashout(ctx context.Context, channelID int64, orderNo string, target int64) error {
	_, err := s.rds.ZaddCtx(ctx, keyAutoZSet(channelID), target, orderNo)
	return err
}

// RemoveAutoCashout 从自动兑现队列移除订单。
func (s *SnapshotStore) RemoveAutoCashout(ctx context.Context, channelID int64, orderNo string) error {
	_, err := s.rds.ZremCtx(ctx, keyAutoZSet(channelID), orderNo)
	return err
}

// ListDueAutoCashouts 读取所有达到目标的自动兑现订单。
func (s *SnapshotStore) ListDueAutoCashouts(ctx context.Context, channelID int64, currentComparable int64, limit int64) ([]string, error) {
	if limit <= 0 {
		limit = 200
	}
	pairs, err := s.rds.ZrangebyscoreWithScoresAndLimitCtx(ctx, keyAutoZSet(channelID), 0, currentComparable, 0, int(limit))
	if err != nil {
		return nil, err
	}
	resp := make([]string, 0, len(pairs))
	for _, p := range pairs {
		resp = append(resp, p.Key)
	}
	return resp, nil
}

// MarkIdempotent 把某个请求标记成已经处理。
func (s *SnapshotStore) MarkIdempotent(ctx context.Context, name string, ttlSec int, value string) error {
	return s.rds.SetexCtx(ctx, keyIdempotent(name), value, ttlSec)
}

// GetIdempotent 读取幂等记录。
func (s *SnapshotStore) GetIdempotent(ctx context.Context, name string) (string, error) {
	return s.rds.GetCtx(ctx, keyIdempotent(name))
}

// AcquireOpLock 获取操作锁。
func (s *SnapshotStore) AcquireOpLock(ctx context.Context, name string, ttlSec int) (bool, error) {
	return s.rds.SetnxExCtx(ctx, keyOpLock(name), "1", ttlSec)
}

// IncrSnapshotVersion 在 Redis 中提升局版本号。
func (s *SnapshotStore) IncrSnapshotVersion(ctx context.Context, channelID int64) error {
	snap, err := s.GetSnapshot(ctx, channelID)
	if err != nil {
		return err
	}
	if snap == nil {
		return errors.New("round snapshot not found")
	}
	snap.Version++
	return s.SaveSnapshot(ctx, snap)
}

// SaveWorkerHealth 保存 worker 心跳，仅用于排查。
func (s *SnapshotStore) SaveWorkerHealth(ctx context.Context, workerID string, nowMs int64) error {
	return s.rds.SetexCtx(ctx, keyWorkerHealth(workerID), strconv.FormatInt(nowMs, 10), 30)
}
