# README

### `unilab-backend/auth`: Oauth2认证模块

基于`golang.org/x/oauth2`

**HTTP API:**

`/Login`:登陆入口，跳转进行oauth认证

`/Callback`:oauth回调函数，获取认证码`code`,请求获得`access_token`

**FUNC API:**

`GetUserInfo()(int,map[string]interface{})`:返回（状态码code，通过oauth认证得到的用户信息）。



### `unilab-backend/gitlab_api`: Gitlab仓库相关API

| `get(urls string, data map[string]string, access_token string) ([]map[string]interface{},string)` | Gitlab api v4 get请求通用接口                                |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| **`post(urls string, data map[string]string, access_token string) ([]map[string]interface{},string)`** | **Gitlab api v4 post请求通用接口**                           |
| **`Get_project_info(project_name string) (string,string,string)`** | **根据project_name匹配仓库名获取（id,name,name_with_namespcae）** |
| **`Get_branches(project_id string) []map[string] interface{}`** | **获取仓库分支信息**                                         |
| **`Get_pipeline_jobs_info(project_id string,pipeline_id string)[]map[string]interface{}`** | **获取仓库CI信息**                                           |
| **`Get_job_trace(project_id string,job_id string)`**         | **获取某个CI job的结果trace**                                |
| **`Get_project_traces(project_name string)`**                | **获取一个仓库所有的CI结果**                                 |
