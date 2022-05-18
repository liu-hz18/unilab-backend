package compiler

import (
	"fmt"
	"strconv"
	"strings"
	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/gitlabapi"

	"github.com/gin-gonic/gin"
)

type Task struct {
	UserID     uint32            `json:"userid" form:"userid" uri:"userid" binding:"required"`
	CourseType string            `json:"coursetype" form:"coursetype" uri:"coursetype" binding:"required"`
	CourseName string            `json:"coursename" form:"coursename" uri:"coursename" binding:"required"`
	Extra      map[string]string `json:"extra" form:"extra" uri:"extra" binding:"required"`
}

const OK_PREFIX = "OK "
const FAIL_PREFIX = "FAIL "
const ERR_PREFIX = "ERR "
const RUN_FIRST_LINE_PREFIX = "[32;1m$ PROJ_PATH=..  STEP_UNTIL=`cat step-until.txt` USE_PARALLEL=false minidecaf-tests/check.sh[0;m"

var outputs []database.Output

func add_to_outputs(cur_output_lines []string, last_output_type string, n_ok int, n_fail int, n_err int) {
	raw_text := strings.Join(cur_output_lines, "\n")
	var message string
	if last_output_type == "Run" {
		message = "Test Passed: " + strconv.Itoa(n_ok) + " / " + strconv.Itoa(n_ok+n_fail+n_err)
	} else {
		message = ""
	}
	outputs = append(outputs, database.Output{len(outputs) + 1, last_output_type, message, raw_text})
}

func Grade(ci_output string) ([]database.Test, []database.Output) {
	lines := strings.Split(ci_output, "\n")
	n_ok := 0
	n_fail := 0
	n_err := 0
	tests := []database.Test{}

	last_output_type := "Prepare"
	cur_output_type := "Prepare"
	outputs = []database.Output{}
	cur_output_lines := []string{}

	for _, line := range lines {
		if strings.HasPrefix(line, RUN_FIRST_LINE_PREFIX) {
			cur_output_type = "Run"
		}
		if cur_output_type != last_output_type {
			add_to_outputs(cur_output_lines, last_output_type, n_ok, n_fail, n_err)
			cur_output_lines = []string{}
		}
		last_output_type = cur_output_type
		cur_output_lines = append(cur_output_lines, line)
		if strings.HasPrefix(line, OK_PREFIX) {
			n_ok++
			tests = append(tests, database.Test{n_ok + n_fail + n_err, line[len(OK_PREFIX):], 1, 1})
		}
		if strings.HasPrefix(line, FAIL_PREFIX) {
			n_fail++
			tests = append(tests, database.Test{n_ok + n_fail + n_err, line[len(FAIL_PREFIX):], 0, 1})
		}
		if strings.HasPrefix(line, ERR_PREFIX) {
			n_err++
			tests = append(tests, database.Test{n_ok + n_fail + n_err, line[len(ERR_PREFIX):], 0, 1})
		}
	}
	if cur_output_type == "Run" {
		add_to_outputs(cur_output_lines, last_output_type, n_ok, n_fail, n_err)
	} else {
		add_to_outputs(cur_output_lines, last_output_type, n_ok, n_fail, n_err)
	}
	return tests, outputs
}

func FetchCompilerGrade(c *gin.Context) {
	id := c.Query("id")
	accessToken, err := database.GetUserAccessToken(id)
	if err != nil {
		apis.ErrorResponse(c, apis.ERROR, err.Error())
		return
	}
	user_git_tsinghua_id, err := database.GetUserGitTsinghuaID(id)
	if err != nil {
		apis.ErrorResponse(c, apis.ERROR, err.Error())
		return
	}
	traces := gitlabapi.GetProjectTraces("minidecaf-"+id, id, accessToken)
	// fmt.Println(traces)
	// userId, _ := strconv.ParseUint(user_git_tsinghua_id, 10, 32)
	for trace := range traces {
		tests, _ := Grade(traces[trace])
		fmt.Println(trace)
		fmt.Println(tests)
		// database.CreateGradeRecord(uint32(userId),trace,tests,outputs,"passed")
	}
}
