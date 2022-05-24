package job

import (
	"encoding/json"
	"strings"
)

const (
	SAVEDIR   = "/cron/jobs/"
	KILLDIR   = "/cron/kill/"
	LOCKDIR   = "/cron/locks/"
	HEALTHDIR = "/cron/health/"
)

type Job struct {
	Name     string `json:"name"`     //任务名称
	Command  string `json:"command"`  //任务执行命令
	CronExpr string `json:"cronExpr"` //cron表达式
}

func Unpack(data []byte) (*Job, error) {
	var job Job
	err := json.Unmarshal(data, &job)
	if err != nil {
		return nil, err
	}
	return &job, err
}

func GetDeleteName(fullName []byte) string {
	return strings.TrimPrefix(string(fullName), SAVEDIR)
}

func GetKillName(fullName []byte) string {
	return strings.TrimPrefix(string(fullName), KILLDIR)
}

func GetWorkerName(fullName []byte) string {
	return strings.TrimPrefix(string(fullName), HEALTHDIR)
}
