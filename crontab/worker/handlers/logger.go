package handlers

import (
	"context"
	"dbsmonitor/crontab/worker/config"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

type LogDocument struct {
	Name         string `bson:"name"`
	Command      string `bson:"command"`
	Err          string `bson:"err"`
	OutPut       string `bson:"out_put"`
	PlanTime     int64  `bson:"plan_time"`
	ScheduleTime int64  `bson:"schedule_time"`
	StartTime    int64  `bson:"start_time"`
	EndTime      int64  `bson:"end_time"`
}
type Logger struct {
	logCollection *mongo.Collection
	logBatch      []interface{}
	logChan       chan *LogDocument
	batchSize     int
	batchChan     chan []interface{}
	failed        sync.Map
}

func NewLogger() (*Logger, error) {
	clientOpts := options.Client().ApplyURI(config.Config.MongoUri)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		fmt.Println(1, err)
	}
	database := client.Database("cron")
	collection := database.Collection("logs")
	logger := Logger{
		logCollection: collection,
		logChan:       make(chan *LogDocument, 100),
		logBatch:      make([]interface{}, 0),
		batchSize:     config.Config.MongoBatch,
		batchChan:     make(chan []interface{}, 10),
	}
	go logger.run()
	return &logger, nil
}
func (l *Logger) AddLog(document *LogDocument) {
	if len(l.logChan) == cap(l.logChan) {
		if cap(l.logChan) < 3201 {
			l.logChan = make(chan *LogDocument, cap(l.logChan)*2)
		} else {
			l.logChan = make(chan *LogDocument, cap(l.logChan))
		}
	}
	l.logChan <- document
}
func (l *Logger) run() {
	for {
		select {
		case log := <-l.logChan:
			l.logBatch = append(l.logBatch, log)
			if len(l.logBatch) >= l.batchSize {
				l.batchChan <- l.logBatch
				l.empty()
			}
		case <-time.Tick(time.Second * 3):
			if len(l.logBatch) > 0 {
				l.batchChan <- l.logBatch
				l.empty()
			}
		case data := <-l.batchChan:
			go l.batchCommit(data)
		}
	}
}

func (l *Logger) batchCommit(data []interface{}) {
	_, err := l.logCollection.InsertMany(context.TODO(), data)
	if err != nil {
		fmt.Println(err)
		times, ok := l.failed.Load(data)
		if ok {
			l.failed.Store(data, times.(int)+1)
		} else {
			l.failed.Store(data, 1)
		}
		if times.(int) < 5 {
			l.batchChan <- data
		}
		time.Sleep(time.Second * 5)

		return
	}
}

func (l *Logger) empty() {
	l.logBatch = l.logBatch[:0]
}

//func main() {
//	logger, _ :=New()
//	count:=0
//	for j := 0; j <100; j++ {
//		for i := 0; i < 53; i++ {
//			log:=LogDocument{
//				Name:         "x"+strconv.Itoa(j),
//				Command:      strconv.Itoa(i),
//				Err:          strconv.Itoa(j*10+i),
//				OutPut:       strings.Repeat("afdsfaefaewfawefawef",8000),
//				PlanTime:     0,
//				ScheduleTime: 0,
//				StartTime:    0,
//				EndTime:      0,
//			}
//			logger.AddLog(&log)
//			count++
//
//		}
//		if j%7==0{
//			time.Sleep(time.Second*time.Duration(rand.Intn(4)))
//		}
//	}
//	fmt.Println(count)
//	select {
//
//	}
//
//}
