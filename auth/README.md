# README

### `unilab-backend/auth`: Oauth2认证模块

基于`golang.org/x/oauth2`

**HTTP API:**

`/Login`:登陆入口，跳转进行oauth认证

`/Callback`:oauth回调函数，获取认证码`code`,请求获得`access_token`

**FUNC API:**

`GetUserInfo()(int,map[string]interface{})`:返回（状态码code，通过oauth认证得到的用户信息）。

