package apis

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"

	"unilab-backend/database"
	"unilab-backend/jwt"
	"unilab-backend/logging"
)

// login router
type login struct {
	userid string `validate:"required,len=10"`
	password string `validate:"required"`
}

var permissionMap = map[uint8]string{
	database.UserAdmin: "admin",
	database.UserTeacher: "teacher",
	database.UserStudent: "student",
}

// user login
func UserLoginHandler(c *gin.Context) {
	userid := c.PostForm("userid")
	password := c.PostForm("password")
	logging.Info("user id: ", userid,  ", password: ", password)
	validator := validator.New()
	loginInfo := login{userid: userid, password:password}
	err := validator.Struct(&loginInfo)
	code := INVALID_PARAMS
	data := make(map[string]interface{})
	if err != nil { // code = INVALID_PARAMS
		logging.Info(err)
		data["err"] = err.Error()
	} else {
		// 加上gitlab验证 ?
		existed := database.CheckUserExist(userid)
		username, usertype, err := database.GetUserNameType(userid)
		if existed && err == nil {
			token, err := jwt.TokenGenerator(userid, password)
			if err != nil {
				logging.Info(err)
				code = ERROR_AUTH_TOKEN
				data["err"] = err.Error()
			} else {
				data["token"] = token
				data["username"] = username
				data["permission"] = permissionMap[usertype]
				code = SUCCESS
			}
		} else {
			logging.Info(err)
			data["err"] = err.Error()
			code = ERROR_AUTH
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg": MsgFlags[code],
		"data": data,
	})
}
