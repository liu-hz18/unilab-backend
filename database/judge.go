package database

import (
	"fmt"
	"log"
	"time"
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
		log.Printf("CreateSubmitRecord() begin trans action failed: %v", err)
	}
	// insert into test-run table
	result, err := tx.Exec(`INSERT INTO oj_db_test.oj_test_run
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
		log.Println(err)
		return 0, err
	}
	testID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return 0, err
	}
	// update question submit count
	_, err = tx.Exec(`UPDATE oj_db_test.oj_question SET question_test_total_num=question_test_total_num+1 WHERE question_id=?;`, form.QuestionID)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return 0, err
	}
	// insert into test-run table
	var insertTestCaseSql = "INSERT INTO oj_db_test.oj_testcase_run (testcase_rank, test_id) VALUES "
	for i := 0; i < int(testcase_num)-1; i++ {
		insertTestCaseSql += fmt.Sprintf("(%d, %d),", i, testID)
	}
	insertTestCaseSql += fmt.Sprintf("(%d, %d);", testcase_num-1, testID)
	_, err = tx.Exec(insertTestCaseSql)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return 0, err
	}
	_ = tx.Commit()
	log.Println("CreateSubmitRecord() commit trans action successfully.")
	return uint32(testID), nil
}

func GetQuestionSubmitCounts(questionID, userID uint32) (uint32, error) {
	totalRow, err := db.Query("SELECT COUNT(*) FROM oj_db_test.oj_test_run WHERE question_id=? AND user_id=?;", questionID, userID)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	defer totalRow.Close()
	var total uint32 = 0
	for totalRow.Next() {
		err := totalRow.Scan(&total)
		if err != nil {
			log.Println(err)
			continue
		}
	}
	return total, nil
}

func GetUserSubmitTests(courseID, userID uint32) ([]uint32, error) {
	rows, err := db.Query("SELECT test_id FROM oj_db_test.oj_test_run WHERE user_id=? AND course_id=? ORDER BY test_launch_time desc;", userID, courseID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	var results = []uint32{}
	var testID uint32
	for rows.Next() {
		err := rows.Scan(&testID)
		if err != nil {
			log.Println(err)
			continue
		}
		results = append(results, testID)
	}
	return results, nil
}

type TestCaseDetail struct {
	ID uint32
	State string
	TimeElasped uint32
	MemoryUsage uint32
}

type TestDetail struct {
	ID uint32
	QuestionID uint32
	Name string
	Score uint32
	TotalScore uint32
	Language string
	SubmitTime time.Time
	FileSize   string
	PassSubmission uint32
	TotalSubmission uint32
	TestCases []TestCaseDetail
}

func GetTestDetailsByIDs(testIDs []uint32) ([]TestDetail) {
	var testDetails = []TestDetail{}
	for _, testID := range testIDs {
		var testDetail TestDetail
		testDetail.ID = testID
		var saveDir string
		err := db.QueryRow("SELECT test_launch_time, question_id, language, save_dir FROM oj_db_test.oj_test_run WHERE test_id=?;", testID).Scan(
			&testDetail.SubmitTime,
			&testDetail.QuestionID,
			&testDetail.Language,
			&saveDir,
		)
		if err != nil {
			log.Println(err)
			continue
		}
		// read question details
		var testCaseNum uint32 
		err = db.QueryRow("SELECT question_name, question_score, question_test_ac_num, question_test_total_num, question_testcase_num FROM oj_db_test.oj_question WHERE question_id=?;", testDetail.QuestionID).Scan(
			&testDetail.Name,
			&testDetail.TotalScore,
			&testDetail.PassSubmission,
			&testDetail.TotalSubmission,
			&testCaseNum,
		)
		if err != nil {
			log.Println(err)
			continue
		}
		// read file sizes
		fileSize, err := utils.GetDirSize(saveDir)
		if err != nil {
			log.Println(err)
			continue
		}
		testDetail.FileSize = fmt.Sprintf("%d B", fileSize)
		// read test-run results
		rows, err := db.Query("SELECT testcase_run_id, testcase_run_state, testcase_run_time_elapsed, testcase_run_memory_usage FROM oj_db_test.oj_testcase_run WHERE test_id=? ORDER BY testcase_rank ASC;", testID)
		if err != nil {
			log.Println(err)
			continue
		}
		defer rows.Close()
		var validCaseCount uint32 = 0
		var passCount uint32 = 0
		for rows.Next() {
			var testcaseDetail TestCaseDetail
			err := rows.Scan(&testcaseDetail.ID, &testcaseDetail.State, &testcaseDetail.TimeElasped, &testcaseDetail.MemoryUsage)
			if err != nil {
				log.Println(err)
				continue
			}
			validCaseCount += 1
			if testcaseDetail.State == "Accepted" {
				passCount += 1
			}
			testDetail.TestCases = append(testDetail.TestCases, testcaseDetail)
		}
		if validCaseCount != testCaseNum {
			log.Println("ERROR: test case num DISMATCH between `oj_question` AND `oj_testcase_run`")
			continue
		}
		testDetail.Score = utils.CeilDivUint32(testDetail.TotalScore * passCount, testCaseNum);
		testDetails = append(testDetails, testDetail)
	}
	return testDetails
}
