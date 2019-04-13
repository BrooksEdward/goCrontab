package main

import (
	"flag"
	"fmt"
	"github.com/caiguangyin/goCrontab/worker"
	"runtime"
	"time"
)

var (
	configFile string
)

func initEnv()  {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func initArgs()  {
	flag.StringVar(&configFile, "c", "./config.json", "configuration file")
	flag.Parse()
}

// init()函数会在main()函数之前被Go执行
func init() {
	initArgs()
	initEnv()
}

func main() {
	var (
		err error
	)
	// 初始化配置文件
	if err = worker.InitConfig(configFile); err != nil {
		goto ERR
	}

	fmt.Printf("%#v\n", *worker.G_conf)

	// 启动执行器
	if err = worker.InitExcutor(); err != nil {
		goto ERR
	}


	// 初始化任务调度器
	worker.InitScheduler()

	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(time.Second)
	}

	return
ERR:
	fmt.Println(err)
}

