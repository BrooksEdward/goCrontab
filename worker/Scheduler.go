package worker

import (
	"fmt"
	"github.com/caiguangyin/goCrontab/common"
	"time"
)

// 任务调度，管道里面存放任务变化事件（JobEvent）
type Scheduler struct {
	jobEventChan      chan *common.JobEvent              // etcd任务事件队列
	jobPlanTable      map[string]*common.JobSchedulePlan // 任务调度计划表
	jobExecTable      map[string]*common.JobExecInfo     // 任务执行表
	jobExecResultChan chan *common.JobExecResult         // 任务执行结果队列
}

var (
	G_scheduler *Scheduler
)

// 尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobSchedulePlan *common.JobSchedulePlan) {
	// 调度和执行是2件事情
	// 执行的任务可能会运行很久（假如某任务系每秒执行一次，可是每次执行耗时60秒，那么该任务每分钟会被调度60次，但是只能执行1次，防止并发。）
	// 如果任务存在于任务执行表中，则表示该任务正在执行，则需要调过本次调度
	if _, jobExecuting := scheduler.jobExecTable[jobSchedulePlan.Job.JobName]; jobExecuting {
		fmt.Println("上次执行尚未完成，跳过本次调度: ", jobSchedulePlan.Job.JobName)
		return
	}

	// 构造任务执行信息
	jobExecInfo := common.BuildJobExecInfo(jobSchedulePlan)
	// 将任务执行信息加入任务执行表中
	scheduler.jobExecTable[jobExecInfo.Job.JobName] = jobExecInfo

	// TODO： 执行任务。任务执行完后，还需要将任务执行表中删除此任务，以便任务下一次被成功调度。
	G_excutor.ExcuteJob(jobExecInfo)
	fmt.Println("执行任务： ", jobExecInfo.Job.JobName, jobExecInfo.PlanTime, jobExecInfo.RealTime)
}

// 重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
	var (
		now      time.Time
		jobPlan  *common.JobSchedulePlan
		nearTime *time.Time
	)
	//如果任务表为空的话，随便睡眠多久
	if len(scheduler.jobPlanTable) == 0 {
		scheduleAfter = 1 * time.Second
		return
	}

	// 当前时间
	now = time.Now()

	// 1. 遍历所有任务
	for _, jobPlan = range scheduler.jobPlanTable {
		// 2. 过期的任务立即执行
		// 判断任务的下次执行时间是否早于或等于当前时间，如果是则表示已经到了任务的下次执行时间了
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			// TODO: 尝试执行任务（可能任务上次还在执行中，又到了该任务执行的时间了。）
			scheduler.TryStartJob(jobPlan)
			// 更新下次执行时间
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}

		// 统计最近一个要过期的任务时间
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}

	// 3. 统计最近将要过期的任务的时间
	// 距离下次任务执行间隔时间 (公式：最近要执行的任务调度时间 - 当前时间)
	scheduleAfter = (*nearTime).Sub(now)

	return
}

// 处理任务事件，让内存中维护的任务列表与etcd中保持一致
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	switch jobEvent.EventType {
	case common.SAVE_JOB_EVENT: // TODO: 保存任务事件
		jobSchedulePlan, err := common.BuildJobSchedulePlan(jobEvent.Job)
		if err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.JobName] = jobSchedulePlan
	case common.DEL_JOB_EVENT: // TODO: 删除任务事件
		_, jobExisted := scheduler.jobPlanTable[jobEvent.Job.JobName]
		if jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.JobName)
		}
	}
}

// 处理任务执行结果，从任务执行表中删除执行记录，以便下次调度任务
func (scheduler *Scheduler) handleJobExecResult(jobExecResult *common.JobExecResult) {
	// 删除执行记录
	delete(scheduler.jobExecTable, jobExecResult.JobExecInfo.Job.JobName)

	fmt.Println("任务执行完成：", jobExecResult.JobExecInfo.Job.JobName, string(jobExecResult.OutPut),"错误信息：", jobExecResult.Err)
}

// 循环调度任务
func (scheduler *Scheduler) scheduleLoop() {
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobExecResult *common.JobExecResult
	)

	// 初始化一次(1秒)
	scheduleAfter = scheduler.TrySchedule()

	// 调度的延迟定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	for {
		select {
		case jobEvent = <-scheduler.jobEventChan:
			// TODO: 对内存中维护的任务列表做增删改查
			scheduler.handleJobEvent(jobEvent) // 处理任务事件，保证内存中的任务列表与Etcd中保存的任务列表是完全一致的。
		case <-scheduleTimer.C: // 最近的任务到期了
		case jobExecResult = <-scheduler.jobExecResultChan:
			scheduler.handleJobExecResult(jobExecResult)
		}

		// 调度一次任务
		scheduleAfter = scheduler.TrySchedule()
		// 重置调度间隔时间
		scheduleTimer.Reset(scheduleAfter)
	}
}

// 向队列中推送任务事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

// 向队列中推送任务执行结果
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecResult) {
	scheduler.jobExecResultChan <- jobResult
}

// 初始化任务调度器
func InitScheduler() {
	G_scheduler = &Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulePlan),
		jobExecTable:      make(map[string]*common.JobExecInfo),
		jobExecResultChan: make(chan *common.JobExecResult, 1000),
	}

	// 启动任务调度
	go G_scheduler.scheduleLoop()
	return
}
