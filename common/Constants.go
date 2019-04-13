package common


const (
	// 保存任务的Etcd目录
	SAVE_JOB_DIR = "/cron/jobs/"

	// 杀死任务通知目录
	KILL_JOB_DIR = "/cron/killer/"

	// 保存和更新任务事件
	SAVE_JOB_EVENT = 1

	// 删除任务事件
	DEL_JOB_EVENT = 2
)