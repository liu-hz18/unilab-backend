package judger

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"strconv"
	"strings"
	"unilab-backend/logging"
	"unilab-backend/utils"
)

// const BackendRootDir = "/home/cslab/unilab/unilab-backend/"
const BackendRootDir = "/unilab-backend/"

type TestConfig struct {
	QuestionID  uint32
	TestID      uint32
	TimeLimit   uint32 // ms
	MemoryLimit uint32 // KB
	TestCaseNum uint32
	Language    string
	TotalScore  uint32
	QuestionDir string
	ProgramDir  string
	PrevDir     string // 上次提交的文件夹路径
	// TestCaseScore []uint32
}

type TestCaseResult struct {
	TimeElapsed   uint32
	MemoryUsage   uint32
	ExitCode      int
	RunStatus     uint32
	CheckerStatus uint32
	Accepted      bool
	CheckerOutput string
}

type TestResult struct {
	QuestionID    uint32
	TestID        uint32
	ProgramDir    string
	CaseNum       uint32
	CompileResult string
	ExtraResult   string
	TotalScore    uint32
	RunResults    []TestCaseResult
}

func CheckUIntRange(value, low, high uint32) bool {
	if value >= low && value <= high {
		return true
	}
	return false
}

func checkConfig(cfg TestConfig) string {
	var message string
	if !CheckUIntRange(cfg.TestCaseNum, 1, 100) {
		message = "Test Case Num Not valid!"
		return message
	}
	if !CheckUIntRange(cfg.TimeLimit, 1, 10000) {
		message = "Test Time Limit Not valid!"
		return message
	}
	if !CheckUIntRange(cfg.MemoryLimit, 1, 2048*1024) {
		message = "Test Memory Limit Not valid!"
		return message
	}
	// if len(cfg.TestCaseScore) != int(cfg.TestCaseNum) {
	// 	message = "Test Case Score Not match test case num!"
	// 	return message
	// }
	// for index, score := range cfg.TestCaseScore {
	// 	if !check_uint_range(score, 0, 100) {
	// 		message = fmt.Sprintf("Test Case %d Score out of range [0, 100] !", index)
	// 		return message
	// 	}
	// }
	return ""
}

func checkTestCaseDir(testCaseDir string, testCaseNum int) string {
	// files, err := io.ReadDir(testCaseDir)
	// if err != nil {
	// 	return err.Error()
	// }
	// if len(files) != 2*testCaseNum {
	// 	return "Test Case Num DO NOT match Test Case Dir's TestCase Files!"
	// }
	for i := 1; i <= testCaseNum; i++ {
		_, err := os.ReadFile(path.Join(testCaseDir, fmt.Sprintf("%d.in", i)))
		if err != nil {
			return fmt.Sprintf("Test Case %d need a %d.in !", i, i)
		}
		// if string(data) == "" {
		// 	return fmt.Sprintf("Test Case %d: %d.in is empty !", i, i)
		// }
		_, err = os.ReadFile(path.Join(testCaseDir, fmt.Sprintf("%d.ans", i)))
		if err != nil {
			return fmt.Sprintf("Test Case %d need a %d.ans !", i, i)
		}
		// if string(data) == "" {
		// 	return fmt.Sprintf("Test Case %d: %d.ans is empty !", i, i)
		// }
	}
	return ""
}

func trimJSCompileOutputs(results string) string {
	if results == "" {
		return ""
	}
	lines := strings.Split(results, "\n")
	index := 0
	startline := ""
	for idx, line := range lines {
		if len(line) > 4 && line[0:4] == "/tmp" {
			index = idx + 1
			arr := strings.Split(line, "/")
			startline = arr[len(arr)-1] + "\n"
		}
	}
	return startline + strings.Join(lines[index:], "\n")
}

