package database

import (
	"log"
	"time"
)

type SubmitCodeForm struct {
	CourseID   uint32 `json:"courseid" form:"courseid" uri:"courseid" binding:"required"`
	QuestionID uint32 `json:"questionid" form:"questionid" uri:"questionid" binding:"required"`
	Language   string `json:"language" form:"language" uri:"language" binding:"required"`
}

func CreateSubmitCode(form SubmitCodeForm, userid uint32, save_dir string) (uint32, error) {
	// insert into test-run table
	result, err := db.Exec(`INSERT INTO oj_db_test.oj_test_run
		(test_launch_time, course_id, question_id, user_id, language, save_dir)
		VALUES
		(?, ?, ?, ?, ?, ?)
	`,
		time.Now(),
		form.CourseID,
		form.QuestionID,
		userid,
		form.Language,
		save_dir,
	)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	testID, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return 0, err
	}
	return uint32(testID), nil
}

func GetQuestionSubmitCounts(questionID, userID uint32) (uint32, error) {
	totalRow, err := db.Query("SELECT COUNT(*) FROM oj_db_test.oj_test_run WHERE question_id=? AND user_id=?", questionID, userID)
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
