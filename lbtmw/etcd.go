package lbtmw

import (
	"time"
	"context"
	"strings"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	etcdDelimeter = "/"
)

var etcdClient *clientv3.Client = nil
var etcdContext = context.Background()

func InitEtcdClient(endpoints []string) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	etcdClient = client
	logger.Info("init etcd %v", endpoints)
}

func CloseEtcdClient() error {
	if err := etcdClient.Close(); err != nil {
		return err
	}
	return nil
}

func GenEtcdKey(dirs ...string) string {
	return strings.Join(dirs, etcdDelimeter)
}

func EtcdPut(k, v string) bool {
	ctx, cancel := context.WithTimeout(etcdContext, 3 * time.Second)
	kvc := clientv3.NewKV(etcdClient)
	_, err := kvc.Put(ctx, k, v)
	cancel()
	if err != nil {
		switch err {
		case context.Canceled:
			logger.Info("etcd put fail 1")
		case context.DeadlineExceeded:
			logger.Info("etcd put fail 2")
		default:
			logger.Info("etcd put fail 3")
		}
		return false
	}
	return true
}

func EtcdGet(k string) string {
	ctx, cancel := context.WithTimeout(etcdContext, 3 * time.Second)
	kvc := clientv3.NewKV(etcdClient)
	resp, err := kvc.Get(ctx, k)
	cancel()
	if err != nil {
		return ""
	}
	kvs := resp.Kvs
	return string(kvs[0].Value)
}
