package database

import (
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strconv"
	"time"
	"unilab-backend/judger"
	"unilab-backend/logging"
	"unilab-backend/setting"
	"unilab-backend/utils"
)

type SubmitCodeForm struct {
	CourseID   uint32 `json:"courseid" form:"courseid" uri:"courseid" binding:"required"`
	QuestionID uint32 `json:"questionid" form:"questionid" uri:"questionid" binding:"required"`
	Language   string `json:"language" form:"language" uri:"language" binding:"required"`
}

func CreateSubmitRecord(form SubmitCodeForm, userid uint32, save_dir string, testcase_num uint32) (uint32, error) {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("CreateSubmitRecord() begin trans action failed: ", err)
	}
	// insert into test-run table
	result, err := tx.Exec(`INSERT INTO oj_test_run
		(test_launch_time, course_id, question_id, user_id, language, save_dir)
		VALUES
		(?, ?, ?, ?, ?, ?);
	`,
		time.Now(),
		form.CourseID,
		form.QuestionID,
		userid,
		form.Language,
		save_dir,
	)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	testID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	// update question submit count
	_, err = tx.Exec(`UPDATE oj_question SET question_test_total_num=question_test_total_num+1 WHERE question_id=?;`, form.QuestionID)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	// insert into test-run table
	var insertTestCaseSql = "INSERT INTO oj_testcase_run (testcase_rank, test_id) VALUES "
	for i := 0; i < int(testcase_num)-1; i++ {
		insertTestCaseSql += fmt.Sprintf("(%d, %d),", i, testID)
	}
	insertTestCaseSql += fmt.Sprintf("(%d, %d);", testcase_num-1, testID)
	_, err = tx.Exec(insertTestCaseSql)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	_ = tx.Commit()
	logging.Info("CreateSubmitRecord() commit trans action successfully.")
	return uint32(testID), nil
}

func RunTest(testID uint32) {
	// read test meta info from `oj_test_run`
	var programDir string
	var questionID uint32
	var language string
	err := db.QueryRow("SELECT save_dir, question_id, language from oj_test_run WHERE test_id=?;", testID).Scan(&programDir, &questionID, &language)
	if err != nil {
		logging.Error(err)
		return
	}
	// read question meta info from `oj_question`
	var name string
	var timeLimit uint32
	var memoryLimit uint32
	var testcaseNum uint32
	err = db.QueryRow("SELECT question_name, question_time_limit, question_memory_limit, question_testcase_num from oj_question WHERE question_id=?;", questionID).Scan(
		&name,
		&timeLimit,
		&memoryLimit,
		&testcaseNum,
	)
	if err != nil {
		logging.Error(err)
		return
	}
	questionDir := setting.QuestionRootDir + strconv.FormatUint(uint64(questionID), 10) + "_" + name + "/"

	config := judger.TestConfig{
		QuestionID:  questionID,
		TestID:      testID,
		TimeLimit:   timeLimit,
		MemoryLimit: memoryLimit * 1024, // frontend:MB -> backend:KB
		TestCaseNum: testcaseNum,
		Language:    language,
	}
	result := judger.LaunchTest(config, questionDir, programDir)
	logging.Info("run result: ", result)
	UpdateTestCaseRunResults(result)
}

func GetQuestionSubmitCounts(questionID, userID uint32) (uint32, error) {
	totalRow, err := db.Query("SELECT COUNT(*) FROM oj_test_run WHERE question_id=? AND user_id=?;", questionID, userID)
	if err != nil {
		logging.Info(err)
		return 0, err
	}
	defer totalRow.Close()
	var total uint32 = 0
	for totalRow.Next() {
		err := totalRow.Scan(&total)
		if err != nil {
			logging.Info(err)
			continue
		}
	}
	return total, nil
}

func GetUserSubmitTests(courseID, userID uint32) ([]uint32, error) {
	rows, err := db.Query("SELECT test_id FROM oj_test_run WHERE user_id=? AND course_id=? ORDER BY test_launch_time desc;", userID, courseID)
	if err != nil {
		logging.Info(err)
		return nil, err
	}
	defer rows.Close()
	var results = []uint32{}
	var testID uint32
	for rows.Next() {
		err := rows.Scan(&testID)
		if err != nil {
			logging.Info(err)
			continue
		}
		results = append(results, testID)
	}
	return results, nil
}

type TestCaseDetail struct {
	ID          uint32
	State       string
	TimeElasped uint32
	MemoryUsage uint32
}

type TestDetail struct {
	ID              uint32
	QuestionID      uint32
	Name            string
	Score           uint32
	TotalScore      uint32
	Language        string
	SubmitTime      time.Time
	FileSize        string
	PassSubmission  uint32
	TotalSubmission uint32
	TestCases       []TestCaseDetail
}

type FileInfo struct {
	Name    string
	Content string
	Lint    string
}

type SubmitDetail struct {
	Fileinfo []FileInfo
	Compile  string
	Extra    string
}

