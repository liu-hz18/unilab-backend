package middleware

import (
	"net/http"
	"unilab-backend/apis"
	"unilab-backend/logging"

	"github.com/gin-gonic/gin"
)

func PriorityMiddleware(user_type_permission uint8) gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}
		code = apis.SUCCESS
		user_type := c.MustGet("user_type").(uint8)
		logging.Info("UserPriority: ", user_type)
		if user_type < user_type_permission {
			code = apis.ERROR_AUTH_USER_PERMISSIONS
			c.JSON(http.StatusForbidden, gin.H{
				"code": code,
				"msg":  apis.MsgFlags[code],
				"data": data,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
