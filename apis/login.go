package apis

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"

	"unilab-backend/database"
	"unilab-backend/jwt"
)

// login router
type login struct {
	userid string `validate:"required,len=10"`
	password string `validate:"required"`
}

// user login
func UserLoginHandler(c *gin.Context) {
	userid := c.PostForm("userid")
	password := c.PostForm("password")
	log.Printf("user id: %s, password: %s", userid, password)
	validator := validator.New()
	loginInfo := login{userid: userid, password:password}
	err := validator.Struct(&loginInfo)
	code := INVALID_PARAMS
	data := make(map[string]interface{})
	if err != nil { // code = INVALID_PARAMS
		log.Println(err)
	} else {
		// 加上gitlab验证 ?
		existed := database.CheckUserExist(userid)
		username, err := database.GetUserName(userid)
		if existed && err == nil {
			token, err := jwt.TokenGenerator(userid, password)
			if err != nil {
				code = ERROR_AUTH_TOKEN
			} else {
				data["token"] = token
				data["username"] = username
				code = SUCCESS
			}
		} else {
			code = ERROR_AUTH
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg": MsgFlags[code],
		"data": data,
	})
}
