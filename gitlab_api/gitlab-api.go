package gitlab_api

import(
	"fmt"
	"strconv"
	"bytes"
	"encoding/json"
	"unilab-backend/auth"
	"net/http"
	"net/url"
	"io/ioutil"
)

const API_ROOT="https://git.tsinghua.edu.cn/api/v4"

func get(urls string, data map[string]string, access_token string) ([]map[string]interface{},string){
	client := &http.Client{}
	params := url.Values{}
	Url,err := url.Parse(API_ROOT+urls)
	if err!=nil {
		return nil,""
	}
	for key,value := range data{
		params.Set(key,value)
	}
	Url.RawQuery = params.Encode()
	req,_ :=http.NewRequest("GET",Url.String(),nil)
	req.Header.Add("Authorization","Bearer "+access_token)
	req.Header.Add("seach","2018011302")
	response,_ :=client.Do(req)
	body,_ := ioutil.ReadAll(response.Body)
	var result []map[string]interface{}
	_=json.Unmarshal(body,&result)
	return result,string(body)
}

func post(url string, data map[string]interface{}, access_token string) map[string]interface{} {
	client := &http.Client{}
	bytesData,_ := json.Marshal(data)
	req,_ := http.NewRequest("POST",API_ROOT+url,bytes.NewReader(bytesData))
	response,_ := client.Do(req)
	body,_ := ioutil.ReadAll(response.Body)
	var result map[string]interface{}
	_=json.Unmarshal(body,&result)
	return result
}

func Get_project_info(project_name string) (string,string,string){
	data := make(map[string] string)
	data["search"]=project_name
	data["search_namespaces"]="true"
	// data["page"]="1"
	// data["per_page"]="1"
	projects,_ :=get("/projects",data,auth.AccessToken)
	project:=projects[0]
	// fmt.Println("id: ",projects["id"])
	// fmt.Println("name: ",projects["name"])
	return strconv.Itoa(int(project["id"].(float64))),project["name"].(string),project["name_with_namespace"].(string)
}

func Get_branches(project_id string) []map[string] interface{}{
	data := make(map[string]string)
	branches,_ :=get("/projects/"+project_id+"/repository/branches",data,auth.AccessToken)
	return branches
}

func Get_pipelines_info(project_id string,branch string) []map[string]interface{}{
	data := make(map[string]string)
	data["ref"]=branch
	pipelines,_ :=get("/projects/"+project_id+"/pipelines",data,auth.AccessToken)
	// fmt.Print(pipelines)
	return pipelines
}

func Get_pipeline_jobs_info(project_id string,pipeline_id string)[]map[string]interface{}{ 
	data := make(map[string]string)
	jobs,_ :=get("/projects/"+project_id+"/pipelines/"+pipeline_id+"/jobs",data,auth.AccessToken)
	return jobs
}

func Get_job_trace(project_id string,job_id string) string{
	data := make(map[string]string)
	fmt.Println("/projects/"+project_id+"/jobs/"+job_id+"/trace")
	_,trace :=get("/projects/"+project_id+"/jobs/"+job_id+"/trace",data,auth.AccessToken)
	// fmt.Println(trace)
	return string(trace)
}

func Get_project_traces(project_name string) string{
	id,_,_:=Get_project_info(project_name)
	// fmt.Println(id)
	// branches:=Get_branches(id)
	// for _,branch :=range branches{
	// 	fmt.Println(branch["name"])
	// }
	// branch:=branches[1]["name"].(string)
	// pipelines:=Get_pipelines_info(id,branch)
	pipelines:=Get_pipelines_info(id,"ch7")
	pipline:=strconv.Itoa(int(pipelines[0]["id"].(float64)))
	jobs:=Get_pipeline_jobs_info(id,pipline)
	job:=strconv.Itoa(int(jobs[0]["id"].(float64)))
	fmt.Println(job)
	trace:=Get_job_trace(id,job)
	return trace
}