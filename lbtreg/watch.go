package lbtreg

import (
	"strconv"
	"context"
)

type WatchCallback func(int, string, []byte)

func StartWatchGateCmd(ctx context.Context, cb WatchCallback, host int) {
	etcdKey := GenEtcdKey(strconv.Itoa(host), "cmd", "gate")
	startWatchJob(ctx, cb, etcdKey)
}

func StartWatchServiceCmd(ctx context.Context, cb WatchCallback, host int) {
	etcdKey := GenEtcdKey(strconv.Itoa(host), "cmd", "service")
	startWatchJob(ctx, cb, etcdKey)
}

func startWatchJob(ctx context.Context, cb WatchCallback, etcdKey string) {
	stopped := false
	defer func() {
		if !stopped {
			go startWatchJob(ctx, cb, etcdKey)
		}
	}()
	for {
		rch := etcdClient.Watch(ctx, etcdKey)
		for {
			select {
			case <-ctx.Done():
				stopped = true
				logger.Info("watch job stopped %s", etcdKey)
				return
			case wresp := <-rch:
				for _, ev := range wresp.Events {	// clientv3.EventTypePut(0)/EventTypeDelete(1)
					logger.Debug("etcd watch event %s %q : %q", ev.Type, ev.Kv.Key, ev.Kv.Value)
					cb(int(ev.Type), string(ev.Kv.Key), ev.Kv.Value)
				}
			}
		}
	}
}
