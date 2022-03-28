package middleware

import (
	"log"
	"net/http"
	"unilab-backend/apis"

	"github.com/gin-gonic/gin"
)

func PriorityMiddleware(user_type_permission uint8) gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}
		code = apis.SUCCESS
		user_type := c.MustGet("user_type").(uint8)
		log.Printf("UserPriority: %d", user_type)
		if user_type < user_type_permission {
			code = apis.SUCCESS
			c.JSON(http.StatusUnauthorized, gin.H{
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

