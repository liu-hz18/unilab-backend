package os

import(
	"fmt"
	"strconv"
	"strings"
	"regexp"
	"unilab-backend/gitlab_api"
	"github.com/gin-gonic/gin"
	"net/http"
	// "encoding/json"
	// "container/list"
)

const PASS_PREFIX = "[92m[PASS][0m "
const FAIL_PREFIX = "[91m[FAIL][0m "
const TEST_PASSED_PREFIX = "Test passed: "
const COMPILE_FIRST_LINE_PREFIX ="make -C user all CHAPTER="
var COMPILE_FAILED_PATTERN = regexp.MustCompile(`/make\[\d+\]: .* Error .*\n/`)
const RUSTSBI_FIRST_LINE_PREFIX = "[rustsbi] RustSBI version "
var CHECK_FIRST_LINE_PATTERN = `python3 check/ch`
// var outputs=list.New()
var outputs []output

type Test struct{
	//member definition
	Id int
	Name string
	Passed bool
	Score int
}

type output struct{
	Id int
	Type string
	Alert_class string
	Message string
	Content string
	Expand bool
}

func add_to_outputs(cur_output_lines []string,last_output_type string,cur_has_fail bool,cur_n_pass int,cur_n_fail int){
	raw_text := strings.Join(cur_output_lines,"\n")
	var has_compile_failed bool
	var compile_failure_message string
	if COMPILE_FAILED_PATTERN.FindAllString(raw_text,-1)!=nil{
		has_compile_failed = true
	}else{
		has_compile_failed = false
	}
	if has_compile_failed == false{
		compile_failure_message=""
	}else{
		compile_failure_message=COMPILE_FAILED_PATTERN.FindAllString(raw_text,-1)[0]
	}
	var alert_class string
	if last_output_type=="Check"{
		if cur_has_fail {
			alert_class="danger"
		}else{
			alert_class="success"
		}
	}else if last_output_type=="Compile" && has_compile_failed{
		alert_class="danger"
	}else{
		alert_class="info"
	}
	var message string
	if last_output_type == "Check"{
		message = "Test Passed: "+strconv.Itoa(cur_n_pass)+" / "+strconv.Itoa(cur_n_pass + cur_n_fail)
	}else if last_output_type== "Compile" && has_compile_failed{
		message=compile_failure_message
	}else{
		message=""
	}
	outputs=append(outputs,output{len(outputs)+1,last_output_type,alert_class,message,raw_text,cur_has_fail || has_compile_failed})
}


func Grade(ci_output string) ([]Test,[]output){
	
	lines:=strings.Split(ci_output,"\n")
	n_pass:=0
	n_fail:=0
	test_passed_n1:=0
	test_passed_n2:=0
	// tests:=list.New()
	tests := []Test {}

	last_output_type:="CI Output"
	cur_output_type:="CI Output"
	outputs = []output {}
	cur_output_lines:=[]string {}
	cur_has_fail:=false
	cur_n_pass:=0
	cur_n_fail:=0
	last_line:=""

	for _,line:=range lines{
		ischeck,_:=regexp.MatchString(CHECK_FIRST_LINE_PATTERN,line)
		if strings.HasPrefix(line,RUSTSBI_FIRST_LINE_PREFIX){
			cur_output_type = "Run"
		}else if ischeck{
			cur_output_type = "Check"
		}else if strings.HasPrefix(line,COMPILE_FIRST_LINE_PREFIX){
			cur_output_type = "Compile"
		}else if strings.HasPrefix(last_line,TEST_PASSED_PREFIX){
			cur_output_type = "CI Output"
		}
		if cur_output_type != last_output_type{
			add_to_outputs(cur_output_lines,last_output_type,cur_has_fail,cur_n_pass,cur_n_fail)
			cur_output_lines=[]string{}
			cur_has_fail=false
			cur_n_pass=0
			cur_n_fail=0
		}
		last_line=line
		last_output_type=cur_output_type
		cur_output_lines=append(cur_output_lines,line) 
		if strings.HasPrefix(line,PASS_PREFIX){
			n_pass++;
			cur_n_pass++;
			tests=append(tests,Test{n_pass+n_fail,line[len(PASS_PREFIX):],true,1})
		}
		if strings.HasPrefix(line,FAIL_PREFIX){
			n_fail++
			cur_n_fail++
			tests=append(tests,Test{n_pass+n_fail,line[len(FAIL_PREFIX):],false,0})
			cur_has_fail=true
		}
		if strings.HasPrefix(line,TEST_PASSED_PREFIX){
			tmp:=strings.Split(strings.Split(line,": ")[1],"/")
			val1,err1:=strconv.Atoi(tmp[0])
			val2,err2:=strconv.Atoi(tmp[1])
			if err1!=nil || err2!=nil{
				fmt.Println("error")
			}
			test_passed_n1+=val1
			test_passed_n2+=val2
		}
	}
	add_to_outputs(cur_output_lines,last_output_type,cur_has_fail,cur_n_pass,cur_n_fail)
	return tests,outputs	
}

func HandleOsGrade(c *gin.Context){
	trace:=gitlab_api.Get_project_traces("labs-2018011302")
    tests,_:=Grade(trace)
	// test,_:=json.Marshal(tests)
	// fmt.Println(tests[0])
	c.JSON(http.StatusOK,gin.H{
		"tests":tests,
		"outputs":outputs,
	})
}