package main

import (
	"flag"
	"fmt"
	"github.com/goCrontab/master"
	"runtime"
	"time"
)

func initEnv() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func initArgs() {
	// master -c ./config.json
	flag.StringVar(&confFile, "c", "./config.json", "configure file")
	flag.Parse()
}

var (
	confFile string //配置文件路径
)

func main() {
	var (
		err error
	)
	//初始化参数
	initArgs()

	// 初始化go采用的核数
	initEnv()

	//初始化配置
	if err = master.InitConfig(confFile); err != nil {
		goto END
	}

	//初始化任务管理员
	if err = master.InitJobMgr(); err != nil {
		goto END
	}

	//初始化Api服务
	if err = master.InitApiServer();err != nil {
		goto END
	}

	for {
		time.Sleep(time.Second)
	}
	return
END:
	fmt.Println(err)
}
