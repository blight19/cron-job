package worker

import "dbsmonitor/crontab/job"

type EventType int

const (
	PutEvent EventType = iota
	DeleteEvent
	KillEvent
)

type JobEvent struct {
	*job.Job
	EventType
}
