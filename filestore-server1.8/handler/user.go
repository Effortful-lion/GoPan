package handler

import (
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

const (
	pwd_salt = "*#890"
)

// 处理用户注册请求
func SignupHandler(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/static/view/signup.html")
}

// 处理注册post请求
func DoSignupHandler(c *gin.Context) {
	// 处理注册逻辑
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	if len(username) < 3 || len(password) < 5 {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "Invalid parameter",
			"code": -1,
		})
		return
	}
	// 加密处理
	enc_pwd := util.Sha1([]byte(password + pwd_salt))
	suc := dblayer.UserSignup(username, enc_pwd)
	if suc {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "SUCCESS",
			"code": 0,
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "FAILED",
			"code": -2,
		})
	}
}

// 返回登录页面
func SignInHandler(c *gin.Context) {
	c.Redirect(http.StatusMovedPermanently, "/static/view/signin.html")
}

// 处理登录post接口
func DoSignInHandler(c *gin.Context) {
	username := c.Request.FormValue("username")
	password := c.Request.FormValue("password")
	encPwd := util.Sha1([]byte(password + pwd_salt))
	// 校验密码
	pwdChecked := dblayer.UserSignin(username, encPwd)
	if !pwdChecked {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "FAILED",
			"code": -2,
		})
		return
	}
	// 生成token
	token := GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		c.JSON(http.StatusOK, gin.H{
			"msg":  "FAILED",
			"code": -2,
		})
		return
	}
	// 重定向到首页
	//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
	location := "http://" + c.Request.Host + "/static/view/home.html"
	fmt.Println(location)
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: location,
			Username: username,
			Token:    token,
		},
	}
	//c.JSON(http.StatusOK, resp)
	c.Data(http.StatusOK, "text/html; charset=utf-8", resp.JSONBytes())
}

// 获取用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	//token := r.Form.Get("token")
	//// 2. 验证token是否有效
	//isValidToken := IsTokenVaild(token)
	//if !isValidToken {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}
	// 3. 查询用户信息
	// 4. 组装并返回响应数据
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		w.WriteHeader(http.StatusForbidden)
		return
	}
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}

// token 是否有效
func IsTokenVaild(token string) bool {
	// TODO: 判断token是否有效：是否过期：后8位
	// TODO：从数据库表tbl_user_token查询username对应的token信息
	// TODO：对比两个token是否一致
	if len(token) != 40 {
		return false
	}
	return true
}

func GenToken(username string) string {
	// 32 + 8 = 40
	// md5(username+timestamps+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}
