package setting

import (
	"fmt"
	"log"
	"time"

	"github.com/go-ini/ini"
)

var (
	Cfg *ini.File
	// app
	RunMode string
	JwtSecret string
	UploadFileRootDir string
	CourseRootDir string
	UserRootDir string
	QuestionRootDir string
	RuntimeRootDir string

	FrontEndBaseUrl string
	BackEndBaseURL string

	// auth
	GitLabBaseURL string
	GitLabAuthURL string
	GitLabTokenURL string
	ClientID string
	ClientSecret string

	// server
	HttpPort int
	ReadTimeout time.Duration
	WriteTimeout time.Duration

	// database
	DBType string
	DBName string
	DBUser string
	DBPassword string
	DBHost string
	ClearOnStart bool
)


func init() {
	var err error
	Cfg, err = ini.Load("conf.ini")
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to parse 'conf.ini': ", err))
	}
	
	sec, err := Cfg.GetSection("app")
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to get section 'app' in 'conf.ini': ", err))
	}
	JwtSecret = sec.Key("JWT_SECRET").MustString("!@)*#)!@U#@*!@!)")
	RunMode = sec.Key("RUN_MODE").MustString("debug") // default is debug
	UploadFileRootDir = sec.Key("UPLOAD_FILE_ROOT_DIR").MustString("../unilab-files/")
	CourseRootDir = UploadFileRootDir + sec.Key("COURSE_SUB_DIR").MustString("course/")
	UserRootDir = UploadFileRootDir + sec.Key("USER_SUB_DIR").MustString("user/")
	QuestionRootDir = UploadFileRootDir + sec.Key("QUESTION_SUB_DIR").MustString("question/")
	RuntimeRootDir = sec.Key("RUNTIME_ROOT_DIR").MustString("runtime/")
	FrontEndBaseUrl = sec.Key("FRONTEND_BASE_URL").MustString("https://lab.cs.tsinghua.edu.cn")
	BackEndBaseURL = sec.Key("BACKEND_BASE_URL").MustString("https://lab.cs.tsinghua.edu.cn/api")

	
	sec, err = Cfg.GetSection("auth")
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to get section 'auth' in 'conf.ini': ", err))
	}
	GitLabBaseURL = sec.Key("GITLAB_BASE_URL").MustString("https://git.tsinghua.edu.cn/api/v4")
	GitLabAuthURL = sec.Key("AUTH_URL").MustString("https://git.tsinghua.edu.cn/oauth/authorize")
	GitLabTokenURL = sec.Key("TOKEN_URL").MustString("https://git.tsinghua.edu.cn/oauth/token")
	ClientID = sec.Key("ID").String()
	ClientSecret = sec.Key("SECRET").String()

	sec, err = Cfg.GetSection("server")
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to get section 'server' in 'conf.ini': ", err))
	}
	HttpPort = sec.Key("HTTP_PORT").MustInt(1323)
	ReadTimeout = time.Duration(sec.Key("READ_TIMEOUT").MustInt(60)) * time.Second
	WriteTimeout = time.Duration(sec.Key("WRITE_TIMEOUT").MustInt(60)) * time.Second
	

	sec, err = Cfg.GetSection("database")
	if err != nil {
		log.Fatal(fmt.Sprintf("Fail to get section 'database' in 'conf.ini': ", err))
	}
	DBType = sec.Key("TYPE").String()
	DBName = sec.Key("NAME").String()
	DBPassword = sec.Key("PASSWORD").String()
	DBUser = sec.Key("USER").String()
	DBHost = sec.Key("HOST").String()
	ClearOnStart = sec.Key("CLEAR_ON_RESTART").MustBool(true)
}
