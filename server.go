package main

import (
	"os"
	// "io"
	"log"
    "fmt"
	"path/filepath"

	"net/http"
	"github.com/gin-gonic/gin"

	// "database/sql"
	// "github.com/jmoiron/sqlx"
)

func main() {
	router := gin.Default()
	// Set a lower memory limit for multipart forms (default is 32 MiB)
	router.MaxMultipartMemory = 8 << 20  // 8 MiB
	
	// Routes
	router.GET("/", func(c *gin.Context){
		c.String(http.StatusOK, "Hello world")
	})
	router.POST("/codeupload", func(c *gin.Context){
		// value
		name := c.PostForm("name")
		id := c.PostForm("id")
		// Multipart form
		form, err := c.MultipartForm()
		if err != nil {
			c.String(http.StatusBadRequest, "get form error: %s", err.Error())
			return
		}
		files := form.File["file"]
		os.MkdirAll(id + "/" + "prj1/", 777)
		for _, file := range files {
			filename := filepath.Base(file.Filename)
			log.Printf("receive file %s", filename)
			dst := id + "/" + "prj1/" + filename
			if err := c.SaveUploadedFile(file, dst); err != nil {
				c.String(http.StatusBadRequest, "save file error: %s", err.Error())
				return
			}
		}
		c.String(http.StatusOK, fmt.Sprintf("%d files uploaded successfully with fields name=%s and id=%s.", len(files), name, id))
	})


	router.Run(":1323")
}

