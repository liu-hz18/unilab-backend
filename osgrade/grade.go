package osgrade

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/gitlabapi"
	"unilab-backend/logging"

	"github.com/gin-gonic/gin"
)

type Task struct {
	UserID     uint32            `json:"userid" form:"userid" uri:"userid" binding:"required"`
	CourseType string            `json:"coursetype" form:"coursetype" uri:"coursetype" binding:"required"`
	CourseName string            `json:"coursename" form:"coursename" uri:"coursename" binding:"required"`
	Extra      map[string]string `json:"extra" form:"extra" uri:"extra" binding:"required"`
}

const PASS_PREFIX = "[92m[PASS][0m "
const FAIL_PREFIX = "[91m[FAIL][0m "
const TEST_PASSED_PREFIX = "Test passed: "
const COMPILE_FIRST_LINE_PREFIX = "make -C user all CHAPTER="

var COMPILE_FAILED_PATTERN = regexp.MustCompile(`/make\[\d+\]: .* Error .*\n/`)

const RUSTSBI_FIRST_LINE_PREFIX = "[rustsbi] RustSBI version "

var CHECK_FIRST_LINE_PATTERN = `python3 check/ch`
var outputs []database.Output

func addToOutputs(curOutputLines []string, lastOutputType string, curHasFail bool, curPassNum int, curFailNum int) {
	rawText := strings.Join(curOutputLines, "\n")
	var compileFailureMessage string
	hasCompileFailed := COMPILE_FAILED_PATTERN.FindAllString(rawText, -1) != nil
	if !hasCompileFailed {
		compileFailureMessage = ""
	} else {
		compileFailureMessage = COMPILE_FAILED_PATTERN.FindAllString(rawText, -1)[0]
	}
	// var alertClass string
	// if lastOutputType=="Check"{
	// 	if cur_has_fail {
	// 		alertClass="danger"
	// 	}else{
	// 		alertClass="success"
	// 	}
	// }else if lastOutputType=="Compile" && hasCompileFailed {
	// 	alertClass="danger"
	// }else{
	// 	alertClass="info"
	// }
	var message string
	switch {
	case lastOutputType == "Check":
		message = "Test Passed: " + strconv.Itoa(curPassNum) + " / " + strconv.Itoa(curPassNum+curFailNum)
	case hasCompileFailed && lastOutputType == "Compile":
		message = compileFailureMessage
	default:
		message = ""
	}
	outputs = append(outputs, database.Output{
		Id:      len(outputs) + 1,
		Type:    lastOutputType,
		Message: message,
		Content: rawText,
	})
}

func Grade(ciOutput string) ([]database.Test, []database.Output) {
	lines := strings.Split(ciOutput, "\n")
	nPass := 0
	nFail := 0
	testPassedN1 := 0
	testPassedN2 := 0
	// tests:=list.New()
	tests := []database.Test{}

	lastOutputType := "CI Output"
	var curOutputType string
	outputs = []database.Output{}
	curOutputLines := []string{}
	curHasFailed := false
	curPassNum := 0
	curFailNum := 0
	lastLine := ""

	for _, line := range lines {
		ischeck, _ := regexp.MatchString(CHECK_FIRST_LINE_PATTERN, line)
		switch {
		case strings.HasPrefix(line, RUSTSBI_FIRST_LINE_PREFIX):
			curOutputType = "Run"
		case ischeck:
			curOutputType = "Check"
		case strings.HasPrefix(line, COMPILE_FIRST_LINE_PREFIX):
			curOutputType = "Compile"
		case strings.HasPrefix(lastLine, TEST_PASSED_PREFIX):
			curOutputType = "CI Output"
		default:
			curOutputType = "CI Output"
		}
		if curOutputType != lastOutputType {
			addToOutputs(curOutputLines, lastOutputType, curHasFailed, curPassNum, curFailNum)
			curOutputLines = []string{}
			curHasFailed = false
			curPassNum = 0
			curFailNum = 0
		}
		lastLine = line
		lastOutputType = curOutputType
		curOutputLines = append(curOutputLines, line)
		if strings.HasPrefix(line, PASS_PREFIX) {
			nPass++
			curPassNum++
			tests = append(tests, database.Test{
				Id:          nPass + nFail,
				Name:        line[len(PASS_PREFIX):],
				Score:       1,
				Total_score: 1,
			})
		}
		if strings.HasPrefix(line, FAIL_PREFIX) {
			nFail++
			curFailNum++
			tests = append(tests, database.Test{
				Id:          nPass + nFail,
				Name:        line[len(FAIL_PREFIX):],
				Score:       0,
				Total_score: 1,
			})
			curHasFailed = true
		}
		if strings.HasPrefix(line, TEST_PASSED_PREFIX) {
			tmp := strings.Split(strings.Split(line, ": ")[1], "/")
			val1, err1 := strconv.Atoi(tmp[0])
			val2, err2 := strconv.Atoi(tmp[1])
			if err1 != nil || err2 != nil {
				fmt.Println("error")
			}
			testPassedN1 += val1
			testPassedN2 += val2
		}
	}
	addToOutputs(curOutputLines, lastOutputType, curHasFailed, curPassNum, curFailNum)
	return tests, outputs
}

func GetOsGradeHandler(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		logging.Info(err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	id := task.UserID
	logging.Info(id)
	logging.Info("start grade")
	// userId,_ := strconv.ParseUint(id, 10, 32)
	gradeDetails, err := database.GetGradeDetailsByID(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"test_status":  "FAIL",
			"gradeDetails": gradeDetails,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"test_status":  "SUCCESS",
			"gradeDetails": gradeDetails,
		})
	}
}

func GetOsBranchGradeHandler(c *gin.Context) {
	id := c.Query("id")
	branch := c.Query("branch")
	userID, _ := strconv.ParseUint(id, 10, 32)
	gradeRecord, _ := database.GetGradeDetailByBranch(uint32(userID), branch)
	c.JSON(http.StatusOK, gin.H{
		"tests":   gradeRecord.Tests,
		"outputs": gradeRecord.Outputs,
	})
}

func FetchOsGrade(c *gin.Context) {
	id := c.Query("id")
	accessToken, err := database.GetUserAccessToken(id)
	if err != nil {
		apis.ErrorResponse(c, apis.ERROR, err.Error())
		return
	}
	userGitTsinghuaID, err := database.GetUserGitTsinghuaID(id)
	if err != nil {
		apis.ErrorResponse(c, apis.ERROR, err.Error())
		return
	}
	traces := gitlabapi.GetProjectTraces("labs-"+id, id, accessToken)
	userID, _ := strconv.ParseUint(userGitTsinghuaID, 10, 32)
	for trace := range traces {
		tests, outputs := Grade(traces[trace])
		err := database.CreateGradeRecord(uint32(userID), trace, tests, outputs, "passed")
		if err != nil {
			apis.ErrorResponse(c, apis.ERROR, err.Error())
			return
		}
	}
	// need response ?
}
