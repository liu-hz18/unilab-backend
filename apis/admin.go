package apis

import (
	"net/http"
	"unilab-backend/database"
	"unilab-backend/logging"

	"github.com/gin-gonic/gin"
)

// admin add teacher to database
func AddTeachersHandler(c *gin.Context) {
	var form []database.UserInfo
	if err := c.ShouldBind(&form); err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	// else: check form
	logging.Info(form)
	if len(form) == 0 {
		ErrorResponse(c, INVALID_PARAMS, "Added teachers are none.")
		return
	}
	// else: insert into database
	err := database.CreateUsersIfNotExists(form, database.UserTeacher, true)
	if err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	}
	data := make(map[string]interface{})
	data["result"] = "successfully add teachers!"
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}
