package OsServer

import(
	// "fmt"
	"os"
	// "bytes"
	"os/signal"
	"syscall"
	"context"
	"net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"github.com/gojuukaze/YTask/v2"
)

type Task struct {
	UserID   uint32 
	CourseType string 
    CourseName string 
    Extra map[string]string
}

type Result struct{
	Details []GradeRecord `json:"gradeDetails"`
}

type GradeRecord struct{
	Id uint32 `json:"Id"`
	Branch_name string `json:"Branch_name"`
	Tests []Test `json:"Tests"`
	Outputs []Output `json:"Outputs"`
}

type Test struct{
	//member definition
	Id int `json:"Id"`
	Name string `json:"Name"`
	Passed bool `json:"Passed"`
	Score int `json:"Score"`
}

type Output struct{
	Id int `json:"Id"`
	Type string `json:"Type"`
	Alert_class string `json:"Alert_class"`
	Message string `json:"Message"`
	Content string `json:"Content"`
	Expand bool `json:"Expand"`
}

func TaskGrade(task Task) []GradeRecord{
	// client := &http.Client{}
	// bytesData,_ := json.Marshal(task)
	// req,_ := http.NewRequest("POST","http://localhost:1323/student/Os/Grade",bytes.NewReader(bytesData))
	// response,_ := client.Do(req)
	// body,_ := ioutil.ReadAll(response.Body)
	// var result map[string]interface{}
	// _=json.Unmarshal(body,&result)
	// fmt.Println("result:")
	// fmt.Println(result)
	// return result

	client := &http.Client{}
	params := url.Values{}
	Url,err := url.Parse("http://localhost:1323/Os/Grade")
	if err != nil{
		return nil
	}
	params.Set("id","2018011302")
	Url.RawQuery = params.Encode()
	req,err :=http.NewRequest("GET",Url.String(),nil)
	if err != nil{
		return nil
	}
	// req.Header.Add("Authorization","Bearer " + access_token)
	response,err :=client.Do(req)
	if err != nil{
		return nil
	}
	body,err := ioutil.ReadAll(response.Body)
	if err != nil{
		return nil
	}
	var result Result
	// var result []map[string]interface{}
	err = json.Unmarshal(body,&result)
	if err != nil{
		return nil
	}
	return result.Details
}

func InitConsumer(){
	broker := ytask.Broker.NewRedisBroker("127.0.0.1","6379","",0,0)
	backend := ytask.Backend.NewRedisBackend("127.0.0.1","6379","",0,0)
	// logger := ytask.Logger.NewYTaskLogger()

	ser := ytask.Server.NewServer(
        ytask.Config.Broker(&broker),
        ytask.Config.Backend(&backend),
        // ytask.Config.Logger(logger),
        ytask.Config.Debug(true),
        ytask.Config.StatusExpires(60*5),
        ytask.Config.ResultExpires(60*5),
    )

	ser.Add("os-server","grade",TaskGrade)
	ser.Run("os-server",3)
	quit := make(chan os.Signal, 1)

    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    ser.Shutdown(context.Background())
}

// func main(){
// 	InitConsumer()
// }
