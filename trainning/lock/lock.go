package main

import (
	"context"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"strconv"
	"time"
)

const prefix = "/lock"

func main() {
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, DialTimeout: 5 * time.Second})
	if err != nil {
		fmt.Println(err)
		return
	}
	log.Println("connect success")
	defer client.Close()

	for i := 1; i < 3; i++ {
		go work(strconv.Itoa(i), client)
	}

	select {}
}

func work(workerId string, cli *clientv3.Client) {
	session, err := concurrency.NewSession(cli, concurrency.WithTTL(10))
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		m := concurrency.NewMutex(session, prefix)
		if err := m.TryLock(context.Background()); err != nil {
			log.Printf("[%s]got lock failed:%s", workerId, err)
			continue
		} else {
			log.Printf("[%s]got lock success", workerId)
		}
		time.Sleep(time.Second)
		m.Unlock(context.Background())
		log.Printf("[%s] unlock!", workerId)
	}

}
