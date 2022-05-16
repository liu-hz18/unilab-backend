package middleware

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/jwt"
	"unilab-backend/logging"
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
		// logging.Info(token)
		if token == "" {
			code = apis.INVALID_PARAMS
		} else {
			token = strings.Fields(token)[1]
			claims, err = jwt.ParseToken(token)
			switch {
			case err != nil:
				code = apis.ERROR_AUTH_CHECK_TOKEN_FAIL
				data["err"] = err.Error()
				logging.Info(err)
			case time.Now().Unix() > claims.ExpiresAt:
				code = apis.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
				data["err"] = "auth check token timeout."
			default:
				logging.Info("access user id: ", claims.Userid, " | user name: ", claims.UserName)
			}
		}
		if code != apis.SUCCESS {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg":  apis.MsgFlags[code],
				"data": data,
			})
			c.Abort()
			return
		}
		// read database to get authorization role
		userType, err := database.GetUserType(claims.Userid)
		userID, _ := strconv.ParseUint(claims.Userid, 10, 32)
		if err != nil {
			code = apis.ERROR
			data["err"] = err.Error()
			logging.Info("JWTMiddleWare", err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg":  apis.MsgFlags[code],
				"data": data,
			})
			c.Abort()
			return
		}
		c.Set("user_type", userType)
		c.Set("user_id", uint32(userID))
		c.Next()
	}
}
