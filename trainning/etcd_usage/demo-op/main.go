package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"

	"time"
)

func main() {
	var (
		config clientv3.Config
		client *clientv3.Client
		err    error
		opResp clientv3.OpResponse
	)
	config = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		fmt.Println(err)
		return
	}

	kv := clientv3.NewKV(client)
	putOp := clientv3.OpPut("/cron/jobs/job8", "i am job8")
	if opResp, err = kv.Do(context.Background(), putOp); err != nil {
		fmt.Println("error", err)
		return
	}
	fmt.Println(opResp.Put().Header.Revision)

}
