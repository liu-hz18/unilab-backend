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

func CreateQuestionHandler(c *gin.Context) {
	var postform database.CreateQuestionForm
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
	// insert into database
	question_id, err := database.CreateQuestion(postform, c.MustGet("user_id").(uint32))
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	// save file to disk
	base_path := database.COURSE_DATA_DIR + strconv.FormatUint(uint64(postform.CourseID), 10) + "_" + course_name + "/questions/"
	question_base_path := base_path + strconv.FormatUint(uint64(question_id), 10) + "_" + postform.Title + "/"
	err = os.MkdirAll(question_base_path, 777)
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// description
	files := form.File["description"]
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		log.Println("receive file: ", filename)
		dst := question_base_path + "description.md"
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			return
		}
	}
	appendix := form.File["appendix"]
	for _, file := range appendix {
		filename := filepath.Base(file.Filename)
		log.Println("receive file: ", filename)
		dst := question_base_path + "appendix.zip"
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			return
		}
	}
	testcase := form.File["testcase"]
	for _, file := range testcase {
		filename := filepath.Base(file.Filename)
		log.Println("receive file: ", filename)
		dst := question_base_path + "testcase.zip"
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


func FetchCourseQuestionsHandler(c *gin.Context) {
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
	questions, err := database.GetQuestionsByCourseID(uint32(courseID))
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["result"] = questions
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}
