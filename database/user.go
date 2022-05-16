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
	var accessToken string
	err := db.QueryRow("SELECT user_token from oj_user where user_id=?;", userid).Scan(&accessToken)
	// logging.Info("GetUserAccessToken()", userid, accessToken)
	return accessToken, err
}

func GetUserGitTsinghuaID(userid string) (string, error) {
	var userGitTsinghuaID string
	err := db.QueryRow("SELECT user_git_tsinghua_id from oj_user where user_id=?;", userid).Scan(&userGitTsinghuaID)
	return userGitTsinghuaID, err
}

func CheckUserExist(userid string) bool {
	var useridDB string
	err := db.QueryRow("SELECT user_id from oj_user where user_id=?;", userid).Scan(&useridDB)
	if err != nil {
		logging.Info("get user failed, err: ", err)
		return false
	}
	return true
}

func GetUserType(userid string) (uint8, error) {
	var userTypeDb uint8
	err := db.QueryRow("SELECT user_type from oj_user where user_id=?;", userid).Scan(&userTypeDb)
	return userTypeDb, err
}

func GetUserName(userid string) (string, error) {
	var userName string
	err := db.QueryRow("SELECT user_name from oj_user where user_id=?;", userid).Scan(&userName)
	return userName, err
}

func GetUserNameType(userid string) (string, uint8, error) {
	var userName string
	var userTypeDb uint8
	err := db.QueryRow("SELECT user_name, user_type from oj_user where user_id=?;", userid).Scan(&userName, &userTypeDb)
	return userName, userTypeDb, err
}

func GetAllUsersNameAndID() ([]UserInfo, error) {
	userInfos := []UserInfo{}
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
		userInfos = append(userInfos, info)
	}
	return userInfos, nil
}

func GetAllTeachersNameAndID() ([]UserInfo, error) {
	userInfos := []UserInfo{}
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
		userInfos = append(userInfos, info)
	}
	return userInfos, nil
}

func CreateUsersIfNotExists(userInfos []UserInfo, authority uint8, updateAuth bool) error {
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
	var insertSQLStr string
	if updateAuth {
		insertSQLStr = "INSERT INTO oj_user (user_id, user_real_name, user_type) VALUES "
	} else {
		insertSQLStr = "INSERT IGNORE INTO oj_user (user_id, user_real_name, user_type) VALUES "
	}
	for index, userInfo := range userInfos {
		if index < len(userInfos)-1 {
			insertSQLStr += fmt.Sprintf("(%d, '%s', %d),", userInfo.ID, userInfo.Name, authority)
		} else {
			insertSQLStr += fmt.Sprintf("(%d, '%s', %d)", userInfo.ID, userInfo.Name, authority)
		}
	}
	if updateAuth {
		insertSQLStr += "ON DUPLICATE KEY UPDATE user_type=values(user_type);"
	} else {
		insertSQLStr += ";"
	}
	_, err = tx.Exec(insertSQLStr)
	if err != nil {
		_ = tx.Rollback()
		logging.Info(err)
		return err
	}
	_ = tx.Commit()
	logging.Info("CreateUsersIfNotExists() commit trans action successfully.")
	return nil
}
