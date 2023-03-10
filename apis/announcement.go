package apis

import (
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"unilab-backend/database"
	"unilab-backend/logging"
	"unilab-backend/setting"
	"unilab-backend/utils"

	"github.com/gin-gonic/gin"
)

func CreateAnnouncementHandler(c *gin.Context) {
	// validate params
	var postform database.CreateAnnouncementForm
	if err := c.ShouldBind(&postform); err != nil {
		logging.Info(err)
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	logging.Info(postform)
	if !database.CheckCourseAccessPermission(postform.CourseID, c.MustGet("user_id").(uint32)) {
		NoAccessResponse(c, "You are not allowed to access this course.")
		return
	}
	// get coursename
	courseName, err := database.GetCourseByID(postform.CourseID)
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// get file
	form, err := c.MultipartForm()
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// save file to disk
	basePath := setting.CourseRootDir + strconv.FormatUint(uint64(postform.CourseID), 10) + "_" + courseName + "/announcements/"
	err = os.MkdirAll(basePath, 0777)
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	files := form.File["file"]
	if len(files) != 1 {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// insert into database
	announcementID, err := database.CreateNewAnnouncement(postform)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		logging.Info("receive file: ", filename)
		dst := basePath + strconv.FormatUint(uint64(announcementID), 10) + "_announcement.md"
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			return
		}
	}
	data := make(map[string]interface{})
	code := SUCCESS
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  MsgFlags[code],
		"data": data,
	})
}

func FetchCourseAnnouncementsHandler(c *gin.Context) {
	courseID, err := utils.StringToUint32(c.Query("courseid"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	if !database.CheckCourseAccessPermission(courseID, c.MustGet("user_id").(uint32)) {
		NoAccessResponse(c, "You are not allowed to access this course.")
		return
	}
	announcements, err := database.GetAnnouncementsByCourseID(courseID)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["result"] = announcements
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}

func GetAnnouncementHandler(c *gin.Context) {
	annoID, err := utils.StringToUint32(c.Query("annoid"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	if !database.CheckAnnouncementAccessPermission(annoID, c.MustGet("user_id").(uint32)) {
		NoAccessResponse(c, "You are not allowed to access this announcement.")
		return
	}
	info, err := database.GetAnnouncementInfo(annoID, c.MustGet("user_id").(uint32))
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["result"] = info
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}
