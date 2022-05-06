package database

import (
	"fmt"
	"time"
	"unilab-backend/logging"
)

// user type
const (
	UserAdmin   uint8 = 2
	UserTeacher uint8 = 1
	UserStudent uint8 = 0
)

type UserInfo struct {
	ID   uint32 `json:"id" form:"id" uri:"id" binding:"required"`
	Name string `json:"name" form:"name" uri:"name" binding:"required"`
}

type CreateUser struct {
	ID             uint32
	UserName       string
	RealName       string
	Email          string
	GitID          uint32
	LastLoginTime  time.Time
	Type           uint8
	GitAccessToken string
}

func CreateNewUser(user CreateUser) error {
	_, err := db.Exec(`INSERT INTO oj_user
		(user_id, user_name, user_real_name, user_email, user_git_tsinghua_id, user_last_login_time, user_signup_time, user_type, user_token)
		VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		user.ID,
		user.UserName,
		user.RealName,
		user.Email,
		user.GitID,
		time.Now(),
		time.Now(),
		user.Type,
		user.GitAccessToken,
	)
	if err != nil {
		logging.Info(err)
		return err
	}
	return nil
}

func UpdateUserAccessToken(userid string, accessToken string) error {
	_, err := db.Exec(`UPDATE oj_user SET
		user_last_login_time = ?,
		user_token = ?
		WHERE
		user_id = ?
	`,
		time.Now(),
		accessToken,
		userid,
	)
	if err != nil {
		logging.Info(err)
		return err
	}
	return nil
}

func UpdateUserInfo(userid string, info CreateUser) error {
	_, err := db.Exec(`UPDATE oj_user SET
		user_name = ?,
		user_real_name = ?,
		user_email = ?,
		user_git_tsinghua_id = ?,
		user_last_login_time = ?,
		user_token = ?
		WHERE
		user_id = ?
	`,
		info.UserName,
		info.RealName,
		info.Email,
		info.GitID,
		time.Now(),
		info.GitAccessToken,
		userid,
	)
	if err != nil {
		logging.Info(err)
		return err
	}
	return nil
}

func GetUserAccessToken(userid string) (string, error) {
	var access_token string
	err := db.QueryRow("SELECT user_token from oj_user where user_id=?;", userid).Scan(&access_token)
	// logging.Info("GetUserAccessToken()", userid, access_token)
	return access_token, err
}

func CheckUserExist(userid string) bool {
	var userid_db string
	err := db.QueryRow("SELECT user_id from oj_user where user_id=?;", userid).Scan(&userid_db)
	if err != nil {
		logging.Info("get user failed, err: ", err)
		return false
	}
	return true
}

func GetUserType(userid string) (uint8, error) {
	var user_type_db uint8
	err := db.QueryRow("SELECT user_type from oj_user where user_id=?;", userid).Scan(&user_type_db)
	// logging.Info("GetUserType()", userid, user_type_db)
	return user_type_db, err
}

func GetUserName(userid string) (string, error) {
	var username string
	err := db.QueryRow("SELECT user_name from oj_user where user_id=?;", userid).Scan(&username)
	return username, err
}

func GetUserNameType(userid string) (string, uint8, error) {
	var username string
	var user_type_db uint8
	err := db.QueryRow("SELECT user_name, user_type from oj_user where user_id=?;", userid).Scan(&username, &user_type_db)
	return username, user_type_db, err
}

func GetAllUsersNameAndID() ([]UserInfo, error) {
	userinfos := []UserInfo{}
	res, err := db.Query("SELECT user_id, user_real_name FROM oj_user;")
	if err != nil {
		logging.Info(err)
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var info UserInfo
		err := res.Scan(&info.ID, &info.Name)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		userinfos = append(userinfos, info)
	}
	return userinfos, nil
}

func GetAllTeachersNameAndID() ([]UserInfo, error) {
	userinfos := []UserInfo{}
	res, err := db.Query("SELECT user_id, user_real_name FROM oj_user WHERE user_type>=?;", UserTeacher)
	if err != nil {
		logging.Info(err)
		return nil, err
	}
	defer res.Close()
	for res.Next() {
		var info UserInfo
		err := res.Scan(&info.ID, &info.Name)
		if err != nil {
			logging.Info(err)
			return nil, err
		}
		userinfos = append(userinfos, info)
	}
	return userinfos, nil
}

func CreateUsersIfNotExists(userInfos []UserInfo, authority uint8, update_auth bool) error {
	// setup a transaction
	tx, err := db.Begin()
	if err != nil {
		if tx != nil {
			_ = tx.Rollback()
		}
		logging.Info("CreateUsersIfNotExists() begin trans action failed, err:", err)
		return err
	}
	// insert new teacher if not existed
	var insertSqlStr string
	if update_auth {
		insertSqlStr = "INSERT INTO oj_user (user_id, user_real_name, user_type) VALUES "
	} else {
		insertSqlStr = "INSERT IGNORE INTO oj_user (user_id, user_real_name, user_type) VALUES "
	}
	for index, user_info := range userInfos {
		if index < len(userInfos)-1 {
			insertSqlStr += fmt.Sprintf("(%d, '%s', %d),", user_info.ID, user_info.Name, authority)
		} else {
			insertSqlStr += fmt.Sprintf("(%d, '%s', %d)", user_info.ID, user_info.Name, authority)
		}
	}
	if update_auth {
		insertSqlStr += "ON DUPLICATE KEY UPDATE user_type=values(user_type);"
	} else {
		insertSqlStr += ";"
	}
	_, err = tx.Exec(insertSqlStr)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	_ = tx.Commit()
	logging.Info("CreateUsersIfNotExists() commit trans action successfully.")
	return nil
}
