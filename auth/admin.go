package auth

import (
	"log"
	"unilab-backend/utils"
)

func isAdmin(userid string) bool {
	userid_int, err := utils.StringToInt(userid)
	if err != nil {
		log.Println(err)
		return false
	}
	// TODO: 请在此处添加 admin 学号
	var adminIDs = []int{2018011446, 2018011302}
	for _, id := range adminIDs {
		if id == userid_int {
			return true
		}
	}
	return false
}
