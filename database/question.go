package database

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
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
	Tag         string `json:"tag" form:"tag" uri:"tag"`
	Language    string `json:"language" form:"language" uri:"language" binding:"required"`
	TotalScore  uint32 `json:"totalScore" form:"totalScore" uri:"totalScore" binding:"required"`
}

type Question struct {
	ID           uint32
	Title        string
	Tag          string
	Creator      string
	Score        uint32
	UserMaxScore uint32
	TestCaseNum  uint32
	MemoryLimit  uint32
	TimeLimit    uint32
	Language     string
	TotalTestNum uint32
	TotalACNum   uint32
	IssueTime    time.Time
}

type QuestionInfo struct {
	ID           uint32
	Title        string
	Tag          string
	Creator      string
	Score        string
	TestCaseNum  uint32
	MemoryLimit  uint32
	TimeLimit    uint32
	Language     string
	TotalTestNum uint32
	TotalACNum   uint32
	IssueTime    string
	Content      string
	AppendixFile string
}

func CreateQuestion(questionForm CreateQuestionForm, creatorID uint32, testCaseNum uint32) (uint32, error) {
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
		creatorID,
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
	// get question id
	questionID, err := result.LastInsertId()
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
		questionID,
		questionForm.CourseID,
	)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	// add user question relation
	// get course users
	var userIDs = []uint32{}
	rows, err := tx.Query("SELECT user_id FROM oj_user_course WHERE course_id=?;", questionForm.CourseID)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	for rows.Next() {
		var userID uint32
		err := rows.Scan(&userID)
		if err != nil {
			_ = tx.Rollback()
			logging.Info(err)
			return 0, err
		}
		userIDs = append(userIDs, userID)
	}
	// create user <-> anno relation
	var insertUserQuestion = "INSERT INTO oj_user_question(user_id, question_id) VALUES "
	for idx, userID := range userIDs {
		if idx < len(userIDs)-1 {
			insertUserQuestion += fmt.Sprintf("(%d, %d),", userID, questionID)
		} else {
			insertUserQuestion += fmt.Sprintf("(%d, %d);", userID, questionID)
		}
	}
	_, err = tx.Exec(insertUserQuestion)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return 0, err
	}
	_ = tx.Commit()
	logging.Info("CreateQuestion() commit trans action successfully. ID: ", questionID)
	return uint32(questionID), nil
}

func GetQuestionTotalNumByCourseID(courseID uint32) (uint32, error) {
	res, err := db.Query("SELECT COUNT(*) FROM oj_question_course WHERE course_id=?;", courseID)
	if err != nil {
		logging.Info(err)
		return 0, err
	}
	defer res.Close()
	var num uint32
	for res.Next() {
		err := res.Scan(&num)
		if err != nil {
			logging.Info(err)
			continue
		}
	}
	return num, nil
}

func GetQuestionsByCourseID(courseID uint32, userID uint32, limit uint32, offset uint32) ([]Question, error) {
	res, err := db.Query("SELECT question_id FROM oj_question_course WHERE course_id=? LIMIT ? OFFSET ?;", courseID, limit, offset)
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
		// read testid and get max score
		err = db.QueryRow("SELECT score FROM oj_test_run WHERE course_id=? AND question_id=? AND user_id=? ORDER BY score DESC;", courseID, question.ID, userID).Scan(
			&question.UserMaxScore,
		)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				question.UserMaxScore = 0
			} else {
				logging.Info(err)
				return nil, err
			}
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

func GetQuestionTitleAndTestCaseNumAndLanguageByID(questionID uint32) (string, uint32, string, error) {
	var title string
	var num uint32
	var language string
	err := db.QueryRow("SELECT question_name, question_testcase_num, question_language FROM oj_question WHERE question_id=?;", questionID).Scan(&title, &num, &language)
	return title, num, language, err
}

func GetQuestionDetailByID(questionID, userID uint32) (QuestionInfo, error) {
	question := QuestionInfo{}
	var userid uint32
	var issueTime time.Time
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
		&issueTime,
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
	question.IssueTime = issueTime.Format("2006/01/02 15:04")
	logging.Info(question)
	// read description from disk
	filesDir := setting.QuestionRootDir + strconv.FormatUint(uint64(questionID), 10) + "_" + question.Title + "/"
	f, err := os.Open(filesDir + "description.md")
	if err != nil {
		logging.Info(err)
		return question, err
	}
	defer f.Close()
	content, err := io.ReadAll(f)
	if err != nil {
		logging.Info(err)
		return question, err
	}
	question.Content = string(content)
	// check appendix
	if utils.FileExists(filesDir + "appendix.zip") {
		question.AppendixFile = "appendix.zip"
	} else {
		question.AppendixFile = ""
	}
	question.ID = questionID
	// update access info
	_, err = db.Exec("UPDATE oj_user_question SET access_count=access_count+1 WHERE question_id=? AND user_id=?;", questionID, userID)
	if err != nil {
		logging.Info(err)
		return question, err
	}
	return question, nil
}

func GetQuestionAppendixPath(questionID uint32) (string, error) {
	var title string
	err := db.QueryRow("SELECT question_name FROM oj_question WHERE question_id=?;", questionID).Scan(
		&title,
	)
	if err != nil {
		return "", err
	}
	appendixPath := setting.QuestionRootDir + strconv.FormatUint(uint64(questionID), 10) + "_" + title + "/appendix.zip"
	return appendixPath, nil
}

func UserAccessAppendix(questionID, userID uint32) error {
	// update access info
	_, err := db.Exec("UPDATE oj_user_question SET download_count=download_count+1 WHERE question_id=? AND user_id=?;", questionID, userID)
	if err != nil {
		logging.Info(err)
		return err
	}
	return nil
}

func CheckQuestionAccessPermission(questionID, userID uint32) bool {
	var count uint32
	err := db.QueryRow("SELECT access_count FROM oj_user_question WHERE user_id=? AND question_id=?;", userID, questionID).Scan(
		&count,
	)
	return err == nil
}
