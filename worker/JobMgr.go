package worker

import (
	"fmt"
	"github.com/Go-zh/net/context"
	"github.com/caiguangyin/goCrontab/common"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"strings"
	"time"
)

type JobMgr struct {
	Client  *clientv3.Client
	Kv      clientv3.KV
	Lease   clientv3.Lease
	Watcher clientv3.Watcher
}

// 声明单例
var G_jobMgr *JobMgr

//监听任务变化
func (jm *JobMgr) watchJobs() (err error) {
	var (
		jobEvent *common.JobEvent
	)

	// 1. get一下/cron/jobs/ 目录下的所有任务，并且获知当前集群的Revision
	getResp, err := jm.Kv.Get(context.TODO(), common.SAVE_JOB_DIR, clientv3.WithPrefix())
	if err != nil {
		return
	}

	// 将当前etcd中的任务取出，并推送给调度协程处理
	for _, kvPair := range getResp.Kvs {
		job, err := common.Deserialize(kvPair.Value)
		if err != nil {
			continue
		}
		jobEvent = common.BuildJobEvent(common.SAVE_JOB_EVENT, job)
		// TODO: 把这个job同步给Scheduler
		G_scheduler.PushJobEvent(jobEvent)
	}


	// 2. 从第1步获取的当前集群Revision之后开始监听变化事件
	go func() {		// 启动一个监听协程
		// 从get revision后开始监听
		startWatch := getResp.Header.Revision + 1
		// 监听/cron/jobs/目录的后续变化
		watchChans := jm.Watcher.Watch(context.TODO(), common.SAVE_JOB_DIR, clientv3.WithRev(startWatch), clientv3.WithPrefix())

		// 处理监听事件
		for watchResp := range watchChans {
			for _, event := range watchResp.Events {
				switch event.Type {
				case mvccpb.PUT:
					// 构建更新事件
					job, err := common.Deserialize(event.Kv.Value)
					if err != nil {
						continue
					}
					jobEvent = common.BuildJobEvent(common.SAVE_JOB_EVENT, job)
				case mvccpb.DELETE:
					jobName := strings.TrimPrefix(string(event.Kv.Key), common.SAVE_JOB_DIR)
					job := &common.Job{JobName: jobName}
					// 构建删除事件
					jobEvent = common.BuildJobEvent(common.DEL_JOB_EVENT, job)
				}
				fmt.Println("===> : ", *jobEvent)
				// TODO: 将事件同步给Scheduler
				G_scheduler.PushJobEvent(jobEvent)
			}
		}
	}()

	return
}

// 初始化任务管理器
func InitJobMgr() (err error) {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   G_conf.EtcdEndPoints,
		DialTimeout: time.Duration(G_conf.EtcdDialTimeout) * time.Millisecond,
	})
	if err != nil {
		return
	}

	kv := clientv3.NewKV(client)
	lease := clientv3.NewLease(client)
	watcher := clientv3.NewWatcher(client)

	G_jobMgr = &JobMgr{
		Client:  client,
		Kv:      kv,
		Lease:   lease,
		Watcher: watcher,
	}

	// 启动任务监听
	if err = G_jobMgr.watchJobs(); err != nil {
		return
	}

	return
}
