package api

import (
	"dbsmonitor/crontab/job"
	"dbsmonitor/crontab/master"
	"dbsmonitor/crontab/master/api/auth"
	"dbsmonitor/crontab/master/config"
	"dbsmonitor/crontab/master/logger"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func ApiResponse(code int, message string, data interface{}) Response {

	return Response{
		Code: code,
		Msg:  message,
		Data: data,
	}
}

func handleJobSave(ctx *gin.Context) {
	var jobItem job.Job
	postJob := ctx.PostForm("job")
	err := json.Unmarshal([]byte(postJob), &jobItem)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiResponse(1, "invalid data", nil))
		return
	}
	oldJob, err := master.JobMgr.SaveJob(&jobItem)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiResponse(2, err.Error(), nil))
		return
	}
	if oldJob != nil {
		ctx.JSON(http.StatusOK, ApiResponse(0, "success", map[string]job.Job{"old-job": *oldJob}))
	} else {
		ctx.JSON(http.StatusOK, ApiResponse(0, "success", nil))
	}
}

func handleJobDelete(ctx *gin.Context) {
	name := ctx.PostForm("name")
	oldJob, err := master.JobMgr.DeleteJob(name)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiResponse(1, err.Error(), nil))
		return
	}
	if oldJob != nil {
		ctx.JSON(http.StatusOK, ApiResponse(0, "success", map[string]job.Job{"old-job": *oldJob}))
	} else {
		ctx.JSON(http.StatusOK, ApiResponse(0, "success", nil))
	}
}

func handleJobList(ctx *gin.Context) {
	jobItem, err := master.JobMgr.ListJob()
	if err != nil {
		ctx.JSON(http.StatusOK, ApiResponse(-1, err.Error(), nil))
		return
	}
	ctx.JSON(http.StatusOK, ApiResponse(0, "success", jobItem))
}

// handleJobKill kill a job with job name
func handleJobKill(ctx *gin.Context) {
	jobName := ctx.PostForm("name")
	err := master.JobMgr.KillJob(jobName)
	if err != nil {
		ctx.JSON(http.StatusOK, ApiResponse(1, err.Error(), nil))
		return
	}
	ctx.JSON(http.StatusOK, ApiResponse(0, "", nil))
}
func handleJobWorker(ctx *gin.Context) {
	workers, err := master.JobMgr.GetWorkers()
	if err != nil {
		ctx.JSON(http.StatusOK, ApiResponse(1, err.Error(), nil))
		return
	}
	ctx.JSON(http.StatusOK, ApiResponse(0, "success", workers))
}
func handleJobLog(ctx *gin.Context) {
	jobName := ctx.PostForm("name")
	pageString := ctx.PostForm("page")
	page, err := strconv.Atoi(pageString)
	if err != nil {
		page = 1
	}
	result := logger.Logger.GetLogs(jobName, int64(page))
	ctx.JSON(http.StatusOK, ApiResponse(0, "success", result))
}

func handleLogin(ctx *gin.Context) {
	username := ctx.PostForm("username")
	password := ctx.PostForm("password")
	if auth.Verify(username, password) {
		token, err := auth.GenToken(username)
		if err != nil {
			ctx.JSON(http.StatusOK, ApiResponse(-1, "generate token failed", nil))
			return
		}
		ctx.JSON(http.StatusOK, ApiResponse(0, "success", token))
	} else {
		ctx.JSON(http.StatusOK, ApiResponse(-1, "auth failed", nil))
	}
}

func handleUserInfo(ctx *gin.Context) {

}

func InitApiServer() error {
	r := gin.Default()
	r.Use(CORSMiddleware())
	//r.StaticFS("/home", http.Dir("webroot"))
	r.POST("/login", handleLogin)
	authorized := r.Group("/", JWTAuth())
	authorized.POST("/job/save", handleJobSave)
	authorized.POST("/job/delete", handleJobDelete)
	authorized.POST("/job/list", handleJobList)
	authorized.POST("/job/kill", handleJobKill)
	authorized.POST("/job/log", handleJobLog)
	authorized.POST("/job/workers", handleJobWorker)
	s := &http.Server{
		Addr:           ":8080",
		Handler:        r,
		ReadTimeout:    time.Duration(config.Config.ApiReadTimeout) * time.Second,
		WriteTimeout:   time.Duration(config.Config.ApiWriteTimeout) * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	err := s.ListenAndServe()
	return err
}

func CORSMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*") // 可将将 * 替换为指定的域名
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		ctx.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization")
		ctx.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if ctx.Request.Method == http.MethodOptions {
			ctx.AbortWithStatus(200)
		} else {
			ctx.Next()
		}
	}
}