func LaunchTest(cfg TestConfig) TestResult {
	var result TestResult
	result.QuestionID = cfg.QuestionID
	result.TestID = cfg.TestID
	result.CaseNum = cfg.TestCaseNum
	result.TotalScore = cfg.TotalScore
	result.ProgramDir = cfg.ProgramDir
	result.CompileResult = ""
	result.ExtraResult = ""
	result.RunResults = []TestCaseResult{}
	testCaseDir, _ := filepath.Abs(cfg.QuestionDir)
	programDir := cfg.ProgramDir
	// check config
	checkCfgMsg := checkConfig(cfg)
	if checkCfgMsg != "" {
		result.ExtraResult = checkCfgMsg
		return result
	}
	logging.Info("Launch Test ID: ", cfg.TestID)
	// check testcase
	checkTestcaseMsg := checkTestCaseDir(testCaseDir, int(cfg.TestCaseNum))
	if checkTestcaseMsg != "" {
		result.ExtraResult = checkTestcaseMsg
		return result
	}
	// create temp directory
	tempDirName, err := os.MkdirTemp("", "unilab-oj-*")
	logging.Info("create temp directory: ", tempDirName)
	if err != nil {
		result.ExtraResult = err.Error()
		return result
	}
	defer os.RemoveAll(tempDirName)
	// copy source code into temp directory
	// err = copyToDstDir(tempDirName, testCaseDir)
	// if err != nil {
	// 	result.ExtraResult = err.Error()
	// 	return result
	// }
	err = utils.CopyToDstDir(tempDirName, programDir)
	if err != nil {
		result.ExtraResult = err.Error()
		return result
	}
	// compile
	files, err := os.ReadDir(tempDirName)
	if err != nil {
		result.ExtraResult = err.Error()
		return result
	}
	// var haveMakeFile bool = false
	for _, file := range files {
		logging.Info(file.Name())
		// if strings.ToLower(file.Name()) == "makefile" {
		// 	haveMakeFile = true
		// 	break
		// }
	}
	// switch languages
	var compileCmd, runType, exeName, runtimeRlimits, compileRlimits string
	var timeOut int
	if langConf, ok := JudgerConfig[cfg.Language]; ok {
		compileCmd = langConf.Compile
		compileCmd = strings.ReplaceAll(compileCmd, "{SourceFile}", path.Join(tempDirName, langConf.SourceFile))
		compileCmd = strings.ReplaceAll(compileCmd, "{Executable}", path.Join(tempDirName, langConf.Executable))
		runType = langConf.RunType
		exeName = langConf.Executable
		runtimeRlimits = langConf.RuntimeLimits
		compileRlimits = langConf.CompileLimits
		timeOut = langConf.Timeout
	} else {
		result.CompileResult = "Language Not Supported: " + cfg.Language
		logging.Error(result.CompileResult)
		return result
	}
	// compile
	response := utils.Subprocess(compileRlimits, 10, compileCmd, tempDirName)
	logging.Info(response)
	if response.ExitCode != 0 {
		if response.StdErr != "" {
			result.CompileResult = response.StdErr
		} else if response.StdOut != "" {
			result.CompileResult = response.StdOut
		} else {
			result.CompileResult = response.ServerErr
		}
		if cfg.Language == "js" {
			result.CompileResult = trimJSCompileOutputs(result.CompileResult)
		}
		return result
	}
	result.CompileResult = response.StdOut

	// sandbox check compile process
	// response := utils.Subprocess(
	// 	compileRlimits, 10, BackendRootDir+"prebuilt/uoj_run", tempDirName, // compile in temp dir
	// 	fmt.Sprintf("--tl=%d", 10*1000),       // Set cpu time limit (in ms)
	// 	fmt.Sprintf("--rtl=%d", 10*1000+1000), // Set real time limit (in ms)
	// 	fmt.Sprintf("--ml=%d", 512*1024),      // Set memory limit (in kb)
	// 	fmt.Sprintf("--ol=%d", (64*1024)),     // Set output limit (in kb)
	// 	fmt.Sprintf("--sl=%d", (64*1024)),     // Set stack limit (in kb)
	// 	fmt.Sprintf("--work-path=%s", tempDirName),
	// 	fmt.Sprintf("--res=%s", path.Join(tempDirName, "run_compile_res.txt")),
	// 	"--type=compiler",
	// 	"--in=/dev/null",
	// 	fmt.Sprintf("--out=%s", path.Join(tempDirName, "compile_stdout.txt")),
	// 	fmt.Sprintf("--err=%s", path.Join(tempDirName, "compile_res.txt")), // compiler stderr
	// 	"--show-trace-details",                                             // ONLY FOR DEBUG
	// 	compileCmd,
	// )
	// logging.Info("Compiler Response: ", response)
	// result.CompileResult = ""
	// // read compile result file
	// compileRes, _ := os.ReadFile(path.Join(tempDirName, "run_compile_res.txt"))
	// logging.Info("Run result: ", string(compileRes))
	// compileResArr := strings.Fields(string(compileRes))
	// if len(compileResArr) != 4 {
	// 	result.CompileResult = "Compile Failed."
	// 	return result
	// }
	// compileStatus, _ := strconv.ParseUint(compileResArr[0], 10, 32)
	// compileExitCode, _ := strconv.ParseUint(compileResArr[3], 10, 32)
	// read compiler output
	// compileStdErr, _ := os.ReadFile(path.Join(tempDirName, "compile_res.txt"))
	// compileStdOut, _ := os.ReadFile(path.Join(tempDirName, "compile_stdout.txt"))
	// logging.Info("compiler stdout: ", string(compileStdOut))
	// logging.Info("compiler stderr: ", string(compileStdErr))

	// if compileStatus != 0 || compileExitCode != 0 {
	// 	if compileStatus == 0 {
	// 		// read compiler err output
	// 		compilerOutput, err := os.ReadFile(path.Join(tempDirName, "compile_res.txt"))
	// 		if err != nil {
	// 			logging.Info(err.Error())
	// 		} else {
	// 			logging.Info("ExitCode != 0, compiler output: ", string(compilerOutput))
	// 		}
	// 		result.CompileResult = string(compilerOutput)
	// 		if cfg.Language == "js" {
	// 			result.CompileResult = trimJSCompileOutputs(result.CompileResult)
	// 		}
	// 	} else if compileStatus == 7 {
	// 		result.CompileResult = "Compile Failed."
	// 		compilerOutput, err := os.ReadFile(path.Join(tempDirName, "compile_res.txt"))
	// 		if err != nil {
	// 			logging.Info(err.Error())
	// 		} else {
	// 			logging.Info("JGF! compiler output: ", string(compilerOutput))
	// 		}
	// 	} else {
	// 		result.CompileResult = fmt.Sprintf("Compiler Runstatus: %d", compileStatus)
	// 	}
	// 	return result
	// }

	// run testcase
	for i := 1; i <= int(cfg.TestCaseNum); i++ {
		response = utils.Subprocess(
			runtimeRlimits, timeOut, BackendRootDir+"prebuilt/uoj_run", tempDirName, // NOTE: work in current dir, not in tmp dir
			fmt.Sprintf("--tl=%d", cfg.TimeLimit),
			fmt.Sprintf("--rtl=%d", cfg.TimeLimit+1000),
			fmt.Sprintf("--ml=%d", cfg.MemoryLimit),
			fmt.Sprintf("--ol=%d", (64*1024)),
			fmt.Sprintf("--sl=%d", (64*1024)),
			fmt.Sprintf("--work-path=%s", tempDirName),                     // Set the work path of the program
			fmt.Sprintf("--res=%s", path.Join(tempDirName, "run_res.txt")), // Set the file name for outputing the sandbox result
			fmt.Sprintf("--type=%s", runType),
			fmt.Sprintf("--in=%s", path.Join(testCaseDir, fmt.Sprintf("%d.in", i))),
			fmt.Sprintf("--out=%s", path.Join(tempDirName, fmt.Sprintf("%d.out", i))),
			fmt.Sprintf("--err=%s", path.Join(tempDirName, fmt.Sprintf("%d.err", i))),
			"--show-trace-details", // ONLY FOR DEBUG
			path.Join(tempDirName, exeName),
		)
		logging.Info("Testcase ", i, " Response: ", response)
		// fill-in results
		var testCaseResult TestCaseResult
		runRes, _ := os.ReadFile(path.Join(tempDirName, "run_res.txt"))
		logging.Info("Run result: ", string(runRes))
		runResArr := strings.Fields(string(runRes))
		var success bool
		testCaseResult.Accepted = false
		testCaseResult.CheckerOutput = ""
		if len(runResArr) != 4 {
			logging.Info("No Output in file run_res.txt")
			testCaseResult.RunStatus = 7
			testCaseResult.TimeElapsed = 0
			testCaseResult.MemoryUsage = 0
			testCaseResult.ExitCode = 1
			success = false
		} else {
			runStatus, _ := strconv.ParseUint(runResArr[0], 10, 32)
			timeElapsed, _ := strconv.ParseUint(runResArr[1], 10, 32)
			memoryUsage, _ := strconv.ParseUint(runResArr[2], 10, 32)
			exitCode, _ := strconv.ParseInt(runResArr[3], 10, 32)
			if exitCode != 0 && runStatus == 0 {
				testCaseResult.RunStatus = RuntimeError
				runRes, _ := os.ReadFile(path.Join(tempDirName, fmt.Sprintf("%d.err", i)))
				logging.Info("program stderr: ", string(runRes))
				runRes, _ = os.ReadFile(path.Join(tempDirName, fmt.Sprintf("%d.out", i)))
				logging.Info("program stdout: ", string(runRes))
			} else {
				testCaseResult.RunStatus = uint32(runStatus)
			}
			testCaseResult.TimeElapsed = uint32(timeElapsed)
			testCaseResult.MemoryUsage = uint32(memoryUsage)
			testCaseResult.ExitCode = int(exitCode)
			success = (testCaseResult.RunStatus == 0)
		}
		// check .ans and .out
		if success {
			response = utils.Subprocess(
				// need `sudo` to run with `CheckerResourceLimiter`
				"", 10, BackendRootDir+"prebuilt/uoj_run", tempDirName,
				fmt.Sprintf("--tl=%d", (5*1000)),
				fmt.Sprintf("--rtl=%d", (10*1000)),
				fmt.Sprintf("--ml=%d", (512*1024)),
				fmt.Sprintf("--ol=%d", (64*1024)),
				fmt.Sprintf("--sl=%d", (64*1024)),
				fmt.Sprintf("--work-path=%s", tempDirName),
				fmt.Sprintf("--res=%s", path.Join(tempDirName, "spj_run_res.txt")),
				fmt.Sprintf("--err=%s", "/dev/stdout"),
				"--show-trace-details", // ONLY FOR DEBUG
				fmt.Sprintf("--add-readable=%s", path.Join(testCaseDir, fmt.Sprintf("%d.in", i))),
				fmt.Sprintf("--add-readable=%s", path.Join(tempDirName, fmt.Sprintf("%d.out", i))),
				fmt.Sprintf("--add-readable=%s", path.Join(testCaseDir, fmt.Sprintf("%d.ans", i))),
				fmt.Sprintf("--add-readable=%s", BackendRootDir+"prebuilt/fcmp"),
				BackendRootDir+"prebuilt/fcmp",
				path.Join(testCaseDir, fmt.Sprintf("%d.in", i)),
				path.Join(tempDirName, fmt.Sprintf("%d.out", i)),
				path.Join(testCaseDir, fmt.Sprintf("%d.ans", i)),
			)
			if response.ExitCode != 0 || len(response.StdOut) < 2 {
				testCaseResult.CheckerStatus = 7
			} else {
				spjRunRes, _ := os.ReadFile(path.Join(tempDirName, "spj_run_res.txt"))
				logging.Info("Judger Run result: ", string(spjRunRes))
				spjRunResArr := strings.Fields(string(spjRunRes))
				if len(spjRunResArr) != 4 {
					logging.Info("No Output in file spj_run_res.txt")
					testCaseResult.CheckerStatus = 7
				} else {
					checkerStatus, _ := strconv.ParseUint(spjRunResArr[0], 10, 32)
					testCaseResult.CheckerStatus = uint32(checkerStatus)
					// parse checker output
					testCaseResult.Accepted = (response.StdOut[:2] == "ok")
					testCaseResult.CheckerOutput = response.StdOut
				}
			}
		}
		result.RunResults = append(result.RunResults, testCaseResult)
	}
	return result
}
