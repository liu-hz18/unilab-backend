package webhook

import (
	"net/http"
	"strconv"
	"unilab-backend/apis"
	"unilab-backend/database"
	"unilab-backend/gitlabapi"
	"unilab-backend/logging"
	"unilab-backend/osgrade"

	"github.com/gin-gonic/gin"
)

type WebhookInfo struct {
	Attributes struct {
		Branch       string `json:"ref"`
		Status       string `json:"status"`
		DetailStatus string `json:"detailed_status"`
	} `json:"object_attributes"`
	UserInfo struct {
		UserID uint32 `json:"id"`
	} `json:"user"`
	ProjectInfo struct {
		ProjectID uint32 `json:"id"`
	} `json:"project"`
	JobInfos []JobInfo `json:"builds"`
}

type JobInfo struct {
	JobID uint32 `json:"id"`
}

func OsWebhookHandler(c *gin.Context) {
	logging.Info("Start Webhook")
	var webhookInfo WebhookInfo
	if err := c.ShouldBindJSON(&webhookInfo); err != nil {
		logging.Info(err)
		apis.ErrorResponse(c, apis.INVALID_PARAMS, err.Error())
		return
	}
	logging.Info(webhookInfo)
	if len(webhookInfo.JobInfos) < 1 {
		logging.Info("array webhookInfo.JobInfos is empty!")
		apis.ErrorResponse(c, apis.INVALID_PARAMS, "array webhookInfo.JobInfos is empty!")
		return
	}
	projectID := strconv.Itoa(int(webhookInfo.ProjectInfo.ProjectID))
	jobID := strconv.Itoa(int(webhookInfo.JobInfos[0].JobID))
	accessToken, err := database.GetUserAccessToken("2018011302")
	if err != nil {
		apis.ErrorResponse(c, apis.ERROR, err.Error())
		return
	}
	if webhookInfo.Attributes.DetailStatus == "passed" || webhookInfo.Attributes.DetailStatus == "failed" {
		// fmt.Println(trace)
		trace := gitlabapi.GetJobTrace(projectID, jobID, "2018011302", accessToken)
		tests, outputs := osgrade.Grade(trace)
		err := database.CreateGradeRecord(webhookInfo.UserInfo.UserID, webhookInfo.Attributes.Branch, tests, outputs, webhookInfo.Attributes.DetailStatus)
		if err != nil {
			logging.Error(err.Error())
		}
	}
	// tests,outputs := os.Grade(trace)
	// if detailed_status=="passed" || detailed_status=="failed"{
	// 	database.CreateGradeRecord(webhookInfo.UserInfo.UserID,webhookInfo.Attributes.Branch,tests,outputs,detailed_status)
	// }
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}
