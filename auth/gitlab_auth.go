package auth

import (
    "fmt"
    "net/http"
	"io/ioutil"
	"encoding/json"
    "golang.org/x/oauth2"
	"github.com/gin-gonic/gin"
)

var Userinfo map[string]interface{}
var AccessToken string

var endpoint = oauth2.Endpoint{
    AuthURL:  "https://git.tsinghua.edu.cn/oauth/authorize",
    TokenURL: "https://git.tsinghua.edu.cn/oauth/token",
}

var OauthConfig = &oauth2.Config{
    ClientID:     "e635224e0544e1d040525509adac08252870403490ece857ee1a6e7afe998b3d",
    ClientSecret: "3ca8ecf6e80c01ff6d5f31310220bfc27178a3742285255064e3abb33a53a0b4",
    RedirectURL:  "http://localhost:8000/GithubCallback",
    Scopes:       []string{"read_user", "read_api", "read_repository"},
    Endpoint:     endpoint,
}

const oauthStateString = "random"

func GetUserInfo() (int,map[string]interface{}) {
	if len(Userinfo) ==0 {
		return 0,nil
	}else{
		return 1,Userinfo
	}
}

func HandleGithubLogin(c *gin.Context){
	url := OauthConfig.AuthCodeURL(oauthStateString)
	c.Redirect(http.StatusTemporaryRedirect,url)
}

func HandleGithubCallback(c *gin.Context){
	// state := c.Query("state")
	code := c.Query("code")
	token, err := OauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil{
		fmt.Printf("Code exchange failed with '%s'\n", err)
		return
	}
	response, err := http.Get("https://git.tsinghua.edu.cn/api/v4/user?access_token=" + token.AccessToken)
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	content :=string(contents)
	err = json.Unmarshal([]byte(content),&Userinfo)
	if err!=nil {
		fmt.Println("error")
	}
	AccessToken=token.AccessToken
	c.Redirect(http.StatusTemporaryRedirect,"/")
}
