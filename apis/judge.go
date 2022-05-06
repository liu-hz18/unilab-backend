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
	var contains_source = false
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		if filename == judger.JudgerConfig[questionLanguage].NeedFile {
			contains_source = true
		}
	}
	if !contains_source {
		ErrorResponse(c, INVALID_PARAMS, fmt.Sprintf("Submit Files does not match contain %s.", judger.JudgerConfig[questionLanguage].NeedFile))
		return
	}
	// get test count
	count, err := database.GetQuestionSubmitCounts(postform.QuestionID, userid)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
		return
	}
	// save upload code to disk
	submit_base_path := setting.UserRootDir + "/" + strconv.FormatUint(uint64(userid), 10) + "/" + strconv.FormatUint(uint64(postform.QuestionID), 10) + "_" + questionName + "/" + strconv.FormatUint(uint64(count+1), 10) + "_test/"
	err = os.MkdirAll(submit_base_path, 777)
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		if filepath.Ext(file.Filename) == ".java" {
			filename = "Main.java"
		}
		logging.Info("SubmitCodeHandler() receive file: ", filename)
		dst := submit_base_path + filename
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			return
		}
	}
	// insert into test database
	testID, err := database.CreateSubmitRecord(postform, userid, submit_base_path, testCaseNum)
	data := make(map[string]interface{})
	data["result"] = testID
	code := SUCCESS
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  MsgFlags[code],
		"data": data,
	})
	// run test
	logging.Info("begin launtch test: ", testID)
	database.RunTest(testID)
}

// fetch all submit test ids
func FetchAllSubmitsStatus(c *gin.Context) {
	courseID, err := utils.StringToUint32(c.Query("courseid"))
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	var userID uint32 = c.MustGet("user_id").(uint32)
	if !database.CheckCourseAccessPermission(courseID, userID) {
		NoAccessResponse(c, "You are not allowed to access this course.")
		return
	}
	results, err := database.GetUserSubmitTests(courseID, userID)
	data := make(map[string]interface{})
	data["result"] = results
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
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
