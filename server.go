package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/middleware"
)


func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	database.InitDB()
	database.PreinitDBTestData()


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
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello world")
	})

	// see http status: https://pkg.go.dev/net/http#pkg-constants
	router.POST("/login", apis.UserLoginHandler)
	studentApis := router.Group("/student")
	studentApis.Use(middleware.JWTMiddleWare(), middleware.PriorityMiddleware(database.UserStudent))
	{
		studentApis.GET("/fetch-my-course", apis.FetchUserCoursesHandler)
	}
	teacherApis := router.Group("/teacher")
	teacherApis.Use(middleware.JWTMiddleWare(), middleware.PriorityMiddleware(database.UserTeacher))
	{
		teacherApis.POST("/create-course", apis.CreateCourseHandler)
		teacherApis.GET("/fetch-all-user", apis.GetAllUsersHandler)
	}
	adminApis := router.Group("/admin")
	adminApis.Use(middleware.JWTMiddleWare(), middleware.PriorityMiddleware(database.UserAdmin))
	{

	}
	router.Run(":1323")
}
