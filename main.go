package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"unilab-backend/OsServer"
	"unilab-backend/apis"
	"unilab-backend/auth"
	"unilab-backend/database"
	"unilab-backend/judger"
	"unilab-backend/logging"
	"unilab-backend/middleware"
	"unilab-backend/osgrade"
	"unilab-backend/setting"
	"unilab-backend/taskqueue"
	"unilab-backend/webhook"

	"github.com/gin-gonic/gin"
)

func testJudger() {
	config := judger.TestConfig{
		TestID:      1,
		TimeLimit:   1000,       // ms
		MemoryLimit: 512 * 1024, // KB
		TestCaseNum: 3,
		Language:    "go",
	}
	result := judger.LaunchTest(config, "../../testcase", "../../program")
	logging.Info(result)
}

func testOs() {
	database.GetGradeDetailsById(2018011302)
}

func initRouter() *gin.Engine {
	gin.DefaultWriter = io.MultiWriter(logging.GetWriter(setting.RuntimeRootDir+"logs/access.log"), os.Stdout)
	gin.DefaultErrorWriter = io.MultiWriter(logging.GetWriter(setting.RuntimeRootDir+"logs/access_error.log"), os.Stderr)
	router := gin.New()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	// custom logger
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))
	router.Use(gin.Recovery())

	// cross-origin routes
	router.Use(middleware.Cors())

	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello World!")
	})
	// Routes
	// see http status: https://pkg.go.dev/net/http#pkg-constants
	router.GET("/login", auth.UserLoginHandler)
	router.GET("/callback", auth.GitLabCallBackHandler)

	router.GET("/Os/FetchGrade", osgrade.FetchOsGrade)
	router.POST("/webhook/os", webhook.OsWebhookHandler)
	studentApis := router.Group("/student")
	studentApis.Use(middleware.JWTMiddleWare(), middleware.PriorityMiddleware(database.UserStudent))
	{
		studentApis.GET("/fetch-my-course", apis.FetchUserCoursesHandler)
		studentApis.GET("/fetch-announcement", apis.FetchCourseAnnouncementsHandler)
		studentApis.GET("/fetch-course-name", apis.GetCourseNameHandler)
		studentApis.GET("/fetch-announcement-detail", apis.GetAnnouncementHandler)
		studentApis.GET("/fetch-question", apis.FetchCourseQuestionsHandler)
		studentApis.GET("/fetch-question-detail", apis.FetchQuestionHandler)
		studentApis.POST("/fetch-question-appendix", apis.FetchQuestionAppendix)
		studentApis.POST("/submit-code", apis.SubmitCodeHandler)
		studentApis.GET("/fetch-assignment", apis.GetAssignmentsInfoHandler)
		studentApis.GET("/fetch-all-testids", apis.FetchAllSubmitsStatus)
		studentApis.POST("/update-tests", apis.UpdateTestDetails)
		studentApis.GET("/fetch-submit-detail", apis.GetSubmitDetail)
		studentApis.POST("/Os/Grade", osgrade.GetOsGradeHandler)
		studentApis.GET("/Os/BranchGrade", osgrade.GetOsBranchGradeHandler)
		studentApis.POST("/submit-task", taskqueue.TaskSubmitHandler)
	}
	teacherApis := router.Group("/teacher")
	teacherApis.Use(middleware.JWTMiddleWare(), middleware.PriorityMiddleware(database.UserTeacher))
	{
		teacherApis.POST("/create-course", apis.CreateCourseHandler)
		teacherApis.GET("/fetch-all-user", apis.GetAllUsersHandler)
		teacherApis.GET("/fetch-all-teacher", apis.GetAllTeachersHandler)
		teacherApis.POST("/create-announcement", apis.CreateAnnouncementHandler)
		teacherApis.POST("/create-question", apis.CreateQuestionHandler)
		teacherApis.POST("/create-assignment", apis.CreateAssignmentHandler)
	}
	adminApis := router.Group("/admin")
	adminApis.Use(middleware.JWTMiddleWare(), middleware.PriorityMiddleware(database.UserAdmin))
	{
		adminApis.POST("/add-teachers", apis.AddTeachersHandler)
	}

	return router
}

func main() {
	logging.Info("Start Golang App")
	database.InitDB()
	taskqueue.InitYTaskServer()
	go OsServer.InitConsumer()

	gin.SetMode(setting.RunMode)
	router := initRouter()
	endPoint := fmt.Sprintf(":%d", setting.HttpPort)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        router,
		ReadTimeout:    setting.ReadTimeout,
		WriteTimeout:   setting.WriteTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	// gracefully shutdown
	// Initializing the server in a goroutine so that it won't block the graceful shutdown handling below
	go func() {
		if err := server.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			logging.Info("listen: ", err)
		} else {
			logging.Info("server listening: ", endPoint)
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logging.Info("Shutting down server...")
	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logging.Fatal("Server forced to shutdown:", err)
	}
	logging.Info("Server exiting...")
	// testOs()
	// testJudger()
}
