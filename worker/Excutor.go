package worker

import (
	"github.com/Go-zh/net/context"
	"github.com/caiguangyin/goCrontab/common"
	"os/exec"
	"time"
)

type Excutor struct {

}

var (
	G_excutor *Excutor
)

func InitExcutor() (err error) {
	G_excutor = &Excutor{

	}
	return
}

func (executor *Excutor) ExcuteJob(jobExecInfo *common.JobExecInfo) {

	// 启动协程并发执行任务shell命令
	go func() {

		jobExecResult := &common.JobExecResult{
			JobExecInfo:jobExecInfo,
			OutPut:make([]byte, 0),
		}
		// 记录任务开始执行的时间
		jobExecResult.StartExecTime = time.Now()

		// 执行任务Shell命令，并获取命令输出内容
		cmd := exec.CommandContext(context.TODO(), G_conf.BashPath[0], G_conf.BashPath[1], jobExecInfo.Job.Command)
		output, err := cmd.CombinedOutput()

		// 记录任务执行结束的时间
		jobExecResult.EndExecTime = time.Now()

		// 记录任务执行结果
		jobExecResult.OutPut = output
		jobExecResult.Err = err

		// 任务执行完成后，把执行结果返回给Scheduler，然后Scheduler会从ExecutingTable中删除执行记录
		G_scheduler.PushJobResult(jobExecResult)
	}()
}
