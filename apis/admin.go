package apis

import (
	"log"
	"net/http"
	"unilab-backend/database"

	"github.com/gin-gonic/gin"
)


type CreateTeachers struct {
	Teachers []uint32 `json:"teachers" form:"teachers" uri:"teachers" binding:"required"`
}

// admin add teacher to database
func AddTeachersHandler(c *gin.Context) {
	var form CreateTeachers
	if err := c.ShouldBind(&form); err != nil {
		ErrorResponse(c, INVALID_PARAMS, err.Error())
		return
	} else {
		log.Println(form)
		if len(form.Teachers) == 0 {
			ErrorResponse(c, INVALID_PARAMS, "Added teachers are none.")
			return
		} else {
			// insert into database
			err := database.CreateUsersIfNotExists(form.Teachers, database.UserTeacher)
			if err != nil {
				ErrorResponse(c, INVALID_PARAMS, err.Error())
				return
			}
		}
	}
	data := make(map[string]interface{})
	data["result"] = "successfully add teachers!"
	c.JSON(http.StatusOK, gin.H{
		"code": SUCCESS,
		"msg":  MsgFlags[SUCCESS],
		"data": data,
	})
}
