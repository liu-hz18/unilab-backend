package database

import (
	"unilab-backend/judger"
	"unilab-backend/logging"
	"unilab-backend/setting"

	"github.com/panjf2000/ants/v2"
)

var p *ants.PoolWithFunc

func init() {
	p, _ = ants.NewPoolWithFunc(setting.JudgerPoolSize, func(i interface{}) {
		test_config := i.(judger.TestConfig)
		logging.Info("begin launtch test: ", test_config.TestID)
		result := judger.LaunchTest(test_config)
		logging.Info("run result: ", result)
		logging.Info("running goroutines: ", p.Running())
		UpdateTestCaseRunResults(result)
	}, ants.WithPreAlloc(true))
}

func LaunchTestAsync(config judger.TestConfig) {
	p.Invoke(config)
}

func LaunchTestSync(config judger.TestConfig) {
	result := judger.LaunchTest(config)
	logging.Info("run result: ", result)
	UpdateTestCaseRunResults(result)
}

func Release() {
	defer p.Release()
}
