package apis

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"unilab-backend/database"
	"unilab-backend/judger"
	"unilab-backend/logging"
	"unilab-backend/setting"
	"unilab-backend/utils"

	"github.com/gin-gonic/gin"
)

func SubmitCodeHandler(c *gin.Context) {
	var postform database.SubmitCodeForm
	if err := c.ShouldBind(&postform); err != nil {
		logging.Info(err)
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	logging.Info(postform)
	var userid uint32 = c.MustGet("user_id").(uint32)
	if !database.CheckCourseAccessPermission(postform.CourseID, userid) {
		NoAccessResponse(c, "You are not allowed to access this course.")
		return
	}
	form, err := c.MultipartForm()
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	// get question name
	questionName, testCaseNum, questionLanguage, err := database.GetQuestionTitleAndTestCaseNumAndLanguageByID(postform.QuestionID)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	if questionLanguage != postform.Language {
		ErrorResponse(c, INVALID_PARAMS, fmt.Sprintf("Submit Language %s does not match required language %s.", postform.Language, questionLanguage))
		return
	}
	// check submit codes
	files := form.File["file"]
	var containsSource = false
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		logging.Info("SubmitCodeHandler() receive file: ", filename)
		if filename == judger.JudgerConfig[questionLanguage].NeedFile {
			containsSource = true
		}
	}
	if !containsSource {
		ErrorResponse(c, INVALID_PARAMS, fmt.Sprintf("Submit Files does not match contain %s.", judger.JudgerConfig[questionLanguage].NeedFile))
		return
	}
	// read last submit
	prevDir, err := database.GetLastSubmitDir(postform.CourseID, c.MustGet("user_id").(uint32), postform.QuestionID)
	if err != nil {
		logging.Info(err)
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	// insert into test database
	testID, err := database.CreateSubmitRecord(postform, userid, testCaseNum)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	// save upload code to disk
	submitBasePath := setting.UserRootDir + strconv.FormatUint(uint64(userid), 10) + "/" + strconv.FormatUint(uint64(postform.QuestionID), 10) + "_" + questionName + "/" + strconv.FormatUint(uint64(testID), 10) + "_test/"
	err = os.MkdirAll(submitBasePath, 0777)
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		if filename == "main.java" {
			filename = "Main.java"
		}
		dst := submitBasePath + filename
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			return
		}
	}
	data := make(map[string]interface{})
	data["result"] = testID
	code := SUCCESS
	// run test
	database.RunTest(testID, postform.QuestionID, submitBasePath, prevDir, postform.Language)
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  MsgFlags[code],
		"data": data,
	})
}

// fetch all submit test ids
func FetchAllSubmitsIDs(c *gin.Context) {
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
	var userID uint32 = c.MustGet("user_id").(uint32)
	if !database.CheckCourseAccessPermission(courseID, userID) {
		NoAccessResponse(c, "You are not allowed to access this course.")
		return
	}
	totalNum, err := database.GetUserSubmitsTestCount(courseID, userID)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	results, err := database.GetUserSubmitTestIDs(courseID, userID, limit, offset)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": map[string]interface{}{
			"result": map[string]interface{}{
				"testids":  results,
				"totalNum": totalNum,
			},
		},
	})
}

// update one specific submit detail info
func UpdateTestDetails(c *gin.Context) {
	var testIDs []uint32
	if err := c.ShouldBind(&testIDs); err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	details := database.GetTestDetailsByIDs(testIDs)
	data := make(map[string]interface{})
	data["result"] = details
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}

func GetSubmitDetail(c *gin.Context) {
	testID, err := utils.StringToUint32(c.Query("testid"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	result := database.GetSubmitDetail(testID)
	data := make(map[string]interface{})
	data["result"] = result
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}
