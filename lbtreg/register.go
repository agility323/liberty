package lbtreg

import (
	"time"
	"strconv"
	"context"
	"encoding/json"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func GateRegisterKey(host int, addr string) string {
	return GenEtcdKey(strconv.Itoa(host), "gate", addr)
}

func ServiceRegisterKey(host int, typ, addr string) string {
	return GenEtcdKey(strconv.Itoa(host), "service", typ, addr)
}

type RegData interface {
	Marshal() (string, error)
	Unmarshal([]byte) error
}

type BasicRegData struct {
	Version string
}

func (d *BasicRegData) Marshal() (string, error) {
	b, err := json.Marshal(d)
	return string(b), err
}

func (d *BasicRegData) Unmarshal(b []byte) error {
	return json.Unmarshal(b, d)
}

func StartRegisterGate(ctx context.Context, tickTime int, host int, addr string, data RegData) {
	etcdKey := GenEtcdKey(strconv.Itoa(host), "gate", addr)
	startRegJob(ctx, tickTime, etcdKey, data)
}

func StartRegisterService(ctx context.Context, tickTime int, host int, serviceType string, addr string, data RegData) {
	etcdKey := GenEtcdKey(strconv.Itoa(host), "service", serviceType, addr)
	startRegJob(ctx, tickTime, etcdKey, data)
}

func StartRegisterJob(ctx context.Context, tickTime int, etcdKey string, data RegData) {
	startRegJob(ctx, tickTime, etcdKey, data)
}

func startRegJob(ctx context.Context, tickTime int, etcdKey string, data RegData) {
	stopped := false
	ticker := time.NewTicker(time.Duration(tickTime) * time.Second)
	defer func() {
		ticker.Stop()
		if !stopped {
			go startRegJob(ctx, tickTime, etcdKey, data)
		}
	}()
	// create etcd value
	etcdVal, err := data.Marshal()// TODO regenerate (version update)
	if err != nil {
		logger.Warn("register job failed: etcd value marshal %v %v", data, err)
		return
	}
	// create lease
	lctx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	margin := tickTime / 5
	if margin < 1 { margin = 1 }
	resp, err := etcdClient.Grant(lctx, int64(tickTime + margin))
	cancel()
	if err != nil {
		logger.Warn("register job failed: etcd grant %s", etcdKey)
		return
	}
	leaseID	:= resp.ID
	// etcd put
	lctx, cancel = context.WithTimeout(ctx, 5 * time.Second)
	kvc := clientv3.NewKV(etcdClient)
	_, err = kvc.Put(lctx, etcdKey, etcdVal, clientv3.WithLease(leaseID))
	cancel()
	if err != nil {
		logger.Warn("register job failed: etcd put %s", etcdKey)
		return
	}
	// keep alive tick
	for {
		select {
		case <-ctx.Done():
			stopped = true
			lctx, cancel = context.WithTimeout(ctx, 5 * time.Second)
			if _, err = etcdClient.Revoke(lctx, leaseID); err != nil {
				logger.Error("register job revoke fail %s", etcdKey)
			}
			cancel()
			logger.Info("register job stopped %s", etcdKey)
			return
		case <-ticker.C:
			lctx, cancel = context.WithTimeout(ctx, 5 * time.Second)
			_, err := etcdClient.KeepAliveOnce(lctx, leaseID)
			cancel()
			if err != nil {
				logger.Warn("register job failed: etcd ka %s", etcdKey)
				return
			}
		}
	}
}
