package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/jwt"
	"unilab-backend/logging"
	"unilab-backend/setting"

	// "unilab-backend/database"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"golang.org/x/oauth2"
)

const GitLabAPIBaseURL = "https://git.tsinghua.edu.cn/api/v4"

var endpoint = oauth2.Endpoint{
    AuthURL:  "https://git.tsinghua.edu.cn/oauth/authorize",
    TokenURL: "https://git.tsinghua.edu.cn/oauth/token",
}

var OauthConfig = &oauth2.Config{
    ClientID:     "e635224e0544e1d040525509adac08252870403490ece857ee1a6e7afe998b3d",
    ClientSecret: "3ca8ecf6e80c01ff6d5f31310220bfc27178a3742285255064e3abb33a53a0b4",
    RedirectURL:  setting.BackEndBaseURL + "/callback",
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
		logging.Info("Code exchange failed with err: ", err)
		return
	}
	response, err := http.Get(GitLabAPIBaseURL + "/user?access_token=" + token.AccessToken)
	logging.Info("token: %s\n", token.AccessToken)
	if err != nil {
		logging.Info("Get GitLab response failed with err: ", err)
		return
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	content := string(contents)
	// var Userinfo map[string]interface{}
	var userinfo GitLabResponse
	err = json.Unmarshal([]byte(content), &userinfo)
	logging.Info("git callback content: ", content)
	if err != nil {
		logging.Info("error: ", err)
		return
	}

	// Generate token
	logging.Info(userinfo.Identities)
	if len(userinfo.Identities) == 0 {
		logging.Info("Error when parsing gitlab callback contents: len(userinfo.Identities) == 0")
		return
	}

	user_id_str := userinfo.Identities[0].UID
	user_id_int64, err := strconv.ParseUint(user_id_str, 10, 32)
	if err != nil {
		logging.Info("Error when parsing gitlab callback contents: ", err)
		return
	}
	user_id := uint32(user_id_int64)
	user_name := userinfo.UserName
	validator := validator.New()
	loginInfo := login{user_name, user_id_str}
	err = validator.Struct(&loginInfo)
	if err != nil {
		logging.Info(err)
		return
	}
	// confirm user authority
	var user_type_tmp uint8 = database.UserStudent
	if isAdmin(user_id_str) {
		user_type_tmp = database.UserAdmin
	}
	newUser := database.CreateUser{
		ID: user_id,
		UserName: userinfo.UserName,
		RealName: userinfo.RealName,
		Email: userinfo.Email,
		GitID: userinfo.ID,
		LastLoginTime: time.Now(),
		Type:  user_type_tmp, // lowest authority, UserStudent
		GitAccessToken: token.AccessToken,
	}
	if database.CheckUserExist(user_id_str) {
		err = database.UpdateUserInfo(user_id_str, newUser)
		// err = database.UpdateUserAccessToken(user_id_str, token.AccessToken)
		if err != nil {
			logging.Info("error in update user info: ", err)
			return
		}
	} else {
		err = database.CreateNewUser(newUser)
		if err != nil {
			logging.Info("error in create new user: ", err)
			return
		}
	}

	retcode := apis.INVALID_PARAMS
	query := url.Values{}
	user_type, err := database.GetUserType(user_id_str)
	if err != nil {
		logging.Info(err)
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
	c.Redirect(http.StatusTemporaryRedirect, setting.FrontEndBaseUrl + "/login?" + query.Encode())
}
