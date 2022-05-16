package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"
	"unilab-backend/logging"
	"unilab-backend/setting"
)

var db *sql.DB

func InitDB() {
	var err error
	db, err = sql.Open(setting.DBType, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local",
		setting.DBUser,
		setting.DBPassword,
		setting.DBHost,
		setting.DBName,
	))
	if err != nil {
		logging.Fatal(err)
		return
	}
	// setup a MySQL connection pool
	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxIdleConns(16)
	db.SetMaxOpenConns(128)
	// verify connection
	if err = db.Ping(); err != nil {
		logging.Fatal(err)
		return
	}
	logging.Info("db connection established.")
	// 静态低频查询信息
	// - 多对多: user table, course table, 建立关联表，总共3个表
	// - 多对多: question table, homework table, 建立关联表，总共3个
	// - 多对多: course table, question table
	// - 一对多: course table, homework table
	// 一对多: file table -> question table, user table
	// 动态高频查询信息, 和评测相关
	// test-run table: user, file, question, course
	// user log
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
		logging.Fatal(err)
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
			logging.Fatal(err)
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
			logging.Fatal(err)
			return
		}
	}
	logging.Info("pre test data already attached.")
}
