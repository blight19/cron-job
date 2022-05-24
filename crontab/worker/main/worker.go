package main

import (
	"dbsmonitor/crontab/worker/config"
	"dbsmonitor/crontab/worker/manager"
	"dbsmonitor/crontab/worker/scheduler"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
)

var (
	confFile string
	err      error
)

func init() {
	flag.StringVar(&confFile, "config", "./worker.json", "指定配置文件路径")
	flag.Parse()
	if err = config.InitConfig(confFile); err != nil {
		panic(err)
	} else {
		log.Println("Read Config success!!!")
	}
}
func main() {
	scheduler.InitScheduler()
	err := manager.InitJobMgr()
	if err != nil {
		fmt.Println(err)
		return
	}
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	select {}
}
