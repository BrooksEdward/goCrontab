package master

import (
	"encoding/json"
	"fmt"
	"github.com/goCrontab/common"
	"net"
	"net/http"
	"strconv"
	"time"
)

// 任务的HTTP接口
type ApiServer struct {
	httpServer *http.Server
}

var G_apiServer *ApiServer

//保存任务接口
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		bytes   []byte
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	// 获取form表示提交的job任务，然后反序列化
	postJob = req.PostForm.Get("job")
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}

	oldJob, err = G_jobMgr.JobSave(&job)
	if err != nil {
		goto ERR
	}

	// 返回正常应答 {"errno":0, "msg": "successed", "data": {oldjob}}
	bytes, err = common.BuildResponse(0, "success", oldJob)
	if err == nil {
		resp.Write(bytes)
	}

	return

ERR:
	// 返回错误应答 {"errno":-1, "msg": err.Error(), "data": nil}
	bytes, err = common.BuildResponse(-1, err.Error(), nil)
	if err == nil {
		resp.Write(bytes)
	}
}

// 任务删除接口
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		bytes   []byte
	)

	if err = req.ParseForm(); err != nil {
		fmt.Println(1)
		goto ERR
	}
	// 获取任务
	postJob = req.PostForm.Get("job")
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		fmt.Println(2)
		goto ERR
	}

	// 从etcd中删除任务
	oldJob, err = G_jobMgr.JobDelete(&job)
	if err != nil {
		fmt.Println(3)
		goto ERR
	}

	// 构建响应
	bytes, err = common.BuildResponse(0, "success", oldJob)
	if err == nil {
		resp.Write(bytes)
	}

	return
ERR:
	// 构建错误响应
	bytes, err = common.BuildResponse(-1, err.Error(), nil)
	if err == nil {
		resp.Write(bytes)
	}
}

// 查看所有任务
func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		err	error
		jobList []*common.Job
		bytes []byte
	)
	// 获取任务列表
	jobList, err = G_jobMgr.JobList()
	if err != nil {
		goto ERR
	}
	// 构建任务列表响应
	bytes, err = common.BuildResponse(0, "success", jobList)
	if err == nil {
		resp.Write(bytes)
	}

	return
ERR:
	bytes, err = common.BuildResponse(-1, err.Error(), nil)
	if err == nil {
		resp.Write(bytes)
	}
}

// 初始化服务
func InitApiServer() (err error) {
	var (
		httpServer *http.Server
		mux        *http.ServeMux
		listener   net.Listener
	)

	//配置路由
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/del", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)

	//启动TCP监听
	listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_conf.ApiPort))
	if err != nil {
		return
	}

	//创建一个HTTP服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_conf.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_conf.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}
	// 赋值单例
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}

	// 启动服务
	go httpServer.Serve(listener)

	return
}
