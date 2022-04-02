package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"unilab-backend/apis"
	"unilab-backend/auth"
	"unilab-backend/database"
	"unilab-backend/middleware"
	"unilab-backend/os"
)


func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	database.InitDB()
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
	// router.POST("/login", apis.UserLoginHandler)
	router.GET("/login", auth.UserLoginHandler)
	
	router.GET("/callback", auth.GitLabCallBackHandler)
	router.GET("/Os/Grade", os.GetOsGradeHandler)
	studentApis := router.Group("/student")
	studentApis.Use(middleware.JWTMiddleWare(), middleware.PriorityMiddleware(database.UserStudent))
	{
		studentApis.GET("/fetch-my-course", apis.FetchUserCoursesHandler)
		studentApis.GET("/fetch-announcement", apis.FetchCourseAnnouncementsHandler)
		studentApis.GET("/fetch-course-name", apis.GetCourseNameHandler)
		studentApis.GET("/fetch-annocement", apis.GetAnnouncementHandler)
		studentApis.GET("/fetch-question", apis.FetchCourseQuestionsHandler)
	}
	teacherApis := router.Group("/teacher")
	teacherApis.Use(middleware.JWTMiddleWare(), middleware.PriorityMiddleware(database.UserTeacher))
	{
		teacherApis.POST("/create-course", apis.CreateCourseHandler)
		teacherApis.GET("/fetch-all-user", apis.GetAllUsersHandler)
		teacherApis.POST("/create-announcement", apis.CreateAnnouncementHandler)
		teacherApis.POST("/create-question", apis.CreateQuestionHandler)
	}
	adminApis := router.Group("/admin")
	adminApis.Use(middleware.JWTMiddleWare(), middleware.PriorityMiddleware(database.UserAdmin))
	{

	}
	router.Run(":1323")
}
