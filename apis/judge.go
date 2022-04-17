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
	questionName, testCaseNum, err := database.GetQuestionTitleAndTestCaseNumByID(postform.QuestionID)
	if err != nil {
		ErrorResponse(c, ERROR, err.Error())
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
	files := form.File["file"]
	for _, file := range files {
		filename := filepath.Base(file.Filename)
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
		"msg": MsgFlags[code],
		"data": data,
	})
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

