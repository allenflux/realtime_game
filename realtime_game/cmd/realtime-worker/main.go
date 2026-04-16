package main

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	rtconfig "crash/realtime_game/config"
	appctx "crash/realtime_game/context"
	"crash/realtime_game/service"

	"github.com/google/uuid"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "realtime_game/etc/realtime-worker.yaml", "配置文件")

type runner struct {
	cancel context.CancelFunc
}

var (
	mu      sync.Mutex
	runners = map[int64]*runner{}
)

func runChannelLoop(ctx context.Context, svc *service.Services, channelID int64, workerID string) {
	workerSvc := service.NewWorkerService(svc)
	interval := time.Duration(svc.Ctx.Config.Runtime.TickMs) * time.Millisecond
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	nextTick := time.Now()

	for {
		now := time.Now()
		if now.Before(nextTick) {
			timer := time.NewTimer(time.Until(nextTick))
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		} else {
			lag := now.Sub(nextTick)
			if lag > interval {
				skips := lag / interval
				nextTick = nextTick.Add((skips + 1) * interval)
			} else {
				nextTick = nextTick.Add(interval)
			}
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		ch, err := svc.Ctx.ChannelModel.FindOne(ctx, channelID)
		if err != nil || ch == nil {
			continue
		}

		if err := workerSvc.TickOne(ctx, ch, workerID); err != nil {
			logx.Errorf("tick channel=%d err=%v", channelID, err)
		}
	}
}

func main() {
	flag.Parse()

	var c rtconfig.Config
	conf.MustLoad(*configFile, &c)
	ctx := appctx.New(c)
	svc := service.New(ctx)

	workerID := fmt.Sprintf("%s-%s", c.Name, uuid.NewString())
	tick := time.NewTicker(time.Duration(c.Runtime.TickMs) * time.Millisecond)
	retryTick := time.NewTicker(time.Duration(c.Runtime.RetryIntervalSeconds) * time.Second)
	defer tick.Stop()
	defer retryTick.Stop()

	logx.Infof("realtime-worker start, worker_id=%s", workerID)
	for {
		select {
		case <-tick.C:
			_ = ctx.SnapshotStore.SaveWorkerHealth(context.Background(), workerID, time.Now().UnixMilli())
			//channels, err := ctx.ChannelModel.GetAllActiveGames(context.Background())
			//if err != nil {
			//	logx.Errorf("load channels failed: %v", err)
			//	continue
			//}
			//for _, channel := range channels {
			//	if err := service.NewWorkerService(svc).TickOne(context.Background(), channel, workerID); err != nil {
			//		logx.Errorf("tick channel=%d failed: %v", channel.Id, err)
			//	}
			//}
			//_ = ctx.SnapshotStore.SaveWorkerHealth(context.Background(), workerID, time.Now().UnixMilli())

			channels, err := ctx.ChannelModel.GetAllActiveGames(context.Background())
			if err != nil {
				logx.Errorf("load channels failed: %v", err)
				continue
			}

			active := map[int64]struct{}{}

			for _, ch := range channels {
				if !service.ShouldRunRuntimeChannel(ch) {
					continue
				}
				active[ch.Id] = struct{}{}

				mu.Lock()
				if _, ok := runners[ch.Id]; !ok {
					ctx2, cancel := context.WithCancel(context.Background())
					runners[ch.Id] = &runner{cancel: cancel}

					go runChannelLoop(ctx2, svc, ch.Id, workerID)
					logx.Infof("start channel loop: %d", ch.Id)
				}
				mu.Unlock()
			}

			// 停掉已关闭 channel
			mu.Lock()
			for id, r := range runners {
				if _, ok := active[id]; !ok {
					r.cancel()
					delete(runners, id)
					logx.Infof("stop channel loop: %d", id)
				}
			}
			mu.Unlock()
		case <-retryTick.C:
			retrySvc := service.NewRetryService(svc)
			if err := retrySvc.RunCashoutRetry(context.Background()); err != nil {
				logx.Errorf("cashout retry failed: %v", err)
			}
			if err := retrySvc.RunRefundRetry(context.Background()); err != nil {
				logx.Errorf("refund retry failed: %v", err)
			}
		}
	}
}
