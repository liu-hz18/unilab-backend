package apis

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"unilab-backend/database"

	"github.com/gin-gonic/gin"
)


func CreateAnnouncementHandler(c *gin.Context) {
	// validate params
	var postform database.CreateAnnouncementForm
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
	// get coursename
	course_name, err := database.GetCourseByID(postform.CourseID)
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
	base_path := database.COURSE_DATA_DIR + strconv.FormatUint(uint64(postform.CourseID), 10) + "_" + course_name + "/announcements/"
	err = os.MkdirAll(base_path, 777)
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
	announcement_id, err := database.CreateNewAnnouncement(postform)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		log.Println("receive file: ", filename)
		dst := base_path + strconv.FormatUint(uint64(announcement_id), 10) + "_announcement.md"
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			return
		}
	}
	data := make(map[string]interface{})
	code := SUCCESS
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg": MsgFlags[code],
		"data": data,
	})
}


func FetchCourseAnnouncementsHandler(c *gin.Context) {
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
	announcements, err := database.GetAnnouncementsByCourseID(uint32(courseID))
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
	annoid_str := c.Query("annoid")
	if annoid_str == "" {
		ErrorResponse(c, ERROR, "there's not Course ID or Announcement ID in query params.")
		return
	}
	annoID, err := strconv.ParseUint(annoid_str, 10, 32)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	if !database.CheckAnnouncementAccessPermission(uint32(annoID), c.MustGet("user_id").(uint32)) {
		NoAccessResponse(c, "You are not allowed to access this announcement.")
		return
	}
	info, err := database.GetAnnouncementInfo(uint32(annoID))
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
