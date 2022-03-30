package database

import (
	"fmt"
	"log"
)

// user type
const (
	UserAdmin   uint8 = 2
	UserTeacher uint8 = 1
	UserStudent uint8 = 0
)

type UserInfo struct {
	ID uint32 `json:"id"`
	Name string `json:"name"`
}

func CheckUserExist(userid string) bool {
	var userid_db string
	err := db.QueryRow("SELECT user_id from oj_user where user_id=?;", userid).Scan(&userid_db)
	if err != nil {
		log.Printf("get user failed, err: %v\n", err)
		return false
	}
	return true
}

func GetUserType(userid string) (uint8, error) {
	_, err := db.Exec("USE oj_db_test;")
	var user_type_db uint8
	err = db.QueryRow("SELECT user_type from oj_db_test.oj_user where user_id=?;", userid).Scan(&user_type_db)
	log.Println("GetUserType()", userid, user_type_db)
	return user_type_db, err
}

func GetUserName(userid string) (string, error) {
	var username string
	err := db.QueryRow("SELECT user_name from oj_db_test.oj_user where user_id=?;", userid).Scan(&username)
	return username, err
}

func GetUserNameType(userid string) (string, uint8, error) {
	var username string
	var user_type_db uint8
	err := db.QueryRow("SELECT user_name, user_type from oj_db_test.oj_user where user_id=?;", userid).Scan(&username, &user_type_db)
	return username, user_type_db, err
}

func GetAllUsersNameAndID() ([]UserInfo, error) {
	userinfos := []UserInfo{}
	res, err := db.Query("SELECT user_id, user_name FROM oj_db_test.oj_user;")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var info UserInfo
		err := res.Scan(&info.ID, &info.Name)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		userinfos = append(userinfos, info)
	}
	return userinfos, nil
}

func checkTableExists(tableName string) bool {
	mySqlStr := fmt.Sprintf("SELECT * FROM oj_db_test.%s;", tableName)
	res, err := db.Query(mySqlStr)
	if err != nil {
		log.Println("table ", tableName, "not existed. ERROR:", err.Error())
		return false;
	}
	defer res.Close()
	return true;
}
