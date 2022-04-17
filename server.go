package main

import (
	"log"

	"unilab-backend/apis"
	"unilab-backend/auth"
	"unilab-backend/database"
	"unilab-backend/judger"
	"unilab-backend/middleware"
	"unilab-backend/os"
	"unilab-backend/taskqueue"
	// "unilab-backend/os-server"
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
	log.Println(result)
}


func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	database.InitDB()
	taskqueue.InitYTaskServer()
	// OsServer.InitConsumer()
	// database.PreinitDBTestData()

	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20 // 8 MiB
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	gin.SetMode(gin.DebugMode)
	// gin.SetMode(gin.ReleaseMode)

	// cross-origin routes
	router.Use(middleware.Cors())


	// Routes
	// see http status: https://pkg.go.dev/net/http#pkg-constants
	router.GET("/login", auth.UserLoginHandler)
	router.GET("/callback", auth.GitLabCallBackHandler)
	// router.POST("/submit-task",taskqueue.TaskSubmitHandler)
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
		studentApis.POST("/submit-task",taskqueue.TaskSubmitHandler)
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

	router.Run(":1323")

	// testJudger()
}
