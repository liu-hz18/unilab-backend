package apis

import (
	"net/http"
	"strconv"
	"unilab-backend/database"
	"unilab-backend/logging"
	"unilab-backend/utils"

	"github.com/gin-gonic/gin"
)

func GetCourseNameHandler(c *gin.Context) {
	courseIDStr := c.Query("courseid")
	if courseIDStr == "" {
		ErrorResponse(c, ERROR, "there's not Course ID in query params.")
		return
	}
	courseID, err := strconv.ParseUint(courseIDStr, 10, 32)
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
		logging.Info(err)
		data["err"] = err.Error()
	} else {
		logging.Info(courseForm)
		if len(courseForm.Teachers) == 0 || len(courseForm.Students) == 0 {
			data["err"] = "Course's teachers or students are none."
			logging.Info(data["err"])
		} else {
			// insert into database
			err := database.CreateNewCourse(courseForm)
			if err != nil {
				data["err"] = err.Error()
				logging.Info(err)
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
	userIDStr := c.Query("id")
	data["result"] = []database.Course{}
	if userIDStr == "" {
		code = INVALID_PARAMS
		data["err"] = "invalid parameters"
		c.JSON(http.StatusOK, gin.H{
			"code": code,
			"msg":  MsgFlags[code],
			"data": data,
		})
		return
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
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

func UserAccessCourse(c *gin.Context) {
	courseID, err := utils.StringToUint32(c.Query("courseid"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	err = database.UserAccessCourse(courseID, c.MustGet("user_id").(uint32))
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["result"] = "ok"
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}
