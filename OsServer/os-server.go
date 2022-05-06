package OsServer

import(
	// "fmt"
	"os"
	"bytes"
	"unilab-backend/setting"
	"os/signal"
	"syscall"
	"context"
	"net/http"
	// "net/url"
	"io/ioutil"
	"encoding/json"
	"github.com/gojuukaze/YTask/v2"
)

type Task struct {
	UserID     uint32 `json:"userid" form:"userid" uri:"userid" binding:"required"`
	CourseType string `json:"coursetype" form:"coursetype" uri:"coursetype" binding:"required"`
	CourseName string `json:"coursename" form:"coursename" uri:"coursename" binding:"required"`
	Token      string
	Extra      map[string]string `json:"extra" form:"extra" uri:"extra" binding:"required"`
}

type Result struct {
	Test_status string        `json:"test_status"`
	Details     []GradeRecord `json:"gradeDetails"`
}

type GradeRecord struct {
	Id        uint32   `json:"Id"`
	Test_name string   `json:"Test_name"`
	Tests     []Test   `json:"Tests"`
	Outputs   []Output `json:"Outputs"`
}

type Test struct {
	//member definition
	Id   int    `json:"Id"`
	Name string `json:"Name"`
	// Passed bool `json:"Passed"`
	Score       int `json:"Score"`
	Total_score int `json:"Total_Score"`
}

type Output struct {
	Id   int    `json:"Id"`
	Type string `json:"Type"`
	// Alert_class string `json:"Alert_class"`
	Message string `json:"Message"`
	Content string `json:"Content"`
	// Expand bool `json:"Expand"`
}

func TaskGrade(task Task) []GradeRecord {
	client := &http.Client{}
	bytesData, _ := json.Marshal(&task)
	req, err := http.NewRequest("POST", "http://localhost:1323/student/Os/Grade", bytes.NewReader(bytesData))
	if err != nil {
		return nil
	}
	req.Header.Add("Authorization", task.Token)
	response, err := client.Do(req)
	if err != nil {
		return nil
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil
	}
	var result Result
	// var result []map[string]interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil
	}
	// client := &http.Client{}
	// params := url.Values{}
	// Url,err := url.Parse("http://localhost:1323/student/Os/Grade")
	// if err != nil{
	// 	return nil
	// }
	// params.Set("id","2018011302")
	// Url.RawQuery = params.Encode()
	// req,err :=http.NewRequest("GET",Url.String(),nil)
	// if err != nil{
	// 	return nil
	// }
	// req.Header.Add("Authorization",task.Token)
	// response,err :=client.Do(req)
	// if err != nil{
	// 	return nil
	// }
	// body,err := ioutil.ReadAll(response.Body)
	// if err != nil{
	// 	return nil
	// }
	// var result Result
	// // var result []map[string]interface{}
	// err = json.Unmarshal(body,&result)
	// if err != nil{
	// 	return nil
	// }
	return result.Details
}

func InitConsumer() {
	broker := ytask.Broker.NewRedisBroker(setting.RedisHost, setting.RedisPort, setting.RedisPassword, 0, 0)
	backend := ytask.Backend.NewRedisBackend(setting.RedisHost, setting.RedisPort, setting.RedisPassword, 0, 0)
	// logger := ytask.Logger.NewYTaskLogger()

	ser := ytask.Server.NewServer(
		ytask.Config.Broker(&broker),
		ytask.Config.Backend(&backend),
		// ytask.Config.Logger(logger),
		ytask.Config.Debug(true),
		ytask.Config.StatusExpires(60*5),
		ytask.Config.ResultExpires(60*5),
	)

	ser.Add("os-server", "grade", TaskGrade)
	ser.Run("os-server", 3)
	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	ser.Shutdown(context.Background())
}
