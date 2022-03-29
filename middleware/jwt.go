package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"unilab-backend/apis"
	"unilab-backend/jwt"
)

// JWT 中间件
func JWTMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}
		code = apis.SUCCESS
		token := c.Query("token")
		if token == "" {
			code = apis.INVALID_PARAMS
		} else {
			claims, err := jwt.ParseToken(token)
			if err != nil {
				code = apis.ERROR_AUTH_CHECK_TOKEN_FAIL
			} else if time.Now().Unix() > claims.ExpiresAt {
				code = apis.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
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
		c.Next()
	}
}

