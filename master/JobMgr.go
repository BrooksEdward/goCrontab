package master

import (
	"encoding/json"
	"github.com/Go-zh/net/context"
	"github.com/goCrontab/common"
	"github.com/pkg/errors"
	"go.etcd.io/etcd/clientv3"
	"time"
)

type JobMgr struct {
	Client *clientv3.Client
	Kv     clientv3.KV
	Lease  clientv3.Lease
}

// 声明单例
var G_jobMgr *JobMgr

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

	G_jobMgr = &JobMgr{
		Client: client,
		Kv:     kv,
		Lease:  lease,
	}

	return
}

// 任务保存或更新
func (jm *JobMgr) JobSave(job *common.Job) (oldJob *common.Job, err error) {
	var oldJobObj = &common.Job{}

	// job key
	jobKey := common.SAVE_JOB_DIR + job.JobName

	// 序列化job struct
	jobJson, err := json.Marshal(job)
	if err != nil {
		return
	}

	// 将Job存入etcd
	putResp, err := jm.Kv.Put(context.TODO(), jobKey, string(jobJson), clientv3.WithPrevKV())
	if err != nil {
		return
	}
	// 如果PrevKv不为nil，则表示是更新操作，然后获取更新前的Job
	if putResp.PrevKv != nil {
		prevJob := putResp.PrevKv.Value
		if err = json.Unmarshal(prevJob, oldJobObj); err != nil {
			err = nil
			return
		}
		oldJob = oldJobObj
		return
	}
	return
}

// 任务删除
func (jm *JobMgr) JobDelete(job *common.Job) (oldJob *common.Job, err error) {
	// 定义job key
	var jobKey = common.SAVE_JOB_DIR + job.JobName

	// 根据job key删除任务
	delResp, err := jm.Kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV())
	if err != nil {
		return
	}

	if len(delResp.PrevKvs) > 0 {
		// 反序列化被删除的任务并返回
		oldJobObj := &common.Job{}
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, oldJobObj); err != nil {
			return
		}
		oldJob = oldJobObj
	} else {
		err = errors.New("The task does not exist.")
	}

	return
}

func (jm *JobMgr) JobList() (jobList []*common.Job, err error) {
	jobDir := common.SAVE_JOB_DIR

	getResp, err := jm.Kv.Get(context.TODO(), jobDir, clientv3.WithPrefix())
	if err != nil {
		return
	}

	if len(getResp.Kvs) > 0 {
		for _, v := range getResp.Kvs {
			job := &common.Job{}
			_ = json.Unmarshal(v.Value, job)
			jobList = append(jobList, job)
		}
	} else {
		err = errors.New("No task.")
	}

	return
}