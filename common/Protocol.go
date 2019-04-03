package common

import "encoding/json"

const (
	SAVE_JOB_DIR = "/cron/jobs/"
)

type Job struct {
	JobName  string `json:"jobName"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	// 定义响应内容
	var response Response

	response.Errno = errno
	response.Msg = msg
	response.Data = data

	// 序列化成json数据
	resp, err = json.Marshal(response)
	return
}
