package lbtreg

import (
	"time"
	"strconv"
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func StartRegisterGate(tickTime int, stopCh chan bool, host int, addr string) {
	etcdKey := GenEtcdKey(strconv.Itoa(host), "gate", addr)
	startRegisterJob(tickTime, stopCh, etcdKey)
}

func StartRegisterService(tickTime int, stopCh chan bool, host int, serviceType string, addr string) {
	etcdKey := GenEtcdKey(strconv.Itoa(host), "service", serviceType, addr)
	startRegisterJob(tickTime, stopCh, etcdKey)
}

func startRegisterJob(tickTime int, stopCh chan bool, etcdKey string) {
	stopped := false
	ticker := time.NewTicker(time.Duration(tickTime) * time.Second)
	defer func() {
		ticker.Stop()
		if !stopped {
			go startRegisterJob(tickTime, stopCh, etcdKey)
		}
	}()
	// create lease
	ctx, cancel := context.WithTimeout(etcdContext, 3 * time.Second)
	resp, err := etcdClient.Grant(ctx, int64(tickTime + 5))
	cancel()
	if err != nil {
		logger.Warn("register job failed: etcd grant %s", etcdKey)
		return
	}
	leaseID	:= resp.ID
	// etcd put
	ctx, cancel = context.WithTimeout(etcdContext, 3 * time.Second)
	kvc := clientv3.NewKV(etcdClient)
	_, err = kvc.Put(ctx, etcdKey, "1", clientv3.WithLease(leaseID))
	cancel()
	if err != nil {
		logger.Warn("register job failed: etcd put %s", etcdKey)
		return
	}
	// keep alive tick
	for {
		select {
		case <- stopCh:
			stopped = true
			logger.Info("register job stopped %s", etcdKey)
			return
		case <- ticker.C:
			ctx, cancel = context.WithTimeout(etcdContext, 3 * time.Second)
			_, err := etcdClient.KeepAliveOnce(ctx, leaseID)
			cancel()
			if err != nil {
				logger.Warn("register job failed: etcd ka %s", etcdKey)
				return
			}
		}
	}
}
