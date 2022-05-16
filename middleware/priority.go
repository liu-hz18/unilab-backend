package middleware

import (
	"net/http"
	"unilab-backend/apis"

	"github.com/gin-gonic/gin"
)

func PriorityMiddleware(userTypePermission uint8) gin.HandlerFunc {
	return func(c *gin.Context) {
		data := make(map[string]interface{})
		userType := c.MustGet("user_type").(uint8)
		if userType < userTypePermission {
			data["err"] = "permission denyed!"
			c.JSON(http.StatusForbidden, gin.H{
				"code": apis.ERROR_AUTH_USER_PERMISSIONS,
				"msg":  apis.MsgFlags[apis.ERROR_AUTH_USER_PERMISSIONS],
				"data": data,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
