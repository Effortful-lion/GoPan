package handler

import (
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"net/http"
	"os"
	"time"
)

const (
	pwd_salt = "*#890"
)

// 处理用户注册请求
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回注册页面
		data, err := os.ReadFile("./static/view/signup.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)

	} else if r.Method == "POST" {
		// 处理注册逻辑
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		if len(username) < 3 || len(password) < 5 {
			w.Write([]byte("Invalid parameter"))
			return
		}
		// 加密处理
		enc_pwd := util.Sha1([]byte(password + pwd_salt))
		suc := dblayer.UserSignup(username, enc_pwd)
		if suc {
			w.Write([]byte("SUCCESS"))
		} else {
			w.Write([]byte("FAILED"))
		}
	}
}

// 登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// 返回登录页面
		data, err := os.ReadFile("./static/view/signin.html")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(data)
	} else if r.Method == "POST" {
		r.ParseForm()
		username := r.Form.Get("username")
		password := r.Form.Get("password")
		encPwd := util.Sha1([]byte(password + pwd_salt))
		// 校验密码
		pwdChecked := dblayer.UserSignin(username, encPwd)
		if !pwdChecked {
			w.Write([]byte("Password FAILED"))
			return
		}
		// 生成token
		token := GenToken(username)
		upRes := dblayer.UpdateToken(username, token)
		if !upRes {
			w.Write([]byte("token FAILED"))
			return
		}
		// 重定向到首页
		//w.Write([]byte("http://" + r.Host + "/static/view/home.html"))
		location := "http://" + r.Host + "/static/view/home.html"
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
		w.Write(resp.JSONBytes())
	}
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
