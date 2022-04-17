package database

import (
	"io/ioutil"
	"os"
	"strconv"
	"time"
	"unilab-backend/logging"
	"unilab-backend/setting"
	"unilab-backend/utils"
)

type CreateQuestionForm struct {
	Title       string `json:"title" form:"title" uri:"title" binding:"required"`
	CourseID    uint32 `json:"courseid" form:"courseid" uri:"courseid" binding:"required"`
	TimeLimit   uint32 `json:"timeLimit" form:"timeLimit" uri:"timeLimit" binding:"required"`
	MemoryLimit uint32 `json:"memoryLimit" form:"memoryLimit" uri:"memoryLimit" binding:"required"`
	Tag         string `json:"tag" form:"tag" uri:"tag" binding:"required"`
	Language    string `json:"language" form:"language" uri:"language" binding:"required"`
	TotalScore  uint32 `json:"totalScore" form:"totalScore" uri:"totalScore" binding:"required"`
}

type Question struct {
	ID uint32
	Title string
	Tag string
	Creator string
	Score string
	TestCaseNum uint32
	MemoryLimit uint32
	TimeLimit uint32
	Language string
	TotalTestNum uint32
	TotalACNum uint32
	IssueTime time.Time
}

type QuestionInfo struct {
	ID uint32
	Title string
	Tag string
	Creator string
	Score string
	TestCaseNum uint32
	MemoryLimit uint32
	TimeLimit uint32
	Language string
	TotalTestNum uint32
	TotalACNum uint32
	IssueTime string
	Content string
	AppendixFile string
}

func CreateQuestion(questionForm CreateQuestionForm, creator_id uint32, testCaseNum uint32) (uint32, error) {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("CreateQuestion() begin trans action failed: %v", err)
	}
	// insert a new question
	result, err := tx.Exec(`INSERT INTO oj_question
		(question_name, question_tag, question_creator, question_score, question_testcase_num, question_memory_limit, question_time_limit, question_language, question_compile_options, question_test_total_num, question_test_ac_num, issue_time)
		VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`,
		questionForm.Title,
		questionForm.Tag,
		creator_id,
		questionForm.TotalScore,
		testCaseNum,
		questionForm.MemoryLimit,
		questionForm.TimeLimit,
		questionForm.Language,
		"",
		0,
		0,
		time.Now(),
	)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	// get announcement id
	question_id, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	// oj_question_course
	_, err = tx.Exec(`INSERT INTO oj_question_course
		(question_id, course_id)
		VALUES
		(?, ?)
	`,
		question_id,
		questionForm.CourseID,
	)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	_ = tx.Commit()
	logging.Info("CreateQuestion() commit trans action successfully. ID: ", question_id)
	return uint32(question_id), nil
}


func GetQuestionsByCourseID(courseID uint32) ([]Question, error) {
	res, err := db.Query("SELECT question_id FROM oj_question_course WHERE course_id=?;", courseID)
	if err != nil {
		logging.Info(err)
		return nil, err
	}
	defer res.Close()
	questions := []Question{}
	for res.Next() {
		var question Question
		err := res.Scan(&question.ID)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		var userid uint32
		err = db.QueryRow("SELECT question_name, question_tag, question_creator, question_score, question_testcase_num, question_memory_limit, question_time_limit, question_language, question_test_total_num, question_test_ac_num, issue_time FROM oj_question WHERE question_id=?;", question.ID).Scan(
			&question.Title,
			&question.Tag,
			&userid,
			&question.Score,
			&question.TestCaseNum,
			&question.MemoryLimit,
			&question.TimeLimit,
			&question.Language,
			&question.TotalTestNum,
			&question.TotalACNum,
			&question.IssueTime,
		)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		username, err := GetUserName(strconv.FormatUint(uint64(userid), 10))
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		question.Creator = username
		questions = append(questions, question)
	}
	logging.Info("GetQuestionsByCourseID() commit trans action successfully.")
	return questions, nil
}

func GetQuestionTitleAndTestCaseNumByID(questionID uint32) (string, uint32, error) {
	var title string
	var num uint32
	err := db.QueryRow("SELECT question_name, question_testcase_num FROM oj_question WHERE question_id=?;", questionID).Scan(&title, &num)
	return title, num, err
}

func GetQuestionDetailByID(questionID uint32) (QuestionInfo, error) {
	question := QuestionInfo{}
	var userid uint32
	var issue_time time.Time
	question.ID = 0
	err := db.QueryRow("SELECT question_name, question_tag, question_creator, question_score, question_testcase_num, question_memory_limit, question_time_limit, question_language, question_test_total_num, question_test_ac_num, issue_time FROM oj_question WHERE question_id=?;", questionID).Scan(
		&question.Title,
		&question.Tag,
		&userid,
		&question.Score,
		&question.TestCaseNum,
		&question.MemoryLimit,
		&question.TimeLimit,
		&question.Language,
		&question.TotalTestNum,
		&question.TotalACNum,
		&issue_time,
	)
	if err != nil {
		logging.Info(err)
		return question, err
	}
	username, err := GetUserName(strconv.FormatUint(uint64(userid), 10))
	if err != nil {
		logging.Info(err)
		return question, err
	}
	question.Creator = username
	question.IssueTime = issue_time.Format("2006/01/02 15:04")
	logging.Info(question)
	// read description from disk
	files_dir := setting.QuestionRootDir + strconv.FormatUint(uint64(questionID), 10) + "_" + question.Title + "/"
	f, err := os.Open(files_dir + "description.md")
	if err != nil {
		logging.Info(err)
		return question, err
	}
	defer f.Close()
 	content, err := ioutil.ReadAll(f)
	if err != nil {
		logging.Info(err)
		return question, err
	}
	question.Content = string(content)
	// check appendix
	if utils.FileExists(files_dir + "appendix.zip") {
		question.AppendixFile = files_dir + "appendix.zip"
	} else {
		question.AppendixFile = ""
	}
	question.ID = questionID
	return question, nil
}
