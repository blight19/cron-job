package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

func main() {
	var (
		cli     *clientv3.Client
		err     error
		conf    clientv3.Config
		kv      clientv3.KV
		putResp *clientv3.PutResponse
	)

	conf = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second}
	if cli, err = clientv3.New(conf); err != nil {
		fmt.Println("Error:", err)
		return
	}
	kv = clientv3.NewKV(cli)
	if putResp, err = kv.Put(context.Background(), "/cron/jobs/job1", "hello3", clientv3.WithPrevKV()); err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(putResp.Header.Revision)
		fmt.Println(string(putResp.PrevKv.Value))
	}

}
