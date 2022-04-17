package gitlab_api

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"unilab-backend/auth"
	"unilab-backend/logging"
)


func get(urls string, data map[string]string, userid string, access_token string) ([]map[string]interface{},string){
	client := &http.Client{}
	params := url.Values{}
	Url,err := url.Parse(auth.GitLabAPIBaseURL + urls)
	if err!=nil {
		return nil,""
	}
	for key,value := range data{
		params.Set(key,value)
	}
	Url.RawQuery = params.Encode()
	req,_ :=http.NewRequest("GET",Url.String(),nil)
	req.Header.Add("Authorization","Bearer " + access_token)
	req.Header.Add("search", userid)
	response,_ :=client.Do(req)
	body,_ := ioutil.ReadAll(response.Body)
	var result []map[string]interface{}
	_=json.Unmarshal(body,&result)
	return result, string(body)
}

func post(url string, data map[string]interface{}) map[string]interface{} {
	client := &http.Client{}
	bytesData,_ := json.Marshal(data)
	req,_ := http.NewRequest("POST", auth.GitLabAPIBaseURL + url,bytes.NewReader(bytesData))
	response,_ := client.Do(req)
	body,_ := ioutil.ReadAll(response.Body)
	var result map[string]interface{}
	_=json.Unmarshal(body,&result)
	return result
}

func Get_project_info(project_name string, userid string, accessToken string) (string,string,string) {
	data := make(map[string] string)
	data["search"]=project_name
	data["search_namespaces"]="true"
	// data["page"]="1"
	// data["per_page"]="1"
	projects,_ := get("/projects", data, userid, accessToken)
	if projects == nil || len(projects) == 0 {
		return "", "", ""
	}
	project := projects[0]
	logging.Info("id: ", int(project["id"].(float64)))
	logging.Info("name: ", project["name"].(string))
	return strconv.Itoa(int(project["id"].(float64))),project["name"].(string),project["name_with_namespace"].(string)
}

func Get_branches(project_id string, userid string, accessToken string) []map[string] interface{}{
	data := make(map[string]string)
	branches,_ :=get("/projects/"+project_id+"/repository/branches",data, userid, accessToken)
	return branches
}

func Get_pipelines_info(project_id string, branch string, userid string, accessToken string) []map[string]interface{} {
	data := make(map[string]string)
	data["ref"]=branch
	pipelines,_ :=get("/projects/"+project_id+"/pipelines", data, userid, accessToken)
	// logging.Info(pipelines)
	return pipelines
}

func Get_pipeline_jobs_info(project_id string, pipeline_id string, userid string, accessToken string)[]map[string]interface{}{ 
	data := make(map[string]string)
	jobs,_ :=get("/projects/"+project_id+"/pipelines/"+pipeline_id+"/jobs",data, userid, accessToken)
	// logging.Info(jobs)
	return jobs
}

func Get_job_trace(project_id string, job_id string, userid string, accessToken string) string{
	data := make(map[string]string)
	logging.Info("/projects/"+project_id+"/jobs/"+job_id+"/trace")
	_,trace :=get("/projects/"+project_id+"/jobs/"+job_id+"/trace", data, userid, accessToken)
	// logging.Info(trace)
	return string(trace)
}

func Get_project_traces(project_name string, userid string, access_token string) string {
	id,_,_ := Get_project_info(project_name, userid, access_token)
	if id == "" {
		return ""
	}
	logging.Info(id)
	// branches:=Get_branches(id)
	// for _,branch :=range branches{
	// 	fmt.Println(branch["name"])
	// }
	// branch:=branches[1]["name"].(string)
	// pipelines:=Get_pipelines_info(id,branch)
	pipelines := Get_pipelines_info(id, "ch7", userid, access_token)
	pipline := strconv.Itoa(int(pipelines[0]["id"].(float64)))
	jobs := Get_pipeline_jobs_info(id,pipline, userid, access_token)
	job := strconv.Itoa(int(jobs[0]["id"].(float64)))
	// logging.Info(job)
	trace := Get_job_trace(id ,job, userid, access_token)
	return trace
}
