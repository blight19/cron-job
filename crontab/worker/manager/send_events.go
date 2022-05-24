package manager

import (
	"dbsmonitor/crontab/job"
	"dbsmonitor/crontab/worker"
	"dbsmonitor/crontab/worker/scheduler"
	clientv3 "go.etcd.io/etcd/client/v3"
	"log"
)

func putEventSend(event *clientv3.Event) {
	job, err := job.Unpack(event.Kv.Value)
	if err != nil {
		log.Println("unpack job error ", err)
		return
	}
	log.Println("Received Update Job signal", job.Name)
	jobEvent := worker.JobEvent{
		Job:       job,
		EventType: worker.PutEvent,
	}
	scheduler.Scheduler.UpdateJob(&jobEvent)
}

func delEventSend(event *clientv3.Event) {
	jobName := job.GetDeleteName(event.Kv.Key)
	log.Println("Received Delete Job signal", jobName)
	jobEvent := worker.JobEvent{
		Job:       &job.Job{Name: jobName},
		EventType: worker.DeleteEvent,
	}
	scheduler.Scheduler.UpdateJob(&jobEvent)
}

func killEventSend(event *clientv3.Event) {
	jobName := job.GetKillName(event.Kv.Key)
	jobEvent := worker.JobEvent{
		Job:       &job.Job{Name: jobName},
		EventType: worker.KillEvent,
	}
	log.Println("Received Kill Job signal", jobName)
	scheduler.Scheduler.UpdateJob(&jobEvent)
}
