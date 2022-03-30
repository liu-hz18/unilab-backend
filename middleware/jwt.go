package middleware

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/jwt"
)

// JWT 中间件
func JWTMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		data := make(map[string]interface{})
		var claims *jwt.Claims
		var err error
		code = apis.SUCCESS
		// 读取header中的token
		token := c.Request.Header.Get("Authorization")
		log.Println(token)
		if token == "" {
			code = apis.INVALID_PARAMS
		} else {
			token = strings.Fields(token)[1]
			claims, err = jwt.ParseToken(token)
			log.Println("claims:", claims)
			if err != nil {
				code = apis.ERROR_AUTH_CHECK_TOKEN_FAIL
				data["err"] = err.Error()
			} else if time.Now().Unix() > claims.ExpiresAt {
				code = apis.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
				data["err"] = "auth check token timeout."
			}
		}
		if code != apis.SUCCESS {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg": apis.MsgFlags[code],
				"data": data,
			})
			c.Abort()
			return
		}
		// read database to get authorization role
		user_type, err := database.GetUserType(claims.Userid)
		if err != nil {
			code = apis.ERROR
			data["err"] = err.Error()
			log.Println("JWTMiddleWare", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg": apis.MsgFlags[code],
				"data": data,
			})
			c.Abort()
			return
		}
		c.Set("user_type", user_type)
		user_id, err := strconv.ParseUint(claims.Userid, 10, 32)
		c.Set("user_id", uint32(user_id))
		c.Next()
	}
}

