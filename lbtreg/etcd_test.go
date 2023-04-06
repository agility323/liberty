package lbtreg

import (
	"testing"
	"fmt"
	"context"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestPrintEtcd(t *testing.T){
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:	 []string{"10.1.71.78:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return
	}
	fmt.Println("connect to etcd success")
	defer cli.Close()

	resp, err := cli.Grant(context.Background(), 999)
	if err != nil {
		fmt.Printf("create lease fail %v", err)
		return
	}

	key := "abc"
	kvc := clientv3.NewKV(cli)
	_, err = kvc.Put(context.Background(), "abc", "123", clientv3.WithLease(resp.ID))
	printKey(cli, key)

	_, err = cli.Revoke(context.Background(), resp.ID)
	printKey(cli, key)
}

func printAll(cli *clientv3.Client) {
	resp, err := cli.Get(context.Background(), "", clientv3.WithPrefix())
	//fmt.Printf("resp %v\n", resp)
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s:%s\n", string(ev.Key), string(ev.Value))
		//fmt.Printf("%v:%v\n", (ev.Key), (ev.Value))
	}
}

func printKey(cli *clientv3.Client, key string) {
	resp, err := cli.Get(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	if len(resp.Kvs) == 0 {
		fmt.Printf("%s:%s\n", key, "NOT_FOUND")
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%s:%s\n", string(ev.Key), string(ev.Value))
	}
}

