package judger

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unilab-backend/logging"
)

const (
	RunFinished         uint32 = 0
	RuntimeError        uint32 = 2
	MemoryLimitExceeded uint32 = 3
	TimeLimitExceeded   uint32 = 4
	OutputLimitExceeded uint32 = 5
	DangerousSystemCall uint32 = 6
	JudgeFailed         uint32 = 7
)

const ResourceLimiter = "ulimit -t 5 && ulimit -v 524288 && ulimit -f 20480"

type TestConfig struct {
	ProblemName   string
	TimeLimit     uint32 // ms
	MemoryLimit   uint32 // KB
	TestCaseNum   uint32
	TestCaseScore []uint32
}

type TestCaseResult struct {
	TimeElasped   uint32
	MemoryUsage   uint32
	ExitCode      int
	RunStatus     uint32
	CheckerStatus uint32
	Accepted      bool
	CheckerOutput string
}

type TestResult struct {
	CompileResult string
	ExtraResult   string
	RunResults    []TestCaseResult
}

func check_uint_range(value, low, high uint32) bool {
	if value >= low && value <= high {
		return true
	}
	return false
}

func check_cfg(cfg TestConfig) string {
	var message string
	if cfg.ProblemName == "" {
		message = "Problem Name Not valid!"
		return message
	}
	if !check_uint_range(cfg.TestCaseNum, 1, 100) {
		message = "Test Case Num Not valid!"
		return message
	}
	if !check_uint_range(cfg.TimeLimit, 1, 10000) {
		message = "Test Time Limit Not valid!"
		return message
	}
	if !check_uint_range(cfg.MemoryLimit, 1, 2097152) {
		message = "Test Memory Limit Not valid!"
		return message
	}
	if len(cfg.TestCaseScore) != int(cfg.TestCaseNum) {
		message = "Test Case Score Not match test case num!"
		return message
	}
	for index, score := range cfg.TestCaseScore {
		if !check_uint_range(score, 0, 100) {
			message = fmt.Sprintf("Test Case %d Score out of range [0, 100] !", index)
			return message
		}
	}
	return ""
}

func check_test_case_dir(testCaseDir string, testCaseNum int) string {
	files, err := ioutil.ReadDir(testCaseDir)
	if err != nil {
		return err.Error()
	}
	if len(files) != 2*testCaseNum {
		return "Test Case Num DO NOT match Test Case Dir!"
	}
	for i := 1; i <= testCaseNum; i++ {
		data, err := ioutil.ReadFile(path.Join(testCaseDir, fmt.Sprintf("%d.in", i)))
		if err != nil {
			return fmt.Sprintf("Test Case %d need a %d.in !", i, i)
		}
		if string(data) == "" {
			return fmt.Sprintf("Test Case %d: %d.in is empty !", i, i)
		}
		data, err = ioutil.ReadFile(path.Join(testCaseDir, fmt.Sprintf("%d.ans", i)))
		if err != nil {
			return fmt.Sprintf("Test Case %d need a %d.ans !", i, i)
		}
		if string(data) == "" {
			return fmt.Sprintf("Test Case %d: %d.ans is empty !", i, i)
		}
	}
	return ""
}

func copyToDstDir(dstDir, srcDir string) error {
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		return err
	}
	for _, file := range files {
		dstFile, err := os.Create(path.Join(dstDir, file.Name()))
		if err != nil {
			return err
		}
		srcFile, err := os.Open(path.Join(srcDir, file.Name()))
		if err != nil {
			return err
		}
		_, err = io.Copy(dstFile, srcFile)
		if err != nil {
			return err
		}
	}
	return nil
}


func getErrorExitCode(err error) int {
	// fail, non-zero exit status conditions
	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.Sys().(syscall.WaitStatus).ExitStatus()
	}
	// fails that do not define an exec.ExitError (e.g. unable to identify executable on system PATH)
	return 1 // assign a default non-zero fail code value of 1
}

type Response struct {
	StdOut string
	StdErr string
	ExitCode int
}

