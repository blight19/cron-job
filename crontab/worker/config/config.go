package config

import (
	"encoding/json"
	"io/ioutil"
)

type config struct {
	EtcdEndpoints   []string `json:"etcdEndpoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	WorkerId        string   `json:"worker-id"`
	MongoUri        string   `json:"mongo-uri"`
	MongoBatch      int      `json:"mongo-batch"`
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
