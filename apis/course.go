package apis

import (
	"log"
	"net/http"
	"strconv"
	"unilab-backend/database"

	"github.com/gin-gonic/gin"
)


func GetCourseNameHandler(c *gin.Context) {
	courseid_str := c.Query("courseid")
	if courseid_str == "" {
		ErrorResponse(c, ERROR, "there's not Course ID in query params.")
		return
	}
	courseID, err := strconv.ParseUint(courseid_str, 10, 32)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	if !database.CheckCourseAccessPermission(uint32(courseID), c.MustGet("user_id").(uint32)) {
		NoAccessResponse(c, "You are not allowed to access this course.")
		return
	}
	courseName, err := database.GetCourseByID(uint32(courseID))
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["result"] = courseName
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}

func CreateCourseHandler(c *gin.Context) {
	var courseForm database.CreateCourseForm
	data := make(map[string]interface{})
	code := SUCCESS
	if err := c.ShouldBind(&courseForm); err != nil {
		code = INVALID_PARAMS
		log.Println(err)
		data["err"] = err.Error()
	} else {
		log.Println(courseForm)
		if len(courseForm.Teachers) == 0 || len(courseForm.Students) == 0 {
			data["err"] = "Course's teachers or students are none."
			log.Println(data["err"])
		} else {
			// insert into database
			err := database.CreateNewCourse(courseForm)
			if err != nil {
				data["err"] = err.Error()
				log.Println(err)
			}
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
