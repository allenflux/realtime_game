package store

import (
	"context"

	goredis "github.com/zeromicro/go-zero/core/stores/redis"
)

// LeaseStore 管理当前局 owner 租约。
type LeaseStore struct {
	rds *goredis.Redis
}

func NewLeaseStore(rds *goredis.Redis) *LeaseStore { return &LeaseStore{rds: rds} }

// Acquire 尝试抢占某个渠道的 owner。
func (s *LeaseStore) Acquire(ctx context.Context, channelID int64, workerID string, ttlSec int) (bool, error) {
	return s.rds.SetnxExCtx(ctx, keyLease(channelID), workerID, ttlSec)
}

// Renew 只有当前 owner 才能续约。
func (s *LeaseStore) Renew(ctx context.Context, channelID int64, workerID string, ttlSec int) (bool, error) {
	val, err := s.rds.GetCtx(ctx, keyLease(channelID))
	if err != nil || val == "" {
		return false, err
	}
	if val != workerID {
		return false, nil
	}
	if err := s.rds.SetexCtx(ctx, keyLease(channelID), workerID, ttlSec); err != nil {
		return false, err
	}
	return true, nil
}

// Owner 返回当前 owner。
func (s *LeaseStore) Owner(ctx context.Context, channelID int64) (string, error) {
	return s.rds.GetCtx(ctx, keyLease(channelID))
}

// Release 释放 owner。
func (s *LeaseStore) Release(ctx context.Context, channelID int64, workerID string) error {
	owner, err := s.Owner(ctx, channelID)
	if err != nil || owner == "" {
		return nil
	}
	if owner == workerID {
		_, _ = s.rds.DelCtx(ctx, keyLease(channelID))
	}
	return nil
}
