package main

import (
	"log"
	"time"

	"github.com/gorhill/cronexpr"
)

type CronJob struct {
	expr     *cronexpr.Expression
	nextTime time.Time
}

func main() {
	sheduleTable := make(map[string]*CronJob)
	cronExpr1 := "*/5 * * * * * *"
	if expr1, err := cronexpr.Parse(cronExpr1); err != nil {

		log.Fatalln("Error:invalid expr", cronExpr1)
	} else {
		sheduleTable["job1"] = &CronJob{
			expr:     expr1,
			nextTime: expr1.Next(time.Now()),
		}
	}

	cronExpr2 := "*/3 * * * * * *"
	if expr2, err := cronexpr.Parse(cronExpr2); err != nil {

		log.Fatalln("Error:invalid expr", cronExpr1)
	} else {
		sheduleTable["job1"] = &CronJob{
			expr:     expr2,
			nextTime: expr2.Next(time.Now()),
		}
	}

}