func Subprocess(timeout int, executable string, pwd string, args ...string) Response {
	var res Response
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout+5) * time.Second)
	defer cancel()
	cmdarray := append([]string{"-c", fmt.Sprintf("%s && %s", ResourceLimiter, executable)}, args...)
	logging.Info("Exec Command: bash " + strings.Join(cmdarray, " "))
	cmd := exec.CommandContext(ctx, "bash", cmdarray...)
	if pwd != "" {
		cmd.Dir = pwd
	}
	// bash -c ulimit -t 5 && ulimit -v 524288 && ulimit -f 20480 && g++ main.cpp -O2 -o main.exe -fdiagnostics-color=always
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		logging.Info(err)
		res.StdErr = err.Error()
		res.StdOut = ""
		res.ExitCode = getErrorExitCode(err)
		return res
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		logging.Info(err)
		res.StdErr = err.Error()
		res.StdOut = ""
		res.ExitCode = getErrorExitCode(err)
		return res
	}
	defer stderr.Close()
	cmd.Start()
	outContent, err := ioutil.ReadAll(stdout)
	if err != nil {
		logging.Info(err)
		res.StdErr = err.Error()
		res.StdOut = ""
		res.ExitCode = getErrorExitCode(err)
		cmd.Process.Kill()
		return res
	}
	errContent, err := ioutil.ReadAll(stderr)
	if err != nil {
		logging.Info(err)
		res.StdErr = err.Error()
		res.StdOut = ""
		res.ExitCode = getErrorExitCode(err)
		cmd.Process.Kill()
		return res
	}
	err = cmd.Wait()
	if err != nil {
		logging.Info(err, string(outContent), string(errContent))
		res.StdErr = err.Error()
		res.StdOut = ""
		res.ExitCode = getErrorExitCode(err)
		return res
	}
	res.StdOut = string(outContent)
	res.StdErr = string(errContent)
	res.ExitCode = cmd.ProcessState.Sys().(syscall.WaitStatus).ExitStatus()
	return res
}


