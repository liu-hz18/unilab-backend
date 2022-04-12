package database

import (
	"fmt"
	"log"
)

const (
	CourseTypeInvalid uint8 = 0
	CourseTypeOJ      uint8 = 1
	CourseTypeGitLab  uint8 = 2
	CourseTypeSystem  uint8 = 3
)

type StudentInfo struct {
	ID uint32 `json:"id" form:"id" uri:"id" binding:"required"`
	Name string `json:"name" form:"name" uri:"name" binding:"required"`
}

type CreateCourseForm struct {
	CourseName string   `json:"name" form:"name" uri:"name" binding:"required"`
	CourseType int      `json:"type" form:"type" uri:"type" binding:"required"`
	Creator    string   `json:"creator" form:"creator" uri:"creator" binding:"required"`
	Term       string   `json:"term" form:"term" uri:"term" binding:"required"`
	Teachers   []uint32 `json:"teachers" form:"teachers" uri:"teachers" binding:"required"`
	Students   []StudentInfo `json:"students" form:"students" uri:"students" binding:"required"`
}

func CreateNewCourse(courseform CreateCourseForm) error {
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		log.Printf("CreateNewCourse() begin trans action failed, err:%v\n", err)
		return err
	}
	// insert new course
	result, err := tx.Exec(`INSERT INTO oj_db_test.oj_course
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
		log.Println(err)
		return err
	}
	// get course id
	course_id, err := result.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return err
	}
	// create students if not existed
	// WARN: teachers are not allowed to add teachers
	var insertSqlStr string = "INSERT IGNORE INTO oj_db_test.oj_user (user_id, user_real_name, user_type) VALUES "
	for index, user_info := range courseform.Students {
		if index < len(courseform.Students)-1 {
			insertSqlStr += fmt.Sprintf("(%d, '%s', %d),", user_info.ID, user_info.Name, UserStudent)
		} else {
			insertSqlStr += fmt.Sprintf("(%d, '%s', %d);", user_info.ID, user_info.Name, UserStudent)
		}
	}
	_, err = tx.Exec(insertSqlStr)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return err
	}
	// add course <-> user relation
	var insertCourseTeacher string = "INSERT INTO oj_db_test.oj_user_course(course_id, user_id, user_type) VALUES "
	for index, teacherID := range courseform.Teachers {
		if index < len(courseform.Teachers)-1 {
			insertCourseTeacher += fmt.Sprintf( "(%d, %d,'%s'),", course_id, teacherID, "teacher")
		} else {
			insertCourseTeacher += fmt.Sprintf( "(%d, %d,'%s');", course_id, teacherID, "teacher")
		}
	}
	log.Println(insertCourseTeacher)
	_, err = tx.Exec(insertCourseTeacher)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return err
	}
	var insertCourseStudent string = "INSERT INTO oj_db_test.oj_user_course(course_id, user_id, user_type) VALUES "
	var haveStudents bool = false
	for index, student := range courseform.Students {
		var existInTeacher bool = false;
		for _, teacherID := range courseform.Teachers {
			if teacherID == student.ID {
				existInTeacher = true;
				break;
			}
		}
		if existInTeacher {
			continue;
		}
		if index < len(courseform.Students)-1 {
			insertCourseStudent += fmt.Sprintf( "(%d, %d,'%s'),", course_id, student.ID, "student")
		} else {
			insertCourseStudent += fmt.Sprintf( "(%d, %d,'%s');", course_id, student.ID, "student")
			haveStudents = true
		}
	}
	if haveStudents {
		_, err = tx.Exec(insertCourseStudent)
		if err != nil {
			_ = tx.Rollback()
			log.Println(err)
			return err
		}
	}
	_ = tx.Commit()
	log.Printf("CreateNewCourse() commit trans action successfully.")
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
	res, err := db.Query("SELECT course_id FROM oj_db_test.oj_user_course where user_id=?;", userid)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var courseID uint32
		err := res.Scan(&courseID)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		courseIDs = append(courseIDs, courseID)
	}
	// get courses info
	courses := []Course{}
	for _, course_id := range courseIDs {
		var course Course
		err := db.QueryRow(
			"SELECT course_id, course_name, course_teacher, course_term, course_type FROM oj_db_test.oj_course WHERE course_id=?;",
			course_id).Scan(
				&course.CourseID,
				&course.CourseName,
				&course.CourseTeacher,
				&course.CourseTerm,
				&course.CourseType,
		)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		totalRow, err := db.Query("SELECT COUNT(*) FROM oj_db_test.oj_announcement WHERE course_id=?;", course_id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer totalRow.Close()
		var totalAnno uint32 = 0
		for totalRow.Next() {
			err := totalRow.Scan(&totalAnno)
			if err != nil {
				log.Println(err)
				continue
			}
		}
		course.CourseAnnounce = totalAnno
		totalRow, err = db.Query("SELECT homework_due_time FROM oj_db_test.oj_homework WHERE course_id=? AND unix_timestamp(homework_begin_time) <= unix_timestamp(NOW()) AND unix_timestamp(NOW()) <= unix_timestamp(homework_due_time) ORDER BY homework_due_time DESC;", course_id)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		defer totalRow.Close()
		var totalAssignment uint32 = 0
		var nearestDue string = ""
		for totalRow.Next() {
			err := totalRow.Scan(&nearestDue)
			if err != nil {
				log.Println(err)
				continue
			}
			totalAssignment += 1
		}
		course.CourseAssignments = totalAssignment
		course.NearestDue = nearestDue
		courses = append(courses, course)
	}
	log.Printf("GetUserCourses() commit trans action successfully.")
	return courses, nil
}

func GetCourseByID(courseID uint32) (string, error) {
	var course_name string
	err := db.QueryRow("SELECT course_name from oj_db_test.oj_course where course_id=?;", courseID).Scan(&course_name)
	if err != nil {
		log.Println(err)
		return "", err
	}
	return course_name, nil
}


func CheckCourseAccessPermission(courseID, userID uint32) bool {
	var course_id uint32
	err := db.QueryRow("SELECT course_id FROM oj_db_test.oj_user_course WHERE course_id=? AND user_id=?;", courseID, userID).Scan(&course_id)
	if err != nil {
		log.Println(err)
		return false;
	}
	return true;
}
