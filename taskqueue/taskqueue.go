package taskqueue

import (
	"log"
	// "fmt"
	"net/http"
	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/logging"
	"unilab-backend/setting"

	"github.com/gin-gonic/gin"

	"time"

	ytask "github.com/gojuukaze/YTask/v2"
	"github.com/gojuukaze/YTask/v2/server"
)

type Task struct {
	UserID     uint32 `json:"userid" form:"userid" uri:"userid" binding:"required"`
	CourseType string `json:"coursetype" form:"coursetype" uri:"coursetype" binding:"required"`
	CourseName string `json:"coursename" form:"coursename" uri:"coursename" binding:"required"`
	Token      string
	Extra      map[string]string `json:"extra" form:"extra" uri:"extra" binding:"required"`
}

var client server.Client

func InitYTaskServer() {
	broker := ytask.Broker.NewRedisBroker(setting.RedisHost, setting.RedisPort, setting.RedisPassword, 0, 5)
	backend := ytask.Backend.NewRedisBackend(setting.RedisHost, setting.RedisPort, setting.RedisPassword, 0, 5)
	logger := ytask.Logger.NewYTaskLogger()
	ser := ytask.Server.NewServer(
		ytask.Config.Broker(&broker),
		ytask.Config.Backend(&backend),
		ytask.Config.Logger(logger),
		ytask.Config.Debug(true),
		ytask.Config.StatusExpires(60*5),
		ytask.Config.ResultExpires(60*5),
	)
	client = ser.GetClient()
}

func TaskSubmitHandler(c *gin.Context) {
	var task Task
	if err := c.ShouldBind(&task); err != nil {
		log.Println(err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	task.Token = c.Request.Header.Get("Authorization")
	// fmt.Println(task)
	taskID, _ := client.Send("os-server", "grade", task)
	result, _ := client.GetResult(taskID, 2*time.Second, 300*time.Millisecond)
	var gradeDetails []database.GradeRecord
	if result.IsSuccess() {
		err := result.Get(0, &gradeDetails)
		if err != nil {
			logging.Error(err.Error())
			apis.ErrorResponse(c, apis.ERROR, err.Error())
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"gradeDetails": gradeDetails,
	})
}
