package database

import (
	"fmt"
	"log"
)

type CreateCourseForm struct {
	CourseName string   `json:"name" form:"name" uri:"name" binding:"required"`
	CourseType string   `json:"type" form:"type" uri:"type" binding:"required"`
	Creator    string   `json:"creator" form:"creator" uri:"creator" binding:"required"`
	Term       string   `json:"term" form:"term" uri:"term" binding:"required"`
	Teachers   []uint32 `json:"teachers" form:"teachers" uri:"teachers" binding:"required"`
	Students   []uint32 `json:"students" form:"students" uri:"students" binding:"required"`
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
	_, err = tx.Exec(`INSERT INTO oj_db_test.oj_course
		(course_name, course_teacher, course_term, course_type)
		VALUES
		(?, ?, ?, ?);
	`,
		courseform.CourseName,
		courseform.Creator,
		courseform.Term,
		courseform.CourseType,
	)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return err
	}
	// get course id
	var course_id uint32
	err = tx.QueryRow("SELECT course_id from oj_db_test.oj_course where course_name=?;", courseform.CourseName).Scan(&course_id)
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
	for index, studentID := range courseform.Students {
		var existInTeacher bool = false;
		for _, teacherID := range courseform.Teachers {
			if teacherID == studentID {
				existInTeacher = true;
				break;
			}
		}
		if existInTeacher {
			continue;
		}
		if index < len(courseform.Students)-1 {
			insertCourseStudent += fmt.Sprintf( "(%d, %d,'%s'),", course_id, studentID, "student")
		} else {
			insertCourseStudent += fmt.Sprintf( "(%d, %d,'%s');", course_id, studentID, "student")
		}
	}
	_, err = tx.Exec(insertCourseStudent)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return err
	}
	_ = tx.Commit()
	log.Printf("CreateNewCourse() commit trans action successfully.")
	return nil
}


type Course struct {
	CourseID      uint32
	CourseName    string
	CourseTeacher string
	CourseTerm    string
	CourseType    string
}

func GetUserCourses(userid uint32) ([]Course, error) {
	// read oj_user_course to get user-related courses' ids
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		log.Printf("GetUserCourses() begin trans action failed, err:%v\n", err)
		return nil, err
	}
	courseIDs := []uint32{}
	if !checkTableExists("oj_user_course") {
		_ = tx.Rollback()
		log.Println(err)
		return nil, err
	}
	res, err := db.Query("SELECT course_id FROM oj_db_test.oj_user_course where user_id=?;", userid)
	if err != nil {
		_ = tx.Rollback()
		log.Println(err)
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var courseID uint32
		err := res.Scan(&courseID)
		if err != nil {
			_ = tx.Rollback()
			log.Println(err)
			return nil, err
		}
		courseIDs = append(courseIDs, courseID)
	}
	// get courses info
	courses := []Course{}
	for _, course_id := range courseIDs {
		var course Course
		err := tx.QueryRow("SELECT course_id, course_name, course_teacher, course_term, course_type from oj_db_test.oj_course where course_id=?;", course_id).Scan(&course.CourseID, &course.CourseName, &course.CourseTeacher, &course.CourseTerm, &course.CourseType)
		if err != nil {
			_ = tx.Rollback()
			log.Println(err)
			return nil, err
		}
		courses = append(courses, course)
	}
	_ = tx.Commit()
	log.Printf("GetUserCourses() commit trans action successfully.")
	return courses, nil
}
