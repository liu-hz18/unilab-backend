package OsServer

import(
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"context"
	"github.com/gojuukaze/YTask/v2"
)

type Task struct {
	UserID   uint32 
	CourseType string 
    CourseName string 
    Extra map[string]string
}

func TaskGrade(task Task) {
	fmt.Println(task)
}

func InitConsumer(){
	broker := ytask.Broker.NewRedisBroker("127.0.0.1","6379","",0,0)
	backend := ytask.Backend.NewRedisBackend("127.0.0.1","6379","",0,0)
	logger := ytask.Logger.NewYTaskLogger()

	ser := ytask.Server.NewServer(
        ytask.Config.Broker(&broker),
        ytask.Config.Backend(&backend),
        ytask.Config.Logger(logger),
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

func main(){
	InitConsumer()
}
