package master

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Config struct {
	ApiPort         int      `json:"apiPort"`         //apiserver端口
	ApiReadTimeout  int      `json:"apiReadTimeout"`  //apiserver http服务读超时时间
	ApiWriteTimeout int      `json:"apiWriteTimeout"` //apiserver http服务写超时时间
	EtcdEndPoints   []string `json:"etcdEndPoints"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
}

var G_conf *Config

func InitConfig(confFile string) (err error) {
	var (
		c    []byte
		conf *Config
	)

	c, err = ioutil.ReadFile(confFile)
	if err != nil {
		return
	}

	conf = &Config{}
	err = json.Unmarshal(c, conf)
	if err != nil {
		return
	}
	fmt.Println(*conf)

	// 赋值单例
	G_conf = conf

	return
}
