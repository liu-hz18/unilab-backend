package gitlab_api

import (
    "fmt"
    // "io/ioutil"
    "net/http"
	"unilab-backend/auth"
    // "unilab-backend/gitlab_api"
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
	r.GET("/GithubLogin",auth.HandleGithubLogin)
	r.GET("/GithubCallback",auth.HandleGithubCallback)
    r.Run(":8000")
}

func handleMain(c *gin.Context){
    status,_ :=auth.GetUserInfo()
    if status==0{
        fmt.Println("Please login.")
    }else{
        Get_project_traces("minidecaf-2018011302")
        // fmt.Println(userinfo)
    }
	c.Header("Content-Type","text/html; charset=utf-8")
	c.String(http.StatusOK,htmlIndex)
}