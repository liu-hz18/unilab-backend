package auth

import (
    // "fmt"
    // "io/ioutil"
    "net/http"
	// "unilab-backend/auth"
    // "golang.org/x/oauth2"
	"github.com/gin-gonic/gin"
)

const htmlIndex = `<html><body>
<a href="/GithubLogin">Log in with Github</a>
</body></html>
`

func main(){
	r := gin.Default()
	r.GET("/",handleMain)
	r.GET("/GithubLogin",HandleGithubLogin)
	r.GET("/GithubCallback",HandleGithubCallback)
    r.Run(":8080")
}

func handleMain(c *gin.Context){
	c.Header("Content-Type","text/html; charset=utf-8")
	c.String(http.StatusOK,htmlIndex)
}

// func handleMain(w http.ResponseWriter, r *http.Request) {
//     fmt.Fprintf(w, htmlIndex)
// }