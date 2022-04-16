package lbtreg

import (
	"time"
	"strconv"
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type DiscoverCallback func(map[string][]byte)

func StartDiscoverProxy(tickTime int, stopCh chan bool, cb DiscoverCallback, host int) {
	etcdKey := GenEtcdKey(strconv.Itoa(host), "gate", "")
	startDiscoverJob(tickTime, stopCh, cb, etcdKey)
}

func StartDiscoverService(tickTime int, stopCh chan bool, cb DiscoverCallback, host int) {
	etcdKey := GenEtcdKey(strconv.Itoa(host), "service", "")
	startDiscoverJob(tickTime, stopCh, cb, etcdKey)
}

func startDiscoverJob(tickTime int, stopCh chan bool, cb DiscoverCallback, etcdKey string) {
	stopped := false
	ticker := time.NewTicker(time.Duration(tickTime) * time.Second)
	defer func() {
		ticker.Stop()
		if !stopped {
			go startDiscoverJob(tickTime, stopCh, cb, etcdKey)
		}
	}()
	// keep alive tick
	lenPrefix := len(etcdKey)
	for {
		select {
		case <- stopCh:
			stopped = true
			logger.Info("discover job stopped %s", etcdKey)
			return
		case <- ticker.C:
			// etcd get
			ctx, cancel := context.WithTimeout(etcdContext, 3 * time.Second)
			kvc := clientv3.NewKV(etcdClient)
			resp, err := kvc.Get(ctx, etcdKey, clientv3.WithPrefix())
			cancel()
			if err != nil {
				logger.Warn("discover job failed: etcd get %s", etcdKey)
				return
			}
			m := make(map[string][]byte)
			for _, kv := range resp.Kvs {
				k := string(kv.Key)
				if len(k) > lenPrefix { m[k[lenPrefix:]] = kv.Value }
			}
			cb(m)
		}
	}
}
