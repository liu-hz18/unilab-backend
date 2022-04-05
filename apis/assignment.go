package apis

import (
	"log"
	"net/http"
	"time"
	"unilab-backend/database"
	"unilab-backend/utils"

	"github.com/gin-gonic/gin"
)


func CreateAssignmentHandler(c *gin.Context) {
	var postform database.CreateAssignmentForm
	if err := c.ShouldBind(&postform); err != nil {
		log.Println(err)
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	log.Println(postform)
	if !database.CheckCourseAccessPermission(postform.CourseID, c.MustGet("user_id").(uint32)) {
		NoAccessResponse(c, "You are not allowed to access this course.")
		return
	}
	var form = database.CreateAssignmentInfo{
		Title: postform.Title,
		Description: postform.Description,
		BeginTime: utils.StringToTime(postform.BeginTime),
		DueTime: utils.StringToTime(postform.DueTime),
		CourseID: postform.CourseID,
		QuestionIDs: postform.QuestionIDs,
	}
	// check params
	if form.BeginTime.After(form.DueTime) || form.BeginTime.Equal(form.DueTime) {
		ErrorResponse(c, INVALID_PARAMS, "参数错误：截止时间 应该在 发布时间 之后。")
		return
	}
	if time.Now().After(form.DueTime) || time.Now().Equal(form.DueTime) {
		ErrorResponse(c, INVALID_PARAMS, "参数错误：截止时间 应该在 当前时间 之后。")
		return
	}
	if len(form.QuestionIDs) <= 0 {
		ErrorResponse(c, INVALID_PARAMS, "参数错误：发布作业时请至少包含一个题目。")
		return
	}
	assignmentID, err := database.CreateAssignment(form)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["result"] = assignmentID
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg": MsgFlags[SUCCESS],
		"data": data,
	})
}


func GetAssignmentsInfoHandler(c *gin.Context) {
	courseid_str := c.Query("courseid")
	if courseid_str == "" {
		ErrorResponse(c, ERROR, "there's not Course ID in query params.")
		return
	}
	courseID, err := utils.StringToUint32(courseid_str)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	userID := c.MustGet("user_id").(uint32)
	if !database.CheckCourseAccessPermission(courseID, userID) {
		NoAccessResponse(c, "You are not allowed to access this course.")
		return
	}
	info, err := database.GetAssignemntInfo(courseID, userID)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["result"] = info
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg": MsgFlags[SUCCESS],
		"data": data,
	})
}

