package auth

import (
	"unilab-backend/logging"
	"unilab-backend/utils"
)

func isAdmin(userid string) bool {
	userIDInt, err := utils.StringToInt(userid)
	if err != nil {
		logging.Info(err)
		return false
	}
	// TODO: 请在此处添加 admin 学号
	var adminIDs = []int{2018011446, 2018011302}
	for _, id := range adminIDs {
		if id == userIDInt {
			return true
		}
	}
	return false
}
