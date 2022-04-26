package database

import (
	"fmt"
	"time"
	"unilab-backend/logging"
)

type CreateAssignmentForm struct {
	Title       string    `json:"title" form:"title" uri:"title" binding:"required"`
	Description string `json:"description" form:"description " uri:"description" binding:"required"`
	BeginTime   string `json:"begintime" form:"begintime" uri:"begintime" binding:"required"`
	DueTime     string `json:"duetime" form:"duetime" uri:"duetime" binding:"required"`
	CourseID    uint32 `json:"courseid" form:"courseid" uri:"courseid" binding:"required"`
	QuestionIDs []uint32 `json:"questionids" form:"questionids" uri:"questionids" binding:"required"`
}

type CreateAssignmentInfo struct {
	Title       string    `json:"title" form:"title" uri:"title" binding:"required"`
	Description string `json:"description" form:"description " uri:"description" binding:"required"`
	BeginTime   time.Time `json:"begintime" form:"begintime" uri:"begintime" binding:"required"`
	DueTime     time.Time `json:"duetime" form:"duetime" uri:"duetime" binding:"required"`
	CourseID    uint32 `json:"courseid" form:"courseid" uri:"courseid" binding:"required"`
	QuestionIDs []uint32 `json:"questionids" form:"questionids" uri:"questionids" binding:"required"`
}

type TestRunInfo struct {
	QuestionID uint32
	TestID     uint32
	LaunchTime time.Time
}

type AssignmentQuestionInfo struct {
	ID          uint32
	Title       string
	Tag         string
	Score       uint32
	TestCaseNum uint32
	TotalScore  uint32
}

type Assignment struct {
	ID          uint32
	Title 	    string
	Description string
	BeginTime   time.Time
	DueTime 	time.Time
	Questions []AssignmentQuestionInfo
}

func CreateAssignment(form CreateAssignmentInfo) (uint32, error) {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("CreateAssignment() begin trans action failed: ", err)
	}
	result, err := tx.Exec(`INSERT INTO oj_homework
		(homework_name, homework_begin_time, homework_due_time, homework_description, course_id)
		VALUES
		(?, ?, ?, ?, ?)
	`,
		form.Title,
		form.BeginTime,
		form.DueTime,
		form.Description,
		form.CourseID,
	)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	assignmentID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	// create question & assignment relation
	var insertAssignmentQuestions string = "INSERT INTO oj_question_homework (question_id, homework_id) VALUES "
	for index, questionID := range form.QuestionIDs {
		if index < len(form.QuestionIDs)-1 {
			insertAssignmentQuestions += fmt.Sprintf("(%d, %d),", questionID, assignmentID)
		} else {
			insertAssignmentQuestions += fmt.Sprintf("(%d, %d);", questionID, assignmentID)
		}
	}
	logging.Info(insertAssignmentQuestions)
	_, err = tx.Exec(insertAssignmentQuestions)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	_ = tx.Commit()
	logging.Info("CreateAssignment() commit trans action successfully.")
	return uint32(assignmentID), nil
}


func GetAssignemntInfo(CourseID uint32, UserID uint32) ([]Assignment, error) {
	// read assignments
	// AND unix_timestamp(homework_begin_time) <= unix_timestamp(NOW()) AND unix_timestamp(homework_due_time) >= unix_timestamp(NOW())
	res, err := db.Query(`SELECT homework_id, homework_name, homework_begin_time, homework_due_time, homework_description FROM oj_homework WHERE course_id=? AND unix_timestamp(homework_begin_time) <= unix_timestamp(NOW());`, CourseID)
	if err != nil {
		logging.Info(err)
		return nil, err
	}
	defer res.Close()
	info := []Assignment{}
	for res.Next() {
		var assignment Assignment
		err := res.Scan(&assignment.ID, &assignment.Title, &assignment.BeginTime, &assignment.DueTime, &assignment.Description)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		info = append(info, assignment)
	}
	// read assignment related questions
	for index, assignment := range info {
		res, err := db.Query("SELECT question_id FROM oj_question_homework WHERE homework_id=?;", assignment.ID)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		defer res.Close()
		for res.Next() {
			var questionID uint32
			err := res.Scan(&questionID)
			if err != nil {
				logging.Info(err)
				return nil, err
			}
			var question AssignmentQuestionInfo
			err = db.QueryRow("SELECT question_id, question_name, question_tag, question_score, question_testcase_num FROM oj_question WHERE question_id=?;", questionID).Scan(
				&question.ID,
				&question.Title,
				&question.Tag,
				&question.TotalScore,
				&question.TestCaseNum,
			)
			info[index].Questions = append(info[index].Questions, question)
		}
	}
	// read question test-run infos to get current score
	// test_run_infos := []TestRunInfo{}
	for idxx, assignment := range info {
		for idxy, question := range assignment.Questions {
			test_run_info := TestRunInfo{}
			test_run_info.QuestionID = question.ID
			test_run_info.TestID = 0
			res, err := db.Query("SELECT test_id, test_launch_time FROM oj_test_run WHERE course_id=? AND question_id=? AND user_id=? AND unix_timestamp(test_launch_time) <= unix_timestamp(?) ORDER BY test_launch_time desc;", CourseID, question.ID, UserID, assignment.DueTime)
			if err != nil {
				logging.Info(err)
				return nil, err
			}
			defer res.Close()
			for res.Next() {
				err = res.Scan(&test_run_info.TestID, &test_run_info.LaunchTime)
				if err != nil {
					logging.Info(err)
					return nil, err
				}
				break
				//test_run_infos = append(test_run_infos, test_run_info)
			}
			// get latest test cases run result
			if test_run_info.TestID <= 0 {
				info[idxx].Questions[idxy].Score = 0
			} else {
				res, err = db.Query("SELECT testcase_run_state FROM oj_testcase_run WHERE test_id=?;", test_run_info.TestID)
				if err != nil {
					logging.Info(err)
					return nil, err
				}
				var ac_count uint32 = 0;
				for res.Next() {
					var state string
					err = res.Scan(&state)
					if err != nil {
						logging.Info(err)
						return nil, err
					}
					if state == "Accepted" {
						ac_count += 1
					}
				}
				info[idxx].Questions[idxy].Score = (ac_count * info[idxx].Questions[idxy].TotalScore) / info[idxx].Questions[idxy].TestCaseNum;
			}
		}
	}
	return info, nil
}
