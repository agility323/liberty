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
		Endpoints:	 []string{"10.1.71.45:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		// handle error!
		fmt.Printf("connect to etcd failed, err:%v\n", err)
		return
	}
	fmt.Println("connect to etcd success")
	defer cli.Close()
	resp, err := cli.Get(context.Background(), "", clientv3.WithPrefix())
	//fmt.Printf("resp %v\n", resp)
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%v:%v\n", string(ev.Key), string(ev.Value))
		//fmt.Printf("%v:%v\n", (ev.Key), (ev.Value))
	}
}

