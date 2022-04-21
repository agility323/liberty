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
	// put
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	_, err = cli.Put(ctx, "q1mi1", "dsb")
	cancel()
	if err != nil {
		fmt.Printf("put to etcd failed, err:%v\n", err)
		return
	}
	// get
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	//resp, err := cli.Get(ctx, "Zm9v")
	resp, err := cli.Get(ctx, "", clientv3.WithPrefix())
	fmt.Printf("resp %v\n", resp)
	cancel()
	if err != nil {
		fmt.Printf("get from etcd failed, err:%v\n", err)
		return
	}
	for _, ev := range resp.Kvs {
		fmt.Printf("%v:%v\n", string(ev.Key), string(ev.Value))
		fmt.Printf("%v:%v\n", (ev.Key), (ev.Value))
	}
	//cli.
}

