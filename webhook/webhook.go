package webhook

import (
	"net/http"
	"strconv"
	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/gitlab_api"
	"unilab-backend/logging"
	"unilab-backend/os"

	"github.com/gin-gonic/gin"
)

type WebhookInfo struct {
	Attributes struct {
		Branch        string `json:"ref"`
		Status        string `json:"status"`
		Detail_status string `json:"detailed_status"`
	} `json:"object_attributes"`
	UserInfo struct {
		UserID uint32 `json:"id"`
	} `json:"user"`
	Project_info struct {
		Project_id uint32 `json:"id"`
	} `json:"project"`
	JobInfos []JobInfo `json:"builds"`
}

type JobInfo struct {
	Job_id uint32 `json:"id"`
}

func OsWebhookHandler(c *gin.Context) {
	var webhookInfo WebhookInfo
	if err := c.ShouldBindJSON(&webhookInfo); err != nil {
		logging.Info(err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	project_id := strconv.Itoa(int(webhookInfo.Project_info.Project_id))
	job_id := strconv.Itoa(int(webhookInfo.JobInfos[0].Job_id))
	accessToken, err := database.GetUserAccessToken("2018011302")
	if err != nil {
		apis.ErrorResponse(c, apis.ERROR, err.Error())
		return
	}
	if webhookInfo.Attributes.Detail_status=="passed" || webhookInfo.Attributes.Detail_status=="failed"{
		// fmt.Println(trace)
		trace:=gitlab_api.Get_job_trace(project_id,job_id,"2018011302",accessToken)
		tests,outputs := os.Grade(trace)
		database.CreateGradeRecord(webhookInfo.UserInfo.UserID,webhookInfo.Attributes.Branch,tests,outputs,webhookInfo.Attributes.Detail_status)
	}
	// tests,outputs := os.Grade(trace)
	// if detailed_status=="passed" || detailed_status=="failed"{
	// 	database.CreateGradeRecord(webhookInfo.UserInfo.UserID,webhookInfo.Attributes.Branch,tests,outputs,detailed_status)
	// }
	c.JSON(http.StatusOK,gin.H{
		"message":"success",
	})
}
