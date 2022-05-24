package executor

import (
	"context"
	"dbsmonitor/crontab/job"
	"dbsmonitor/crontab/worker/config"
	"dbsmonitor/crontab/worker/handlers"
	"errors"
	"fmt"
	"github.com/gorhill/cronexpr"
	"go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"os/exec"
	"runtime"
	"sync"
	"time"
)

var JobExistError = errors.New("Job Exist Error ")

//var DoingJob sync.Map

type BeforeJobFunc func(*job.Job) error
type AfterJobFunc func(*job.Job, *ExecInfo)

var goos string

type JobSchedulePlan struct {
	Job      *job.Job
	Expr     *cronexpr.Expression //the parsed job cron expression
	NextTime time.Time            //the moment do the job next
}

type ExecInfo struct {
	*JobSchedulePlan
	Out          string
	Err          error
	PlanTime     int64
	ScheduleTime int64
	StartTime    int64
	EndTime      int64
}

type JobExecutor struct {
	client         *clientv3.Client
	workerId       string
	lockExpireTime int
	jobs           chan *ExecInfo
	result         chan *ExecInfo
	logger         *handlers.Logger
	DoingJob       sync.Map
}

func NewJobExecutor() (*JobExecutor, error) {
	goos = runtime.GOOS
	conf := clientv3.Config{
		Endpoints:   config.Config.EtcdEndpoints,
		DialTimeout: time.Duration(config.Config.EtcdDialTimeout) * time.Millisecond,
	}
	logger, err := handlers.NewLogger()
	if err != nil {
		return nil, err
	}
	if etcdClient, err := clientv3.New(conf); err != nil {
		return nil, err
	} else {
		jobExecutor := &JobExecutor{
			client:         etcdClient,
			workerId:       config.Config.WorkerId,
			jobs:           make(chan *ExecInfo, 1000),
			result:         make(chan *ExecInfo, 1000),
			lockExpireTime: 5,
			logger:         logger,
		}
		go jobExecutor.run()
		return jobExecutor, nil
	}
}

func buildLogDocument(result *ExecInfo) *handlers.LogDocument {
	errString := ""
	if result.Err != nil {
		errString = result.Err.Error()
	}
	return &handlers.LogDocument{
		Name:         result.Job.Name,
		Command:      result.Job.Command,
		Err:          errString,
		OutPut:       result.Out,
		PlanTime:     result.PlanTime,
		ScheduleTime: result.ScheduleTime,
		StartTime:    result.StartTime,
		EndTime:      result.EndTime,
	}
}

func (j *JobExecutor) run() {
	for {
		select {
		case jobItem := <-j.jobs:
			go j.execute(jobItem)
		case result := <-j.result:
			document := buildLogDocument(result)
			j.logger.AddLog(document)
		}
	}
}

// AddJob Add a job to the jobExecutor
func (j *JobExecutor) AddJob(plain *JobSchedulePlan) {
	info := ExecInfo{
		JobSchedulePlan: plain,
		PlanTime:        plain.NextTime.UnixMicro(),
		ScheduleTime:    time.Now().UnixMicro(),
		StartTime:       0,
		EndTime:         0,
	}
	j.jobs <- &info
}

func (j *JobExecutor) execute(info *ExecInfo) {
	session, err := concurrency.NewSession(j.client, concurrency.WithTTL(j.lockExpireTime))
	defer session.Close()
	jobItem := info.Job
	if err != nil {
		info.Err = err
		j.result <- info
		return
	}
	lockName := job.LOCKDIR + jobItem.Name
	m := concurrency.NewMutex(session, lockName)
	defer m.Unlock(context.TODO())
	if err := m.TryLock(context.Background()); err == concurrency.ErrLocked {
		//lock failed , other worker had the lock
		return
	} else if err != nil {
		//other error,return the error
		info.Err = err
		j.result <- info
		return
	}
	log.Printf("[%s] got job:%s \n", j.workerId, jobItem.Name)
	ctx, cancel := context.WithCancel(context.Background())
	err = j.addDoingJob(jobItem, cancel)
	if err == JobExistError {
		// job is working on this worker , return
		return
	}
	defer j.deleteDoingJob(jobItem)
	runCmd(info, ctx)
	j.result <- info
}

func runCmd(info *ExecInfo, ctx context.Context) {

	info.StartTime = time.Now().UnixMicro()
	var cmd *exec.Cmd
	if goos == "windows" {
		cmd = exec.CommandContext(ctx, "C:\\bash.exe", "-c", info.Job.Command)
	} else if goos == "linux" {
		cmd = exec.CommandContext(ctx, "/bin/bash", "-c", info.Job.Command)
	} else {
		panic("unsupported system")
	}
	output, err := cmd.CombinedOutput()
	info.EndTime = time.Now().UnixMicro()
	info.Out = string(output)
	info.Err = err
}

// addDoingJob
// avoid the repeated job on this worker
func (j *JobExecutor) addDoingJob(job *job.Job, cancelFunc context.CancelFunc) error {
	_, ok := j.DoingJob.Load(job.Name)
	if ok {
		return JobExistError
	}
	j.DoingJob.Store(job.Name, cancelFunc)
	return nil
}

func (j *JobExecutor) deleteDoingJob(job *job.Job) {
	j.DoingJob.Delete(job.Name)
}

func echoResult(out *ExecInfo) {
	//if out.Err != nil {
	//	log.Printf("[%s] Error:%s\n", out.Job.Name, out.Err.Error())
	//} else {
	//	log.Printf("[%s] Result:%s\n", out.Job.Name, out.Out)
	//}
	fmt.Println(out)
}
