package config

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	ApiPort         int      `json:"api-port"`
	ApiReadTimeout  int      `json:"api-read-timeout"`
	ApiWriteTimeout int      `json:"api-write-timeout"`
	EtcdEndpoints   []string `json:"etcdEndpoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	MongoUri        string   `json:"mongo-uri"`
	LogPageSize     int64    `json:"log-page-size"`
}

var Config *config

func InitConfig(filename string) error {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	err = json.Unmarshal(content, &Config)
	return err
}
