package logger

import (
	"context"
	"dbsmonitor/crontab/master/config"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

type LogDocument struct {
	Err          string `bson:"err"`
	OutPut       string `bson:"out_put"`
	PlanTime     int64  `bson:"plan_time"`
	ScheduleTime int64  `bson:"schedule_time"`
	StartTime    int64  `bson:"start_time"`
	EndTime      int64  `bson:"end_time"`
}
type PagedLog struct {
	CurrPage  int64         `json:"curr_page"`
	TotalPage int64         `json:"total_page"`
	Log       []LogDocument `json:"log"`
}
type logger struct {
	collection *mongo.Collection
	pageSize   int64
}

var Logger *logger

func InitLogger() error {
	clientOpts := options.Client().ApplyURI(config.Config.MongoUri)
	client, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		fmt.Println(1, err)
	}
	database := client.Database("cron")
	collection := database.Collection("logs")
	Logger = &logger{collection, config.Config.LogPageSize}
	return nil
}

func (l logger) GetLogs(name string, page int64) *PagedLog {
	var skip int64
	skip = (page - 1) * (l.pageSize)
	pageSize := l.pageSize
	opts := options.FindOptions{
		Limit: &pageSize,
		Skip:  &skip,
		Sort:  bson.D{{"plan_time", 1}},
	}
	cursor, err := l.collection.Find(context.TODO(), bson.D{{"name", name}}, &opts)
	defer cursor.Close(context.TODO())
	if err != nil {
		log.Println("find log error ", err)
		return nil
	}
	var result []LogDocument
	for cursor.Next(context.TODO()) {
		var jobLog LogDocument
		if err = cursor.Decode(&jobLog); err != nil {
			continue
		}
		result = append(result, jobLog)
	}
	count, err := l.collection.CountDocuments(context.TODO(), bson.D{{"name", name}})
	if err != nil {
		return nil
	}
	paged := &PagedLog{
		CurrPage:  page,
		TotalPage: count / pageSize,
		Log:       result,
	}
	return paged
}
