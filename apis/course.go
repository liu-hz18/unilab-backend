package apis

import (
	"log"
	"net/http"
	"strconv"
	"unilab-backend/database"

	"github.com/gin-gonic/gin"
)


func CreateCourseHandler(c *gin.Context) {
	var courseForm database.CreateCourseForm
	data := make(map[string]interface{})
	code := SUCCESS
	if err := c.ShouldBind(&courseForm); err != nil {
		code = INVALID_PARAMS
		log.Println(err)
		data["err"] = err.Error()
	} else {
		// insert into database
		log.Println(courseForm)
		err := database.CreateNewCourse(courseForm)
		if err != nil {
			data["err"] = err.Error()
			log.Println(err)
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  MsgFlags[code],
		"data": data,
	})
}

func FetchUserCoursesHandler(c *gin.Context) {
	data := make(map[string]interface{})
	code := SUCCESS
	userid_str := c.Query("id")
	data["result"] = []database.Course{}
	if userid_str == "" {
		code = INVALID_PARAMS
		data["err"] = "invalid parameters"
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  MsgFlags[code],
			"data": data,
		})
		return
	}
	userID, err := strconv.ParseUint(userid_str, 10, 32)
	if err != nil {
		code = INVALID_PARAMS
		data["err"] = err.Error()
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  MsgFlags[code],
			"data": data,
		})
		return
	}
	courses, err := database.GetUserCourses(uint32(userID))
	if err != nil {
		code = ERROR
		data["err"] = err.Error()
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  MsgFlags[code],
			"data": data,
		})
		return
	}
	data["result"] = courses
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  MsgFlags[code],
		"data": data,
	})
}
