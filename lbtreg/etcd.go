package lbtreg

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

func SplitEtcdKey(key string, n int) []string {
	return strings.SplitN(key, etcdDelimeter, n)
}
