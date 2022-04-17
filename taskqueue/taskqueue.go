package taskqueue

import (
    "log"
    "fmt"
    "unilab-backend/apis"
	// "unilab-backend/database"
	"github.com/gin-gonic/gin"
    // "unilab-backend/apis"
	"github.com/gojuukaze/YTask/v2"
    "github.com/gojuukaze/YTask/v2/server"
    "time"
)

type Task struct {
	UserID   uint32 `json:"userid" form:"userid" uri:"userid" binding:"required"`
	CourseType string `json:"coursetype" form:"coursetype" uri:"coursetype" binding:"required"`
    CourseName string `json:"coursename" form:"coursename" uri:"coursename" binding:"required"`
    Extra map[string]string `json:"extra" form:"extra" uri:"extra" binding:"required"`
}

var client server.Client

func InitYTaskServer(){
	broker := ytask.Broker.NewRedisBroker("127.0.0.1", "6379", "", 0, 5)
	backend := ytask.Backend.NewRedisBackend("127.0.0.1", "6379", "", 0, 5)
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

func TaskSubmitHandler(c *gin.Context){
    var task Task
	if err := c.ShouldBind(&task); err != nil {
		log.Println(err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	// fmt.Println(task)
    taskId,_ := client.Send("os-server","grade",task)
    result,_ := client.GetResult(taskId, 2*time.Second, 300*time.Millisecond)
    fmt.Println(result)
}