package auth

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/jwt"

	// "unilab-backend/database"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"golang.org/x/oauth2"
)

const FrontEndBaseURL string = "http://localhost:8080"
const BackEndBaseURL string = "http://localhost:1323"
const GitLabAPIBaseURL = "https://git.tsinghua.edu.cn/api/v4"

var endpoint = oauth2.Endpoint{
    AuthURL:  "https://git.tsinghua.edu.cn/oauth/authorize",
    TokenURL: "https://git.tsinghua.edu.cn/oauth/token",
}

var OauthConfig = &oauth2.Config{
    ClientID:     "e635224e0544e1d040525509adac08252870403490ece857ee1a6e7afe998b3d",
    ClientSecret: "3ca8ecf6e80c01ff6d5f31310220bfc27178a3742285255064e3abb33a53a0b4",
    RedirectURL:  BackEndBaseURL + "/callback",
    Scopes:       []string{"read_user", "read_api", "read_repository"},
    Endpoint:     endpoint,
}

const oauthStateString = "random"

type login struct {
	username string `validate:"required,len=10"`
	userid string `validate:"required"`
}

func UserLoginHandler(c *gin.Context){
	url := OauthConfig.AuthCodeURL(oauthStateString)
	c.JSON(http.StatusOK,gin.H{
		"redirect_url":url,
	})
	// c.Redirect(http.StatusTemporaryRedirect,url)
}

type UserIdentity struct {
	Provider string `json:"provider"`
	UID string `json:"extern_uid"`
}

type GitLabResponse struct {
	ID uint32 `json:"id"`
	UserName string `json:"username"`
	RealName string `json:"name"`
	WebURL string `json:"web_url"`
	Email string `json:"email"`
	Identities []UserIdentity `json:"identities"`
}

func GitLabCallBackHandler(c *gin.Context){
	code := c.Query("code")
	token, err := OauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil{
		log.Printf("Code exchange failed with '%s'\n", err)
		return
	}
	response, err := http.Get(GitLabAPIBaseURL + "/user?access_token=" + token.AccessToken)
	if err != nil {
		log.Printf("Get GitLab response failed with '%s'\n", err)
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	content := string(contents)
	// var Userinfo map[string]interface{}
	var userinfo GitLabResponse
	err = json.Unmarshal([]byte(content), &userinfo)
	log.Println("git callback content: ", content)
	if err != nil {
		log.Println("error: ", err)
		return
	}

	// Generate token
	log.Printf("%T, %s\n", userinfo.Identities, userinfo.Identities)
	if len(userinfo.Identities) == 0 {
		log.Println("Error when parsing gitlab callback contents: len(userinfo.Identities) == 0")
		return
	}

	user_id_str := userinfo.Identities[0].UID
	user_id_int64, err := strconv.ParseUint(user_id_str, 10, 32)
	if err != nil {
		log.Println("Error when parsing gitlab callback contents: ", err)
		return
	}
	user_id := uint32(user_id_int64)
	user_name := userinfo.UserName
	validator := validator.New()
	loginInfo := login{user_name, user_id_str}
	err = validator.Struct(&loginInfo)
	if err != nil {
		log.Println(err)
		return
	}
	// TODO: auto confirm user authority
	if database.CheckUserExist(user_id_str) {
		err = database.UpdateUserAccessToken(user_id_str, token.AccessToken)
		if err != nil {
			log.Println("error in update git token: ", err)
			return
		}
	} else {
		newUser := database.CreateUser{
			ID: user_id,
			UserName: userinfo.UserName,
			RealName: userinfo.RealName,
			Email: userinfo.Email,
			GitID: userinfo.ID,
			LastLoginTime: time.Now(),
			Type: database.UserTeacher, // lowest authority, UserStudent
			GitAccessToken: token.AccessToken,
		}
		err = database.CreateNewUser(newUser)
		if err != nil {
			log.Println("error in create new user: ", err)
			return
		}
	}

	retcode := apis.INVALID_PARAMS
	query := url.Values{}
	user_type, err := database.GetUserType(user_id_str)
	if err != nil {
		log.Println(err)
		return
	}
	rettoken, err := jwt.TokenGenerator(user_id_str, user_name)
	if err != nil {
		retcode = apis.ERROR_AUTH_TOKEN
	} else {
		retcode = apis.SUCCESS
		query.Add("token", rettoken)
		query.Add("username", user_name)
		query.Add("permission", strconv.FormatUint(uint64(user_type), 10))
		query.Add("userid", user_id_str)
	}
	query.Add("code", strconv.Itoa(retcode))
	c.Redirect(http.StatusTemporaryRedirect, FrontEndBaseURL + "/login?" + query.Encode())
}
