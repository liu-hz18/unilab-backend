package database

import (
	"errors"
	"runtime"
	"unilab-backend/judger"
	"unilab-backend/logging"
	"unilab-backend/setting"
	"unilab-backend/utils"

	"github.com/panjf2000/ants/v2"
)

var p *ants.PoolWithFunc

func init() {
	poolSize := utils.Min(setting.JudgerPoolSize, runtime.NumCPU())
	blockingQueueSize := utils.Clap(setting.JudgerPoolSize-poolSize, 4*runtime.NumCPU(), 300)
	logging.Info("Initialize Judger Pool with SIZE=", poolSize, " and QUEUESIZE=", blockingQueueSize)
	// a blocking goroutine pool
	var err error
	p, err = ants.NewPoolWithFunc(poolSize, func(i interface{}) {
		testConfig := i.(judger.TestConfig)
		// stat changes
		statResult := utils.DirStat(testConfig.ProgramDir)
		diffResult := utils.DirDiff(testConfig.PrevDir, testConfig.ProgramDir)
		logging.Info("begin launtch test: ", testConfig.TestID)
		result := judger.LaunchTest(testConfig)
		logging.Info("run result: ", result)
		logging.Info("running goroutines: ", p.Running(), " waiting goroutines: ", p.Waiting())
		UpdateTestCaseRunResults(result, statResult, diffResult)
	}, ants.WithPreAlloc(true), ants.WithNonblocking(false), ants.WithMaxBlockingTasks(blockingQueueSize))
	if err != nil {
		logging.Fatal(err.Error())
	}
}

func LaunchTestAsync(config judger.TestConfig) {
	err := p.Invoke(config)
	if err != nil {
		if errors.Is(err, ants.ErrPoolOverload) { // out of its capacity, should be pumped into pool
			logging.Error("POOL OVERFLOW! launtch test failed: ", config.TestID, " running goroutines: ", p.Running(), " waiting tasks: ", p.Waiting())
			// stat changes
			statResult := utils.DirStat(config.ProgramDir)
			diffResult := utils.DirDiff(config.PrevDir, config.ProgramDir)
			UpdateTestCaseRunResults(judger.TestResult{
				QuestionID:    config.QuestionID,
				TestID:        config.TestID,
				ProgramDir:    config.ProgramDir,
				CaseNum:       config.TestCaseNum,
				CompileResult: "",
				ExtraResult:   "Judger Pool Out Of Capacity!",
				TotalScore:    config.TotalScore,
				RunResults:    []judger.TestCaseResult{},
			}, statResult, diffResult)
		} else {
			logging.Error(err.Error())
		}
	}
}

func LaunchTestSync(config judger.TestConfig) {
	// stat changes
	statResult := utils.DirStat(config.ProgramDir)
	diffResult := utils.DirDiff(config.PrevDir, config.ProgramDir)
	result := judger.LaunchTest(config)
	logging.Info("run result: ", result)
	UpdateTestCaseRunResults(result, statResult, diffResult)
}

func Release() {
	defer p.Release()
}
