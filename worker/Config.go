package worker

import (
	"encoding/json"
	"io/ioutil"
)

type Config struct {
	EtcdEndPoints   []string `json:"etcdEndPoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	BashPath        []string `json:"bashPath"`
}

var G_conf *Config

func InitConfig(confFile string) (err error) {
	bytes, err := ioutil.ReadFile(confFile)
	if err != nil {
		return
	}

	conf := &Config{}
	if err = json.Unmarshal(bytes, conf); err != nil {
		return
	}

	G_conf = conf

	return
}