func GetTestDetailsByIDs(testIDs []uint32) []TestDetail {
	var testDetails = []TestDetail{}
	for _, testID := range testIDs {
		var testDetail TestDetail
		testDetail.ID = testID
		var saveDir string
		err := db.QueryRow("SELECT test_launch_time, question_id, language, save_dir FROM oj_test_run WHERE test_id=?;", testID).Scan(
			&testDetail.SubmitTime,
			&testDetail.QuestionID,
			&testDetail.Language,
			&saveDir,
		)
		if err != nil {
			logging.Info(err)
			continue
		}
		// read question details
		var testCaseNum uint32
		err = db.QueryRow("SELECT question_name, question_score, question_test_ac_num, question_test_total_num, question_testcase_num FROM oj_question WHERE question_id=?;", testDetail.QuestionID).Scan(
			&testDetail.Name,
			&testDetail.TotalScore,
			&testDetail.PassSubmission,
			&testDetail.TotalSubmission,
			&testCaseNum,
		)
		if err != nil {
			logging.Info(err)
			continue
		}
		// read file sizes
		fileSize, err := utils.GetDirSize(saveDir)
		if err != nil {
			logging.Info(err)
			continue
		}
		testDetail.FileSize = fmt.Sprintf("%d B", fileSize)
		// read test-run results
		rows, err := db.Query("SELECT testcase_run_id, testcase_run_state, testcase_run_time_elapsed, testcase_run_memory_usage FROM oj_testcase_run WHERE test_id=? ORDER BY testcase_rank ASC;", testID)
		if err != nil {
			logging.Info(err)
			continue
		}
		defer rows.Close()
		var validCaseCount uint32 = 0
		var passCount uint32 = 0
		for rows.Next() {
			var testcaseDetail TestCaseDetail
			err := rows.Scan(&testcaseDetail.ID, &testcaseDetail.State, &testcaseDetail.TimeElasped, &testcaseDetail.MemoryUsage)
			if err != nil {
				logging.Info(err)
				continue
			}
			validCaseCount += 1
			if testcaseDetail.State == "Accepted" {
				passCount += 1
			}
			testDetail.TestCases = append(testDetail.TestCases, testcaseDetail)
		}
		if validCaseCount != testCaseNum {
			logging.Info("ERROR: test case num DISMATCH between `oj_question` AND `oj_testcase_run`")
			continue
		}
		testDetail.Score = utils.CeilDivUint32(testDetail.TotalScore*passCount, testCaseNum)
		testDetails = append(testDetails, testDetail)
	}
	return testDetails
}

func UpdateTestCaseRunResults(judgerResult judger.TestResult) {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("UpdateTestCaseRunResults() begin trans action failed: ", err)
	}
	_, err = tx.Exec(`
		UPDATE oj_test_run SET
		compile_result=?,
		extra_result=?
		WHERE
		test_id=?;
	`,
		judgerResult.CompileResult,
		judgerResult.ExtraResult,
		judgerResult.TestID,
	)
	logging.Info("extra result:", judgerResult.ExtraResult)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return
	}
	var code uint32
	var is_ac bool = true
	if len(judgerResult.RunResults) == int(judgerResult.CaseNum) {
		for rank, runResults := range judgerResult.RunResults {
			if runResults.RunStatus == judger.RunFinished {
				if runResults.Accepted {
					code = judger.RunFinished
				} else {
					code = judger.WrongAnswer
					is_ac = false
				}
			} else {
				code = runResults.RunStatus
				is_ac = false
			}
			// logging.Info("t: ", runResults.TimeElasped, "m: ", runResults.MemoryUsage)
			_, err = tx.Exec(`
				UPDATE oj_testcase_run SET 
				testcase_run_state=?,
				testcase_run_time_elapsed=?,
				testcase_run_memory_usage=?,
				testcase_checker_output=?
				WHERE
				test_id=? AND testcase_rank=?;
			`,
				judger.RunResultMap[code],
				runResults.TimeElasped,
				runResults.MemoryUsage,
				runResults.CheckerOutput,
				judgerResult.TestID,
				rank,
			)
			if err != nil {
				_ = tx.Rollback()
				logging.Info(err)
				return
			}
		}
	} else {
		is_ac = false
		if judgerResult.CompileResult == "" {
			code = judger.JudgeFailed
		} else {
			code = judger.CompileError
		}
		for rank := 0; rank < int(judgerResult.CaseNum); rank++ {
			_, err = tx.Exec(`
				UPDATE oj_testcase_run SET 
				testcase_run_state=?,
				testcase_run_time_elapsed=?,
				testcase_run_memory_usage=?
				WHERE
				test_id=? AND testcase_rank=?;
			`,
				judger.RunResultMap[code],
				0,
				0,
				judgerResult.TestID,
				rank,
			)
			if err != nil {
				_ = tx.Rollback()
				logging.Info(err)
				return
			}
		}
	}
	if is_ac {
		// update question submit count
		_, err = tx.Exec(`UPDATE oj_question SET question_test_ac_num=question_test_ac_num+1 WHERE question_id=?;`, judgerResult.QuestionID)
		if err != nil {
			_ = tx.Rollback()
			logging.Info(err)
			return
		}
	}
	_ = tx.Commit()
	logging.Info("UpdateTestCaseRunResults() commit trans action successfully.")
	return
}

func GetSubmitDetail(testID uint32) SubmitDetail {
	var result = SubmitDetail{}
	var save_dir string
	err := db.QueryRow("SELECT compile_result, extra_result, save_dir FROM oj_test_run WHERE test_id=?;", testID).Scan(
		&result.Compile,
		&result.Extra,
		&save_dir,
	)
	if err != nil {
		logging.Info(err)
		return result
	}
	files, err := ioutil.ReadDir(save_dir)
	if err != nil {
		logging.Info(err)
		return result
	}
	for _, file := range files {
		var info FileInfo
		info.Name = file.Name()
		path := path.Join(save_dir, info.Name)
		// read file
		content_bytes, err := ioutil.ReadFile(path)
		if err != nil {
			logging.Info(err)
			return result
		}
		info.Content = string(content_bytes)
		if lint, ok := judger.ExtLint[filepath.Ext(info.Name)]; !ok {
			info.Lint = ""
		} else {
			info.Lint = lint
		}
		result.Fileinfo = append(result.Fileinfo, info)
	}
	return result
}
