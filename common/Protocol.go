package common

import (
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

type Job struct {
	JobName  string `json:"jobName"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

// http响应
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

// 构建事件
type JobEvent struct {
	EventType int
	Job       *Job
}

// 任务调度计划
type JobSchedulePlan struct {
	Job      *Job                 // 要调度的任务信息
	Expr     *cronexpr.Expression // 解析好的cronExpr表达式
	NextTime time.Time            // 任务下次调度时间
}

// 任务执行状态表
type JobExecInfo struct {
	Job      *Job      // 任务信息
	PlanTime time.Time // 计划调度任务时间
	RealTime time.Time // 真实任务调度时间
}

// 任务执行结果
type JobExecResult struct {
	JobExecInfo   *JobExecInfo // 任务信息
	OutPut        []byte       // 任务Shell命令执行输出结果
	Err           error        // 任务执行错误
	StartExecTime time.Time    // 任务开始执行时间
	EndExecTime   time.Time    // 任务执行结束时间
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

// 反序列化job
func Deserialize(value []byte) (ret *Job, err error) {
	job := &Job{}
	err = json.Unmarshal(value, job)
	if err != nil {
		return
	}
	ret = job
	return
}

// 构建任务变化事件
func BuildJobEvent(eventType int, job *Job) (event *JobEvent) {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

// 构造任务执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	// 解析job的Cron表达式
	expr, err := cronexpr.Parse(strings.TrimSpace(job.CronExpr))
	if err != nil {
		return
	}

	// 生成任务调度计划对象
	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}
	return
}

// 构造任务执行状态信息
func BuildJobExecInfo(jobSchedulePlan *JobSchedulePlan) (jobExecInfo *JobExecInfo) {
	jobExecInfo = &JobExecInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime, // 计划调度时间
		RealTime: time.Now(),               // 真实调度时间
	}
	return
}
