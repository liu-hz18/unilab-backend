package database

import (
	"fmt"
	"unilab-backend/logging"
)

const (
	CourseTypeInvalid uint8 = 0
	CourseTypeOJ      uint8 = 1
	CourseTypeGitLab  uint8 = 2
	CourseTypeSystem  uint8 = 3
)

type StudentInfo struct {
	ID   uint32 `json:"id" form:"id" uri:"id" binding:"required"`
	Name string `json:"name" form:"name" uri:"name" binding:"required"`
}

type CreateCourseForm struct {
	CourseName string        `json:"name" form:"name" uri:"name" binding:"required"`
	CourseType int           `json:"type" form:"type" uri:"type" binding:"required"`
	Creator    string        `json:"creator" form:"creator" uri:"creator" binding:"required"`
	Term       string        `json:"term" form:"term" uri:"term" binding:"required"`
	Teachers   []uint32      `json:"teachers" form:"teachers" uri:"teachers" binding:"required"`
	Students   []StudentInfo `json:"students" form:"students" uri:"students" binding:"required"`
}

func CreateNewCourse(courseform CreateCourseForm) error {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("CreateNewCourse() begin trans action failed, err:", err)
		return err
	}
	// insert new course
	result, err := tx.Exec(`INSERT INTO oj_course
		(course_name, course_teacher, course_term, course_type)
		VALUES
		(?, ?, ?, ?);
	`,
		courseform.CourseName,
		courseform.Creator,
		courseform.Term,
		uint8(courseform.CourseType),
	)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	// get course id
	courseID, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	// create students if not existed
	// WARN: teachers are not allowed to add teachers
	var insertSQLStr = "INSERT IGNORE INTO oj_user (user_id, user_real_name, user_type) VALUES "
	for index, userInfo := range courseform.Students {
		if index < len(courseform.Students)-1 {
			insertSQLStr += fmt.Sprintf("(%d, '%s', %d),", userInfo.ID, userInfo.Name, UserStudent)
		} else {
			insertSQLStr += fmt.Sprintf("(%d, '%s', %d);", userInfo.ID, userInfo.Name, UserStudent)
		}
	}
	_, err = tx.Exec(insertSQLStr)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	// add course <-> user relation
	var insertCourseTeacher = "INSERT INTO oj_user_course(course_id, user_id, user_type) VALUES "
	for index, teacherID := range courseform.Teachers {
		if index < len(courseform.Teachers)-1 {
			insertCourseTeacher += fmt.Sprintf("(%d, %d,'%s'),", courseID, teacherID, "teacher")
		} else {
			insertCourseTeacher += fmt.Sprintf("(%d, %d,'%s');", courseID, teacherID, "teacher")
		}
	}
	// logging.Info(insertCourseTeacher)
	_, err = tx.Exec(insertCourseTeacher)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	var insertCourseStudent = "INSERT INTO oj_user_course(course_id, user_id, user_type) VALUES "
	var haveStudents = false
	for index, student := range courseform.Students {
		var existInTeacher = false
		for _, teacherID := range courseform.Teachers {
			if teacherID == student.ID {
				existInTeacher = true
				break
			}
		}
		if existInTeacher {
			continue
		}
		if index < len(courseform.Students)-1 {
			insertCourseStudent += fmt.Sprintf("(%d, %d,'%s'),", courseID, student.ID, "student")
		} else {
			insertCourseStudent += fmt.Sprintf("(%d, %d,'%s');", courseID, student.ID, "student")
			haveStudents = true
		}
	}
	if haveStudents {
		_, err = tx.Exec(insertCourseStudent)
		if err != nil {
			_ = tx.Rollback()
			logging.Info(err)
			return err
		}
	}
	_ = tx.Commit()
	logging.Info("CreateNewCourse() commit trans action successfully.")
	return nil
}

type Course struct {
	CourseID          uint32
	CourseName        string
	CourseTeacher     string
	CourseTerm        string
	CourseType        uint8
	CourseAnnounce    uint32
	CourseAssignments uint32
	NearestDue        string
}

func GetUserCourses(userid uint32) ([]Course, error) {
	// read oj_user_course to get user-related courses' ids
	courseIDs := []uint32{}
	res, err := db.Query("SELECT course_id FROM oj_user_course where user_id=?;", userid)
	if err != nil {
		logging.Info(err)
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var courseID uint32
		err := res.Scan(&courseID)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		courseIDs = append(courseIDs, courseID)
	}
	// get courses info
	courses := []Course{}
	for _, courseID := range courseIDs {
		var course Course
		err := db.QueryRow(
			"SELECT course_id, course_name, course_teacher, course_term, course_type FROM oj_course WHERE course_id=?;",
			courseID).Scan(
			&course.CourseID,
			&course.CourseName,
			&course.CourseTeacher,
			&course.CourseTerm,
			&course.CourseType,
		)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		totalRow, err := db.Query("SELECT COUNT(*) FROM oj_announcement WHERE course_id=?;", courseID)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		defer totalRow.Close()
		var totalAnno uint32
		for totalRow.Next() {
			err := totalRow.Scan(&totalAnno)
			if err != nil {
				logging.Info(err)
				continue
			}
		}
		course.CourseAnnounce = totalAnno
		totalRow, err = db.Query("SELECT homework_due_time FROM oj_homework WHERE course_id=? AND unix_timestamp(homework_begin_time) <= unix_timestamp(NOW()) AND unix_timestamp(NOW()) <= unix_timestamp(homework_due_time) ORDER BY homework_due_time DESC;", courseID)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		defer totalRow.Close()
		var totalAssignment uint32
		nearestDue := ""
		for totalRow.Next() {
			err := totalRow.Scan(&nearestDue)
			if err != nil {
				logging.Info(err)
				continue
			}
			totalAssignment++
		}
		course.CourseAssignments = totalAssignment
		course.NearestDue = nearestDue
		courses = append(courses, course)
	}
	logging.Info("GetUserCourses() commit trans action successfully.")
	return courses, nil
}

func GetCourseByID(courseID uint32) (string, error) {
	var courseName string
	err := db.QueryRow("SELECT course_name from oj_course where course_id=?;", courseID).Scan(&courseName)
	if err != nil {
		logging.Info(err)
		return "", err
	}
	return courseName, nil
}

func CheckCourseAccessPermission(courseID, userID uint32) bool {
	var courseIDDB uint32
	err := db.QueryRow("SELECT course_id FROM oj_user_course WHERE course_id=? AND user_id=?;", courseID, userID).Scan(&courseIDDB)
	if err != nil {
		logging.Info(err)
		return false
	}
	return true
}

func UserAccessCourse(courseID, userID uint32) error {
	_, err := db.Exec("UPDATE oj_user_course SET access_count=access_count+1 WHERE course_id=? AND user_id=?;", courseID, userID)
	if err != nil {
		logging.Info(err)
		return err
	}
	return nil
}

func GetCourseUsers(courseID uint32) ([]uint32, error) {
	var userIDs = []uint32{}
	rows, err := db.Query("SELECT user_id FROM oj_user_course WHERE course_id=?;", courseID)
	if err != nil {
		logging.Info(err)
		return userIDs, err
	}
	for rows.Next() {
		var userID uint32
		err := rows.Scan(&userID)
		if err != nil {
			logging.Info(err)
			return userIDs, err
		}
		userIDs = append(userIDs, userID)
	}
	return userIDs, nil
}
