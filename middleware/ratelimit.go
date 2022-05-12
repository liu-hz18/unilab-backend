package middleware

import (
	"unilab-backend/setting"

	"github.com/didip/tollbooth"
	"github.com/gin-gonic/gin"
)

func RateLimitMiddleware() gin.HandlerFunc {
	lmt := tollbooth.NewLimiter(float64(setting.RateLimit), nil) // 1 request/second
	lmt.SetIPLookups([]string{"X-Forwarded-For", "RemoteAddr", "X-Real-IP"}).
		SetMethods([]string{"GET", "POST"}).
		SetMessage("You have reached maximum request limit, please wait for a while...  :)").
		SetMessageContentType("text/plain; charset=utf-8")

	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)
		if httpError != nil {
			c.Data(httpError.StatusCode, lmt.GetMessageContentType(), []byte(httpError.Message))
			c.Abort()
		} else {
			c.Next()
		}
	}
}
