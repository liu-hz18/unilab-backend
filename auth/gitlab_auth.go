package auth

import (
    "fmt"
	"strconv"
    "net/http"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"unilab-backend/jwt"
	"unilab-backend/apis"
	// "unilab-backend/database"
    "golang.org/x/oauth2"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

var Userinfo map[string]interface{}
var AccessToken string

var endpoint = oauth2.Endpoint{
    AuthURL:  "https://git.tsinghua.edu.cn/oauth/authorize",
    TokenURL: "https://git.tsinghua.edu.cn/oauth/token",
	// AuthURL:  "/gitlab_api/oauth/authorize",
    // TokenURL: "/gitlab_api/oauth/token",
}

var OauthConfig = &oauth2.Config{
    ClientID:     "e635224e0544e1d040525509adac08252870403490ece857ee1a6e7afe998b3d",
    ClientSecret: "3ca8ecf6e80c01ff6d5f31310220bfc27178a3742285255064e3abb33a53a0b4",
    RedirectURL:  "http://localhost:1323/callback",
	// RedirectURL:  "http://localhost:8080",
    Scopes:       []string{"read_user", "read_api", "read_repository"},
    Endpoint:     endpoint,
}

const oauthStateString = "random"

type login struct {
	username string `validate:"required,len=10"`
	userid string `validate:"required"`
}

func GetUserInfo() (int,map[string]interface{}) {
	if len(Userinfo) ==0 {
		return 0,nil
	}else{
		return 1,Userinfo
	}
}

func HandleLogin(c *gin.Context){
	url := OauthConfig.AuthCodeURL(oauthStateString)
	// status,user_info:=GetUserInfo()
	c.JSON(http.StatusOK,gin.H{
		"redirect_url":url,
	})
	// c.Redirect(http.StatusTemporaryRedirect,url)
}

func HandleCallback(c *gin.Context){
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
	// fmt.Println(content)
	if err!=nil {
		fmt.Println("error")
	}
	AccessToken=token.AccessToken

	//Generate token
	user_id:=strconv.Itoa(int(Userinfo["id"].(float64)))//use for password
	user_name:=Userinfo["username"].(string)
	validator := validator.New()
	loginInfo := login{user_name,user_id}
	err = validator.Struct(&loginInfo)
	retcode := apis.INVALID_PARAMS
	query:=url.Values{}
	if err!=nil{
		fmt.Println(err)
	}else{
		// existed := database.CheckUserExist(user_name)
		// if existed {
		rettoken, err := jwt.TokenGenerator(user_name, user_id)
		if err != nil {
			retcode = apis.ERROR_AUTH_TOKEN
		} else {
			retcode = apis.SUCCESS
			query.Add("token",rettoken)
			query.Add("username",user_name)
		}
		// } else {
		// 	retcode = apis.ERROR_AUTH
		// }
	}
	query.Add("code",strconv.Itoa(retcode))
	c.Redirect(http.StatusTemporaryRedirect,"http://localhost:8080/login?"+query.Encode())
}