func LaunchTest(cfg TestConfig, testCaseDir string, programDir string) TestResult {
	var result TestResult
	result.CompileResult = ""
	result.ExtraResult = ""
	result.RunResults = []TestCaseResult{}
	// check config
	check_cfg_msg := check_cfg(cfg)
	if check_cfg_msg != "" {
		result.ExtraResult = check_cfg_msg
		return result
	}
	// check testcase
	check_testcase_msg := check_test_case_dir(testCaseDir, int(cfg.TestCaseNum))
	if check_testcase_msg != "" {
		result.ExtraResult = check_testcase_msg
		return result
	}
	// create temp directory
	tempDirName, err := ioutil.TempDir("", "*")
	tempDirName = filepath.ToSlash(tempDirName)
	// tempdir_in_linux := strings.Replace(tempDirName, "C:", "/mnt/c", -1)
	logging.Info("create temp directory: ", tempDirName)
	if err != nil {
		result.ExtraResult = err.Error()
		return result
	}
	// defer os.RemoveAll(tempDirName)
	// copy source code and test case into temp directory
	err = copyToDstDir(tempDirName, testCaseDir)
	if err != nil {
		result.ExtraResult = err.Error()
		return result
	}
	err = copyToDstDir(tempDirName, programDir)
	if err != nil {
		result.ExtraResult = err.Error()
		return result
	}
	// compile
	files, err := ioutil.ReadDir(tempDirName)
	// var haveMakeFile bool = false
	for _, file := range files {
		logging.Info(file.Name())
		// if strings.ToLower(file.Name()) == "makefile" {
		// 	haveMakeFile = true
		// 	break
		// }
	}
	cmd := "g++ main.cpp -O2 -o main.exe -fdiagnostics-color=always"
	response := Subprocess(30, cmd, tempDirName)
	logging.Info(response)
	if response.ExitCode != 0 {
		logging.Info("Compile Error: ", response.StdErr)
		result.CompileResult = response.StdErr
		return result
	}
	result.CompileResult = response.StdOut
	// run testcase
	for i := 1; i <= int(cfg.TestCaseNum); i++ {
		response = Subprocess(
			30, "./prebuilt/uoj_run.exe", "",
			fmt.Sprintf("--tl=%d", cfg.TimeLimit),
			fmt.Sprintf("--rtl=%d", cfg.TimeLimit + 1000),
			fmt.Sprintf("--ml=%d", cfg.MemoryLimit),
			fmt.Sprintf("--ol=%d", (64 * 1024)),
			fmt.Sprintf("--sl=%d", (64 * 1024)),
			fmt.Sprintf("--work-path=."),
			fmt.Sprintf("--res=%s", path.Join(tempDirName, "run_res.txt")),
			fmt.Sprintf("--err=%s", "/dev/stdout"),
			fmt.Sprintf("--in=%s", path.Join(tempDirName, fmt.Sprintf("%d.in", i))),
			fmt.Sprintf("--out=%s", path.Join(tempDirName, fmt.Sprintf("%d.out", i))),
			fmt.Sprintf("--err=%s", path.Join(tempDirName, fmt.Sprintf("%d.err", i))),
			path.Join(tempDirName, "main.exe"),
		)
		logging.Info("Testcase ", i, " Response: ", response)
		// fill-in results
		var test_case_result TestCaseResult
		run_res, _ := ioutil.ReadFile(path.Join(tempDirName, "run_res.txt"))
		logging.Info("Run result: ", string(run_res))
		run_res_arr := strings.Fields(string(run_res))
		var success bool = true
		test_case_result.Accepted = false
		test_case_result.CheckerOutput = ""
		if len(run_res_arr) == 0 {
			logging.Info("No Output in file run_res.txt")
			test_case_result.RunStatus = 7
			test_case_result.TimeElasped = 0
			test_case_result.MemoryUsage = 0
			test_case_result.ExitCode = 1
			success = false
		} else {
			run_status, _ := strconv.ParseUint(run_res_arr[0], 10, 32)
			time_elasped, _ := strconv.ParseUint(run_res_arr[1], 10, 32)
			memory_usage, _ := strconv.ParseUint(run_res_arr[2], 10, 32)
			exit_code, _ := strconv.ParseInt(run_res_arr[3], 10, 32)
			test_case_result.RunStatus = uint32(run_status)
			test_case_result.TimeElasped = uint32(time_elasped)
			test_case_result.MemoryUsage = uint32(memory_usage)
			test_case_result.ExitCode = int(exit_code)
			success = (run_status == 0)
		}
		// check .ans and .out
		if success {
			response = Subprocess(
				30, "./prebuilt/uoj_run.exe", "",
				fmt.Sprintf("--tl=%d", (5 * 1000)),
				fmt.Sprintf("--rtl=%d", (10 * 1000)),
				fmt.Sprintf("--ml=%d", (512 * 1024)),
				fmt.Sprintf("--ol=%d", (64 * 1024)),
				fmt.Sprintf("--sl=%d", (64 * 1024)),
				fmt.Sprintf("--work-path=."),
				fmt.Sprintf("--res=%s", path.Join(tempDirName, "spj_run_res.txt")),
				fmt.Sprintf("--err=%s", "/dev/stdout"),
				fmt.Sprintf("--add-readable=%s", path.Join(tempDirName, fmt.Sprintf("%d.in", i))),
				fmt.Sprintf("--add-readable=%s", path.Join(tempDirName, fmt.Sprintf("%d.out", i))),
				fmt.Sprintf("--add-readable=%s", path.Join(tempDirName, fmt.Sprintf("%d.ans", i))),
				path.Join(tempDirName, "prebuilt/fcmp.exe"),
				path.Join(tempDirName, fmt.Sprintf("%d.in", i)),
				path.Join(tempDirName, fmt.Sprintf("%d.out", i)),
				path.Join(tempDirName, fmt.Sprintf("%d.ans", i)),
			)
			if response.ExitCode != 0 {
				test_case_result.CheckerStatus = 7
			} else {
				spj_run_res, _ := ioutil.ReadFile(path.Join(tempDirName, "spj_run_res.txt"))
				logging.Info("Judger Run result: ", string(spj_run_res))
				spj_run_res_arr := strings.Fields(string(spj_run_res))
				if len(spj_run_res_arr) == 0 {
					logging.Info("No Output in file spj_run_res.txt")
					test_case_result.CheckerStatus = 7
				} else {
					checker_status, _ := strconv.ParseUint(spj_run_res_arr[0], 10, 32)
					test_case_result.CheckerStatus = uint32(checker_status)
					// parse checker output
					test_case_result.Accepted = (response.StdOut[:3] == "ok ")
					test_case_result.CheckerOutput = response.StdOut
				}
			}
		}
		result.RunResults = append(result.RunResults, test_case_result)
	}
	return result	
}
