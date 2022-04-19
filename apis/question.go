package apis

import (
	"fmt"
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


func CreateQuestionHandler(c *gin.Context) {
	var postform database.CreateQuestionForm
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
	// get file
	form, err := c.MultipartForm()
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// get testcase file
	testcase := form.File["testcase"]
	if len(testcase) <= 0 || len(testcase) % 2 != 0 {
		ErrorResponse(c, INVALID_PARAMS, "Test Cases Should be MORE than ONE PAIR of (*.in & *.ans) files!")
		return
	}
	var testCaseNum uint32 = uint32(len(testcase) / 2)
	// check testcase postfix
	for i := 1; i <= int(testCaseNum); i++ {
		var valid_in bool = false
		var valid_ans bool = false
		for _, file := range testcase {
			filename := filepath.Base(file.Filename)
			if (!valid_in && filename == fmt.Sprintf("%d.in", i)) {
				valid_in = true
			}
			if (!valid_ans && filename == fmt.Sprintf("%d.ans", i)) {
				valid_ans = true
			}
		}
		if !valid_in || !valid_ans {
			ErrorResponse(c, INVALID_PARAMS, "Test Cases Should be MORE than ONE PAIR of (*.in & *.ans) files!")
			return
		}
	}
	// insert into database
	question_id, err := database.CreateQuestion(postform, c.MustGet("user_id").(uint32), testCaseNum)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	// save file to disk
	question_base_path := setting.QuestionRootDir + strconv.FormatUint(uint64(question_id), 10) + "_" + postform.Title + "/"
	err = os.MkdirAll(question_base_path, 777)
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// description
	files := form.File["description"]
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		logging.Info("receive file: ", filename)
		dst := question_base_path + "description.md"
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			os.RemoveAll(question_base_path)
			return
		}
	}
	appendix := form.File["appendix"]
	for _, file := range appendix {
		filename := filepath.Base(file.Filename)
		logging.Info("receive file: ", filename)
		dst := question_base_path + "appendix.zip"
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			os.RemoveAll(question_base_path)
			return
		}
	}
	for _, file := range testcase {
		filename := filepath.Base(file.Filename)
		logging.Info("receive file: ", filename)
		dst := question_base_path + filename
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			os.RemoveAll(question_base_path)
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

// fetch question detail
func FetchQuestionHandler(c *gin.Context) {
	question_id_str := c.Query("questionid")
	if question_id_str == "" {
		ErrorResponse(c, ERROR, "there's not Question ID in query params.")
		return
	}
	questionID_uint64, err := strconv.ParseUint(question_id_str, 10, 32)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	questionID := uint32(questionID_uint64)
	question, err := database.GetQuestionDetailByID(questionID)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["result"] = question
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}

// fetch question appendix
func FetchQuestionAppendix(c *gin.Context) {
	appendix_path := c.PostForm("path")
	if appendix_path == "" {
		ErrorResponse(c, ERROR, "there's not Appendix Path in query params.")
		return
	}
	if !utils.FileExists(appendix_path) {
		ErrorResponse(c, ERROR, "File DO NOT Exists.")
		return
	}
	f, err := os.Open(appendix_path)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	defer f.Close()
	// 以流方式下载文件，可以匹配所有类型的文件
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=appendix.zip")
  	c.Header("Content-Transfer-Encoding", "binary")
  	c.Header("Cache-Control", "no-cache")
	c.File(appendix_path)
}