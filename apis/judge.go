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

func SubmitCodeHandler(c *gin.Context) {
	var postform database.SubmitCodeForm
	if err := c.ShouldBind(&postform); err != nil {
		log.Println(err)
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	log.Println(postform)
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
	questionName, err := database.GetQuestionTitleByID(postform.QuestionID)
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
	submit_base_path := database.USER_DATA_DIR + "/" + strconv.FormatUint(uint64(userid), 10) + "/" + strconv.FormatUint(uint64(postform.QuestionID), 10) + "_" + questionName + "/" + strconv.FormatUint(uint64(count+1), 10) + "_test/"
	err = os.MkdirAll(submit_base_path, 777)
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	files := form.File["file"]
	for _, file := range files {
		filename := filepath.Base(file.Filename)
		log.Println("SubmitCodeHandler() receive file: ", filename)
		dst := submit_base_path + filename
		if err := c.SaveUploadedFile(file, dst); err != nil {
			ErrorResponse(c, ERROR, err.Error())
			return
		}
	}
	// insert into test database
	testID, err := database.CreateSubmitCode(postform, userid, submit_base_path)
	data := make(map[string]interface{})
	data["result"] = testID
	code := SUCCESS
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg": MsgFlags[code],
		"data": data,
	})
}

// TODO: fetch all submit infos
func FetchAllSubmitsStatus(c *gin.Context) {

}

// TODO: update one specific submit detail info


