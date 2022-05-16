package apis

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unilab-backend/database"
	"unilab-backend/logging"
	"unilab-backend/setting"
	"unilab-backend/utils"

	"github.com/gin-gonic/gin"
)

var LanguageSupported = []string{"c", "c++", "c++11", "c++14", "c++17", "c++20", "java8", "java11", "java14", "java17", "python2", "python3", "rust", "go", "js"}

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
	if len(testcase) == 0 || len(testcase)%2 != 0 {
		ErrorResponse(c, INVALID_PARAMS, "Test Cases Should be MORE than ONE PAIR of (*.in & *.ans) files!")
		return
	}
	var testCaseNum uint32 = uint32(len(testcase) / 2)
	// check testcase postfix
	for i := 1; i <= int(testCaseNum); i++ {
		var validIn = false
		var validAns = false
		for _, file := range testcase {
			filename := filepath.Base(file.Filename)
			if !validIn && filename == fmt.Sprintf("%d.in", i) {
				validIn = true
			}
			if !validAns && filename == fmt.Sprintf("%d.ans", i) {
				validAns = true
			}
		}
		if !validIn || !validAns {
			ErrorResponse(c, INVALID_PARAMS, "Test Cases Should be MORE than ONE PAIR of (*.in & *.ans) files!")
			return
		}
	}
	postform.Language = strings.ToLower(postform.Language)
	// check language
	if !utils.ArrayContainsString(postform.Language, LanguageSupported) {
		ErrorResponse(c, INVALID_PARAMS, "Language ("+postform.Language+") Not Supported.")
		return
	}
	// replace blank in question title to '-'
	postform.Title = strings.ReplaceAll(postform.Title, " ", "-")
	// insert into database
	questionID, err := database.CreateQuestion(postform, c.MustGet("user_id").(uint32), testCaseNum)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	// save file to disk
	questionBasePath := setting.QuestionRootDir + strconv.FormatUint(uint64(questionID), 10) + "_" + postform.Title + "/"
	err = os.MkdirAll(questionBasePath, 0777)
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// description
	files := form.File["description"]
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		logging.Info("receive file: ", filename)
		dst := questionBasePath + "description.md"
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			os.RemoveAll(questionBasePath)
			return
		}
	}
	appendix := form.File["appendix"]
	for _, file := range appendix {
		filename := filepath.Base(file.Filename)
		logging.Info("receive file: ", filename)
		dst := questionBasePath + "appendix.zip"
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			os.RemoveAll(questionBasePath)
			return
		}
	}
	for _, file := range testcase {
		filename := filepath.Base(file.Filename)
		logging.Info("receive file: ", filename)
		dst := questionBasePath + filename
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			os.RemoveAll(questionBasePath)
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

func FetchCourseQuestionsHandler(c *gin.Context) {
	courseID, err := utils.StringToUint32(c.Query("courseid"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// get offset
	offset, err := utils.StringToUint32(c.Query("offset"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// get limit
	limit, err := utils.StringToUint32(c.Query("limit"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	if !database.CheckCourseAccessPermission(courseID, c.MustGet("user_id").(uint32)) {
		NoAccessResponse(c, "You are not allowed to access this course.")
		return
	}
	totalNum, err := database.GetQuestionTotalNumByCourseID(courseID)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	questions, err := database.GetQuestionsByCourseID(courseID, c.MustGet("user_id").(uint32), limit, offset)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": map[string]interface{}{
			"result": map[string]interface{}{
				"questions": questions,
				"totalNum":  totalNum,
			},
		},
	})
}

// fetch question detail
func FetchQuestionHandler(c *gin.Context) {
	questionID, err := utils.StringToUint32(c.Query("questionid"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	question, err := database.GetQuestionDetailByID(questionID, c.MustGet("user_id").(uint32))
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
	questionID, err := utils.StringToUint32(c.PostForm("questionid"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// check permission
	if !database.CheckQuestionAccessPermission(questionID, c.MustGet("user_id").(uint32)) {
		ErrorResponse(c, ERROR, "You are not allowed to access this question!")
		return
	}
	// get appendix path
	appendixPath, err := database.GetQuestionAppendixPath(questionID)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	if !utils.FileExists(appendixPath) {
		logging.Info("appendixPath: ", appendixPath, " DO NOT Exists.")
		ErrorResponse(c, ERROR, "File DO NOT Exists.")
		return
	}
	f, err := os.Open(appendixPath)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	defer f.Close()
	err = database.UserAccessAppendix(questionID, c.MustGet("user_id").(uint32))
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	// 以流方式下载文件，可以匹配所有类型的文件
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=appendix.zip")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "max-age=864000")
	c.File(appendixPath)
}
