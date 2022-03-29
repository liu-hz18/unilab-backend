package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/middleware"
)

// NIfIDMtN?8In


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

	// cross-source routes
	router.Use(middleware.Cors())


	// Routes
	router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello world")
	})
	// router.POST("/codeupload", func(c *gin.Context){
	// 	// value
	// 	name := c.PostForm("name")
	// 	id := c.PostForm("id")
	// 	// Multipart form
	// 	form, err := c.MultipartForm()
	// 	if err != nil {
	// 		c.String(http.StatusBadRequest, "get form error: %s", err.Error())
	// 		return
	// 	}
	// 	files := form.File["file"]
	// 	os.MkdirAll(FILE_SAVE_ROOT_DIR + id + "/" + "prj1/", 777)
	// 	for _, file := range files {
	// 		filename := filepath.Base(file.Filename)
	// 		log.Printf("receive file %s", filename)
	// 		dst := FILE_SAVE_ROOT_DIR + id + "/" + "prj1/" + filename
	// 		if err := c.SaveUploadedFile(file, dst); err != nil {
	// 			c.String(http.StatusBadRequest, "save file error: %s", err.Error())
	// 			return
	// 		}
	// 	}
	// 	c.String(http.StatusOK, fmt.Sprintf("%d files uploaded successfully with fields name=%s and id=%s.", len(files), name, id))
	// })

	router.POST("/login", apis.UserLoginHandler)
	userApis := router.Group("/user")
	userApis.Use(middleware.JWTMiddleWare())
	{
		// userApis.POST("/login", )
	}

	router.Run(":1323")
}

