package master

import (
	"context"
	"dbsmonitor/crontab/job"
	config2 "dbsmonitor/crontab/master/config"
	"encoding/json"
	"errors"
	"github.com/gorhill/cronexpr"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

type jobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

var (
	JobMgr *jobMgr
)

func InitJobMgr() error {
	var (
		config clientv3.Config
	)
	JobMgr = &jobMgr{}
	config = clientv3.Config{
		Endpoints:   config2.Config.EtcdEndpoints,
		DialTimeout: time.Duration(config2.Config.EtcdDialTimeout) * time.Millisecond,
	}
	client, err := clientv3.New(config)
	if err != nil {
		return err
	}
	JobMgr.client = client
	JobMgr.kv = clientv3.NewKV(client)
	JobMgr.lease = clientv3.NewLease(client)
	return nil
}

// SaveJob add the job to etcd and return the old job
func (j *jobMgr) SaveJob(jobItem *job.Job) (*job.Job, error) {
	jobKey := job.SAVEDIR + jobItem.Name
	var jobValue []byte
	var oldValue *job.Job
	jobValue, err := json.Marshal(jobItem)
	if err != nil {
		return nil, err
	}
	if _, err := cronexpr.Parse(jobItem.CronExpr); err != nil {
		return nil, err
	}
	putResponse, err := j.kv.Put(context.Background(), jobKey, string(jobValue), clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}
	if putResponse.PrevKv == nil {
		return nil, nil
	}
	prevValue := putResponse.PrevKv.Value
	err = json.Unmarshal(prevValue, &oldValue)
	if err != nil {
		err = nil
	}
	return oldValue, err
}

func (j *jobMgr) DeleteJob(name string) (*job.Job, error) {
	jobName := job.SAVEDIR + name
	deleteResponse, err := j.kv.Delete(context.Background(), jobName, clientv3.WithPrevKV())
	if err != nil {
		return nil, err
	}
	if len(deleteResponse.PrevKvs) > 0 {
		var oldValue *job.Job
		prevValue := deleteResponse.PrevKvs[0].Value
		json.Unmarshal(prevValue, &oldValue)
		return oldValue, nil
	}
	return nil, errors.New("invalid job name")
}

func (j *jobMgr) ListJob() ([]*job.Job, error) {
	listResponse, err := j.kv.Get(context.Background(), job.SAVEDIR, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if listResponse == nil {
		return nil, nil
	}
	jobList := make([]*job.Job, 0)
	for _, kv := range listResponse.Kvs {
		var job job.Job
		err := json.Unmarshal(kv.Value, &job)
		if err != nil {
			continue
		}
		jobList = append(jobList, &job)
	}
	return jobList, nil
}

// KillJob watch etcdctl watch "/cron/kill" --prefix
func (j *jobMgr) KillJob(name string) error {
	lease, err := j.lease.Grant(context.Background(), 1)
	if err != nil {
		return err
	}
	jobName := job.KILLDIR + name
	_, err = j.kv.Put(context.Background(), jobName, "", clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}
	return nil
}

func (j *jobMgr) GetWorkers() ([]map[string]string, error) {
	getResp, err := j.kv.Get(context.TODO(), job.HEALTHDIR, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	workers := make([]map[string]string, 1)
	for _, value := range getResp.Kvs {
		workerName := job.GetWorkerName(value.Key)
		workers = append(workers, map[string]string{"name": workerName, "start-time": string(value.Value)})
	}
	return workers, nil
}
