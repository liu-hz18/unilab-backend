package database

import (
	"unilab-backend/judger"
	"unilab-backend/logging"
	"unilab-backend/setting"
	"unilab-backend/utils"

	"github.com/panjf2000/ants/v2"
)

var p *ants.PoolWithFunc

func init() {
	p, _ = ants.NewPoolWithFunc(setting.JudgerPoolSize, func(i interface{}) {
		testConfig := i.(judger.TestConfig)
		// stat changes
		statResult := utils.DirStat(testConfig.ProgramDir)
		diffResult := utils.DirDiff(testConfig.PrevDir, testConfig.ProgramDir)
		logging.Info("begin launtch test: ", testConfig.TestID)
		result := judger.LaunchTest(testConfig)
		logging.Info("run result: ", result)
		logging.Info("running goroutines: ", p.Running())
		UpdateTestCaseRunResults(result, statResult, diffResult)
	}, ants.WithPreAlloc(true))
}

func LaunchTestAsync(config judger.TestConfig) {
	err := p.Invoke(config)
	if err != nil {
		logging.Error(err.Error())
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
