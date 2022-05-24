package scheduler

import (
	"context"
	"dbsmonitor/crontab/worker"
	"dbsmonitor/crontab/worker/executor"
	"fmt"
	"github.com/gorhill/cronexpr"
	"log"
	"time"
)

type scheduler struct {
	jobEventChan chan *worker.JobEvent
	jobPlanTable map[string]*executor.JobSchedulePlan
	jobExecutor  *executor.JobExecutor
}

var Scheduler scheduler

func InitScheduler() {
	jobExecutor, _ := executor.NewJobExecutor()
	Scheduler = scheduler{
		jobEventChan: make(chan *worker.JobEvent, 1000),
		jobPlanTable: make(map[string]*executor.JobSchedulePlan),
		jobExecutor:  jobExecutor,
	}

	go Scheduler.eventLoop()
}

func (s *scheduler) UpdateJob(event *worker.JobEvent) {
	s.jobEventChan <- event
}

// update the status to local
func (s *scheduler) updateJob(j *worker.JobEvent) error {
	switch j.EventType {
	case worker.PutEvent:
		parsedCronExpr, err := cronexpr.Parse(j.CronExpr)
		if err != nil {
			return err
		}
		jobSchedulePlan := executor.JobSchedulePlan{
			Job:      j.Job,
			Expr:     parsedCronExpr,
			NextTime: time.Time{},
		}
		s.jobPlanTable[j.Name] = &jobSchedulePlan
	case worker.DeleteEvent:
		delete(s.jobPlanTable, j.Name)
	case worker.KillEvent:
		s.killJob(j.Name)
	}

	return nil
}

func (s *scheduler) killJob(jobName string) {
	v, ok := s.jobExecutor.DoingJob.Load(jobName)
	if ok {
		if f, ok := v.(context.CancelFunc); ok {
			f()
			s.jobExecutor.DoingJob.Delete(jobName)
			log.Printf("[%s] has been killed\n", jobName)
		}
	}
}

// trySchedule
//if jobPlanTable is Empty return one second
//if jobPlanTable is not Empty return the time to nearly job
func (s *scheduler) trySchedule() time.Duration {
	if len(s.jobPlanTable) == 0 {
		return time.Second
	}
	now := time.Now()
	var nearTime *time.Time
	for _, plan := range s.jobPlanTable {
		if plan.NextTime.Before(now) || plan.NextTime == now {
			//send the job to the executor
			s.jobExecutor.AddJob(plan)
			plan.NextTime = plan.Expr.Next(now)
		}
		// find the nearest time of the execute job and let it be the nearTime
		if nearTime == nil || plan.NextTime.Before(*nearTime) {
			nearTime = &plan.NextTime
		}
	}
	now = time.Now()
	return nearTime.Sub(now)
}

func (s *scheduler) eventLoop() {
	timer := time.NewTimer(time.Second)
	for {
		select {
		// update jobs to local
		case jobEvent := <-s.jobEventChan:
			err := s.updateJob(jobEvent)
			if err != nil {
				fmt.Println(err)
				continue
			}
		// run the job and then have a rest
		case <-timer.C:
			after := s.trySchedule()
			timer.Reset(after)
		}
	}
}
