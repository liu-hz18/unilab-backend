package gitlabapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"unilab-backend/setting"
)

func get(urls string, data map[string]string, userid string, accessToken string) ([]map[string]interface{}, string) {
	client := &http.Client{}
	params := url.Values{}
	url, err := url.Parse(setting.GitLabBaseURL + urls)
	if err != nil {
		return nil, ""
	}
	for key, value := range data {
		params.Set(key, value)
	}
	url.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", url.String(), nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("search", userid)
	response, _ := client.Do(req)
	body, _ := io.ReadAll(response.Body)
	var result []map[string]interface{}
	_ = json.Unmarshal(body, &result)
	return result, string(body)
}

func post(url string, data map[string]interface{}) map[string]interface{} {
	client := &http.Client{}
	bytesData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", setting.GitLabBaseURL+url, bytes.NewReader(bytesData))
	response, _ := client.Do(req)
	body, _ := io.ReadAll(response.Body)
	var result map[string]interface{}
	_ = json.Unmarshal(body, &result)
	return result
}

func GetProjectInfo(projectName string, userid string, accessToken string) (string, string, string) {
	data := make(map[string]string)
	data["search"] = projectName
	data["search_namespaces"] = "true"
	// data["page"]="1"
	// data["per_page"]="1"
	projects, _ := get("/projects", data, userid, accessToken)
	if len(projects) == 0 {
		return "", "", ""
	}
	project := projects[0]
	// logging.Info("id: ", int(project["id"].(float64)))
	// logging.Info("name: ", project["name"].(string))
	return strconv.Itoa(int(project["id"].(float64))), project["name"].(string), project["name_with_namespace"].(string)
}

func GetBranches(projectID string, userid string, accessToken string) []map[string]interface{} {
	data := make(map[string]string)
	branches, _ := get("/projects/"+projectID+"/repository/branches", data, userid, accessToken)
	return branches
}

func GetPipelinesInfo(projectID string, branch string, userid string, accessToken string) []map[string]interface{} {
	data := make(map[string]string)
	data["ref"] = branch
	pipelines, _ := get("/projects/"+projectID+"/pipelines", data, userid, accessToken)
	// logging.Info(pipelines)
	return pipelines
}

func GetPipelineJobsInfo(projectID string, piplineID string, userid string, accessToken string) []map[string]interface{} {
	data := make(map[string]string)
	jobs, _ := get("/projects/"+projectID+"/pipelines/"+piplineID+"/jobs", data, userid, accessToken)
	// logging.Info(jobs)
	return jobs
}

func GetJobTrace(projectID string, jobID string, userid string, accessToken string) string {
	data := make(map[string]string)
	// logging.Info("/projects/"+projectID+"/jobs/"+jobID+"/trace")
	_, trace := get("/projects/"+projectID+"/jobs/"+jobID+"/trace", data, userid, accessToken)
	// logging.Info(trace)
	return string(trace)
}

func GetProjectTraces(projectName string, userID string, accessToken string) map[string]string {
	projectID, _, _ := GetProjectInfo(projectName, userID, accessToken)
	if projectID == "" {
		return nil
	}
	traces := make(map[string]string)

	branches := GetBranches(projectID, userID, accessToken)
	for _, branch := range branches {
		pipelines := GetPipelinesInfo(projectID, branch["name"].(string), userID, accessToken)
		// piplineID := strconv.Itoa(int(pipelines[0]["id"].(float64)))
		for _, pipeline := range pipelines {
			if pipeline["status"].(string) == "success" {
				fmt.Println(branch["name"])
				piplineID := strconv.Itoa(int(pipeline["id"].(float64)))
				jobs := GetPipelineJobsInfo(projectID, piplineID, userID, accessToken)
				jobID := strconv.Itoa(int(jobs[0]["id"].(float64)))
				trace := GetJobTrace(projectID, jobID, userID, accessToken)
				traces[branch["name"].(string)] = trace
				break
			}
		}
	}
	return traces
}
