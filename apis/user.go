package apis

import (
	"log"
	"net/http"
	"unilab-backend/database"

	"github.com/gin-gonic/gin"
)


func GetAllUsersHandler(c *gin.Context) {
	userinfos, err := database.GetAllUsersNameAndID()
	data := make(map[string]interface{})
	code := SUCCESS
	if err != nil {
		code = ERROR
		data["err"] = err.Error()
	}
	log.Printf("fetch %d rows from oj_user in GetAllUsersHandler()", len(userinfos))
	data["result"] = userinfos
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg": MsgFlags[code],
		"data": data,
	})
}
