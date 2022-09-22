package auth

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/jwt"
	"unilab-backend/logging"
	"unilab-backend/setting"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/oauth2"
)

var OauthConfig = &oauth2.Config{
	ClientID:     setting.ClientID,
	ClientSecret: setting.ClientSecret,
	RedirectURL:  setting.BackEndBaseURL + "/callback",
	Scopes:       []string{"read_user", "read_api", "read_repository"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  setting.GitLabAuthURL,
		TokenURL: setting.GitLabTokenURL,
	},
}

const oauthStateString = "random"

type login struct {
	username string `validate:"required,len=10"`
	userid   string `validate:"required"`
}

func UserLoginHandler(c *gin.Context) {
	url := OauthConfig.AuthCodeURL(oauthStateString)
	c.JSON(http.StatusOK, gin.H{
		"redirect_url": url,
	})
	// c.Redirect(http.StatusTemporaryRedirect,url)
}

type UserIdentity struct {
	Provider string `json:"provider"`
	UID      string `json:"extern_uid"`
}

type GitLabResponse struct {
	ID         uint32         `json:"id"`
	UserName   string         `json:"username"`
	RealName   string         `json:"name"`
	WebURL     string         `json:"web_url"`
	Email      string         `json:"email"`
	Identities []UserIdentity `json:"identities"`
}

func GitLabCallBackHandler(c *gin.Context) {
	code := c.Query("code")
	token, err := OauthConfig.Exchange(context.Background(), code)
	if err != nil {
		logging.Error("Code exchange failed with err: ", err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	response, err := http.Get(setting.GitLabBaseURL + "/user?access_token=" + token.AccessToken)
	logging.Info("token: %s\n", token.AccessToken)
	if err != nil {
		logging.Error("Get GitLab response failed with err: ", err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	defer response.Body.Close()
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		logging.Error(err.Error())
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	content := string(contents)
	// var Userinfo map[string]interface{}
	var userinfo GitLabResponse
	err = json.Unmarshal([]byte(content), &userinfo)
	logging.Info("git callback content: ", content)
	if err != nil {
		logging.Error("error: ", err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}

	// Generate token
	logging.Info(userinfo.Identities)
	if len(userinfo.Identities) == 0 {
		logging.Error("Error when parsing gitlab callback contents: len(userinfo.Identities) == 0")
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}

	userIDStr := userinfo.Identities[0].UID
	userIDInt64, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		logging.Error("Error when parsing gitlab callback contents: ", err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	userID := uint32(userIDInt64)
	userName := userinfo.UserName
	validator := validator.New()
	loginInfo := login{userName, userIDStr}
	err = validator.Struct(&loginInfo)
	if err != nil {
		logging.Error(err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	// confirm user authority
	var userTypeTmp uint8 = database.UserStudent
	if isAdmin(userIDStr) {
		userTypeTmp = database.UserAdmin
	}
	newUser := database.CreateUser{
		ID:             userID,
		UserName:       userinfo.UserName,
		RealName:       userinfo.RealName,
		Email:          userinfo.Email,
		GitID:          userinfo.ID,
		LastLoginTime:  time.Now(),
		Type:           userTypeTmp, // lowest authority, UserStudent
		GitAccessToken: token.AccessToken,
	}
	if database.CheckUserExist(userIDStr) {
		err = database.UpdateUserInfo(userIDStr, newUser)
		// err = database.UpdateUserAccessToken(userIDStr, token.AccessToken)
		if err != nil {
			logging.Error("error in update user info: ", err)
			apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
			return
		}
	} else {
		err = database.CreateNewUser(newUser)
		if err != nil {
			logging.Error("error in create new user: ", err)
			apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
			return
		}
	}

	query := url.Values{}
	userType, err := database.GetUserType(userIDStr)
	if err != nil {
		logging.Error(err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	rettoken, err := jwt.TokenGenerator(userIDStr, userName)
	if err != nil {
		query.Add("code", strconv.Itoa(apis.ERROR_AUTH_TOKEN))
	} else {
		query.Add("code", strconv.Itoa(apis.SUCCESS))
		query.Add("token", rettoken)
		query.Add("username", userName)
		query.Add("permission", strconv.FormatUint(uint64(userType), 10))
		query.Add("userid", userIDStr)
	}
	c.Redirect(http.StatusTemporaryRedirect, setting.FrontEndBaseURL+"/login?"+query.Encode())
}
