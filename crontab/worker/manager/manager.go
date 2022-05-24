package manager

import (
	"context"
	"dbsmonitor/crontab/job"
	"dbsmonitor/crontab/worker"
	"dbsmonitor/crontab/worker/config"
	"dbsmonitor/crontab/worker/scheduler"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	"go.etcd.io/etcd/client/v3"
	"time"
)

type manager struct {
	client       *clientv3.Client
	kv           clientv3.KV
	lease        clientv3.Lease
	watcher      clientv3.Watcher
	fromRevision int64
}

var JobMgr manager

func InitJobMgr() error {
	c := clientv3.Config{
		Endpoints:   config.Config.EtcdEndpoints,
		DialTimeout: time.Duration(config.Config.EtcdDialTimeout) * time.Millisecond,
	}
	client, err := clientv3.New(c)
	if err != nil {
		return err
	}
	JobMgr = manager{
		client:  client,
		kv:      clientv3.NewKV(client),
		lease:   clientv3.NewLease(client),
		watcher: clientv3.NewWatcher(client),
	}
	err = JobMgr.healthCheck()
	if err != nil {
		return err
	}
	err = JobMgr.Run()
	if err != nil {
		return err
	}
	return nil
}

func (j *manager) Run() error {
	getResp, err := j.kv.Get(context.Background(), job.SAVEDIR, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, kvPair := range getResp.Kvs {
		job, err := job.Unpack(kvPair.Value)
		if err != nil {
			fmt.Println(err, string(kvPair.Value))
			continue
		}
		scheduler.Scheduler.UpdateJob(&worker.JobEvent{Job: job, EventType: worker.PutEvent})
		fmt.Println(job)
	}
	j.getCurrentRevision(getResp)
	go j.watchUpdateJob()
	go j.watchKillJob()
	return nil
}

func (j *manager) healthCheck() error {
	leaseResp, err := j.lease.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}
	leaseId := leaseResp.ID
	healthName := job.HEALTHDIR + config.Config.WorkerId
	_, err = j.client.Put(context.TODO(), healthName, time.Now().Format("2006-01-02 15:04:05"), clientv3.WithLease(leaseId))
	if err != nil {
		return err
	}
	_, err = j.lease.KeepAlive(context.TODO(), leaseId)
	if err != nil {
		return err
	}
	return nil
}

func (j *manager) getCurrentRevision(response *clientv3.GetResponse) {
	j.fromRevision = response.Header.Revision + 1
}

func (j *manager) watchUpdateJob() {
	watchChan := JobMgr.client.Watch(context.Background(), job.SAVEDIR,
		clientv3.WithRev(j.fromRevision), clientv3.WithPrefix())
	for resp := range watchChan {
		for _, event := range resp.Events {
			switch event.Type {
			case mvccpb.PUT:
				putEventSend(event)
			case mvccpb.DELETE:
				delEventSend(event)
			}
		}
	}
}

func (j *manager) watchKillJob() {
	watchChan := JobMgr.client.Watch(context.Background(), job.KILLDIR,
		clientv3.WithRev(j.fromRevision), clientv3.WithPrefix())
	for resp := range watchChan {
		for _, event := range resp.Events {
			switch event.Type {
			case mvccpb.PUT:
				killEventSend(event)
			}
		}
	}
}
