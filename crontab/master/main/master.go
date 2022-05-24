package main

import (
	"dbsmonitor/crontab/master"
	"dbsmonitor/crontab/master/api"
	"dbsmonitor/crontab/master/config"
	"dbsmonitor/crontab/master/logger"
	"flag"
	"fmt"
	"log"
)

func init() {
	var (
		confFile string
		err      error
	)
	flag.StringVar(&confFile, "config", "./master.json", "config path")
	flag.Parse()
	if err = config.InitConfig(confFile); err != nil {
		panic(err)
	} else {
		log.Println("Read Config success!!!")
	}
	if err = master.InitJobMgr(); err != nil {
		panic(err)
	} else {
		log.Println("Init Job Manager success!!!")
	}
	if err = logger.InitLogger(); err != nil {
		panic(err)
	} else {
		log.Println("Init Logger success!!!")
	}
	if err = api.InitApiServer(); err != nil {
		panic(err)
	} else {
		log.Println("Init Api Server success!!!")
	}

}

func main() {
	fmt.Println("running")
	select {}

}
