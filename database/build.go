package database

import (
	"database/sql"
	"log"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func InitDB() {
	_db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/mysql?charset=utf8&parseTime=true&loc=Local")
	if err != nil {
		log.Fatal(err)
		return
	}
	db = _db
	// verify connection
	if err = db.Ping(); err != nil {
		log.Fatal(err)
		return
	}
	log.Println("db connection established.")
	// 静态低频查询信息
	// - 多对多: user table, course table, 建立关联表，总共3个表
	// - 多对多: question table, homework table, 建立关联表，总共3个
	// - 多对多: course table, question table
	// - 一对多: course table, homework table
	// 一对多: file table -> question table, user table
	// 动态高频查询信息, 和评测相关
	// test-run table: user, file, question, course
	// user log
	
	drop_test_database()
	drop_file_dir()

	// create database
	_, err = db.Exec("CREATE DATABASE IF NOT EXISTS oj_db_test;")
	if err != nil {
		log.Fatal(err)
		return
	}
	_, err = db.Exec("USE oj_db_test;")
	if err != nil {
		log.Fatal(err)
		return
	}
	// create user table
	// user_type: [student, teacher, admin]
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_user(
		user_id INT(10) UNSIGNED NOT NULL PRIMARY KEY,
		user_name VARCHAR(16) NOT NULL, 
		user_real_name VARCHAR(50) NOT NULL,
		user_email VARCHAR(255) NOT NULL,
		user_git_tsinghua_id INT UNSIGNED NOT NULL,
		user_last_login_time DATETIME NOT NULL,
		user_type TINYINT UNSIGNED NOT NULL,
		user_token VARCHAR(255) NOT NULL DEFAULT ''
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create course table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_course(
		course_id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		course_name VARCHAR(32) NOT NULL,
		course_teacher VARCHAR(32) NOT NULL,
		course_term VARCHAR(64) NOT NULL,
		course_type TINYINT UNSIGNED NOT NULL
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create user <-> course joint table
	// user_type: admin, teacher, student
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_user_course(
		course_id INT UNSIGNED NOT NULL,
		user_id INT(10) UNSIGNED NOT NULL,
		user_type VARCHAR(255) NOT NULL,
		CONSTRAINT c_oj_user_course_1 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE,
		CONSTRAINT c_oj_user_course_2 FOREIGN KEY (user_id) REFERENCES oj_user(user_id) ON DELETE CASCADE ON UPDATE CASCADE
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create question table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_question(
		question_id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		question_name VARCHAR(255) NOT NULL,
		question_tag VARCHAR(255) NOT NULL,
		question_creator INT(10) UNSIGNED NOT NULL,
		question_score DECIMAL NOT NULL,
		question_testcase_num INT UNSIGNED NOT NULL,
		question_memory_limit INT UNSIGNED NOT NULL,
		question_time_limit INT UNSIGNED NOT NULL,
		question_language VARCHAR(255) NOT NULL,
		question_compile_options VARCHAR(255) NOT NULL,
		question_test_total_num INT UNSIGNED NOT NULL,
		question_test_ac_num INT UNSIGNED NOT NULL,
		issue_time DATETIME NOT NULL,
		CONSTRAINT c_oj_question FOREIGN KEY (question_creator) REFERENCES oj_user(user_id) ON DELETE CASCADE ON UPDATE CASCADE
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create homework table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_homework(
		homework_id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		homework_name VARCHAR(255) NOT NULL,
		homework_begin_time DATETIME NOT NULL,
		homework_due_time DATETIME NOT NULL,
		homework_description VARCHAR(255) default '',
		course_id INT UNSIGNED NOT NULL,
		CONSTRAINT c_oj_homework_1 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create question <-> homework table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_question_homework(
		question_id INT UNSIGNED NOT NULL,
		homework_id INT UNSIGNED NOT NULL,
		CONSTRAINT c_oj_question_homework_1 FOREIGN KEY (question_id) REFERENCES oj_question(question_id) ON DELETE CASCADE ON UPDATE CASCADE,
		CONSTRAINT c_oj_question_homework_2 FOREIGN KEY (homework_id) REFERENCES oj_homework(homework_id) ON DELETE CASCADE ON UPDATE CASCADE
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create question <-> course table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_question_course(
		question_id INT UNSIGNED NOT NULL,
		course_id INT UNSIGNED NOT NULL,
		CONSTRAINT c_oj_question_course_1 FOREIGN KEY (question_id) REFERENCES oj_question(question_id) ON DELETE CASCADE ON UPDATE CASCADE,
		CONSTRAINT c_oj_question_course_2 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE 
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create announcement table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_announcement(
		announcement_id INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		announcement_title VARCHAR(255) NOT NULL,
		course_id INT UNSIGNED NOT NULL,
		issue_time DATETIME NOT NULL,
		CONSTRAINT c_oj_announcement_1 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create test-run table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_test_run(
		test_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		test_launch_time DATETIME NOT NULL,
		course_id INT UNSIGNED NOT NULL,
		question_id INT UNSIGNED NOT NULL,
		test_case_num INT UNSIGNED NOT NULL,
		upload_file_path VARCHAR(255) NOT NULL,
		user_id INT(10) UNSIGNED NOT NULL,
		CONSTRAINT c_oj_test_run_1 FOREIGN KEY (course_id) REFERENCES oj_course(course_id) ON DELETE CASCADE ON UPDATE CASCADE,
		CONSTRAINT c_oj_test_run_2 FOREIGN KEY (question_id) REFERENCES oj_question(question_id) ON DELETE CASCADE ON UPDATE CASCADE,
		CONSTRAINT c_oj_test_run_3 FOREIGN KEY (user_id) REFERENCES oj_user(user_id) ON DELETE CASCADE ON UPDATE CASCADE
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create testcase-run table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_testcase_run(
		testcase_run_id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
		testcase_run_state VARCHAR(255) NOT NULL,
		testcase_run_time_elapsed INT UNSIGNED NOT NULL,
		testcase_run_memory_usage INT UNSIGNED NOT NULL,
		test_id BIGINT UNSIGNED NOT NULL,
		CONSTRAINT c_oj_testcase_run_1 FOREIGN KEY (test_id) REFERENCES oj_test_run(test_id) ON DELETE CASCADE ON UPDATE CASCADE
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	// create question <-> user table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS oj_question_user(
		question_id INT UNSIGNED NOT NULL,
		user_id INT UNSIGNED NOT NULL,
		latest_score INT UNSIGNED NOT NULL,
		latest_test_id BIGINT UNSIGNED NOT NULL,
		best_score INT UNSIGNED NOT NULL,
		best_test_id BIGINT UNSIGNED NOT NULL,
		launch_times INT UNSIGNED NOT NULL,
		CONSTRAINT c_oj_question_user_1 FOREIGN KEY (question_id) REFERENCES oj_question(question_id) ON DELETE CASCADE ON UPDATE CASCADE,
		CONSTRAINT c_oj_question_user_2 FOREIGN KEY (user_id) REFERENCES oj_user(user_id) ON DELETE CASCADE ON UPDATE CASCADE,
		CONSTRAINT c_oj_question_user_3 FOREIGN KEY (latest_test_id) REFERENCES oj_test_run(test_id) ON DELETE CASCADE ON UPDATE CASCADE,
		CONSTRAINT c_oj_question_user_4 FOREIGN KEY (best_test_id) REFERENCES oj_test_run(test_id) ON DELETE CASCADE ON UPDATE CASCADE
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`)
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("test db successfully created")
}

// just for test
func PreinitDBTestData() {
	// create user
	_, err := db.Exec(`INSERT INTO oj_user
		(user_id, user_name, user_real_name, user_email, user_git_tsinghua_id, user_last_login_time, user_type)
		VALUES
		(?, ?, ?, ?, ?, ?, ?);
	`,
		2018011446,
		"admin",
		"test admin",
		"admin@mails.tsinghua.edu.cn",
		123456,
		"1970-12-31 23:59:59",
		UserAdmin,
	)
	if err != nil {
		log.Fatal(err)
		return
	}
	for i := 0; i < 50; i++ {
		_, err := db.Exec(`INSERT INTO oj_user
			(user_id, user_name, user_real_name, user_email, user_git_tsinghua_id, user_last_login_time, user_type)
			VALUES
			(?, ?, ?, ?, ?, ?, ?);
		`,
			2019011446+i,
			"student"+strconv.Itoa(i),
			"test student"+strconv.Itoa(i),
			strconv.Itoa(i)+"@mails.tsinghua.edu.cn",
			123456,
			"1970-12-31 23:59:59",
			UserStudent,
		)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	for i := 50; i < 100; i++ {
		_, err := db.Exec(`INSERT INTO oj_user
			(user_id, user_name, user_real_name, user_email, user_git_tsinghua_id, user_last_login_time, user_type)
			VALUES
			(?, ?, ?, ?, ?, ?, ?);
		`,
			2020011446+i,
			"teacher"+strconv.Itoa(i),
			"test teacher"+strconv.Itoa(i),
			strconv.Itoa(i)+"@mails.tsinghua.edu.cn",
			123456,
			"1970-12-31 23:59:59",
			UserTeacher,
		)
		if err != nil {
			log.Fatal(err)
			return
		}
	}
	log.Println("pre test data already attached.")
}

func drop_test_database() {
	_, err := db.Exec("DROP DATABASE IF EXISTS oj_db_test;")
	if err != nil {
		log.Fatal(err)
		return
	}
}

func drop_file_dir() {
	err := os.RemoveAll(FILE_SAVE_ROOT_DIR)
	if err != nil {
		log.Fatal(err)
		return
	}
}
