package main

import (
	"fmt"
	"net/http"

	"unilab-backend/apis"
	"unilab-backend/auth"
	"unilab-backend/database"
	"unilab-backend/judger"
	"unilab-backend/logging"
	"unilab-backend/middleware"
	"unilab-backend/os"
	"unilab-backend/setting"

	"github.com/gin-gonic/gin"
)


func testJudger() {
	config := judger.TestConfig{
		"title",
		1000,
		262144,
		3,
		[]uint32{10, 10, 10},
	}
	result := judger.LaunchTest(config, "../testcase", "../program")
	logging.Info(result)
}


func initRouter() *gin.Engine {
	router := gin.New()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.Use(gin.Logger())
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
		studentApis.GET("/Os/Grade", os.GetOsGradeHandler)
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
	// database.PreinitDBTestData()
	gin.SetMode(setting.RunMode)
	router := initRouter()
	endPoint := fmt.Sprintf(":%d", setting.HttpPort)
	maxHeaderBytes := 1 << 20
	
	server := &http.Server{
		Addr: endPoint,
		Handler: router,
		ReadTimeout: setting.ReadTimeout,
		WriteTimeout: setting.WriteTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}
	logging.Info("start http server listening ", endPoint)
	server.ListenAndServe()

	// testJudger()
}