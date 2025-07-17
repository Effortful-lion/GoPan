package http

import (
	"filestore-server/app/gateway/rpc"
	"filestore-server/idl/user/userPb"
	"filestore-server/util"
	"fmt"
	"log"
	"net/http"
	"os"
)

// 返回注册页面
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &userPb.SignupRequest{}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")
	req.UserName = username
	req.Password = password

	res, err := rpc.SignupHandlerPost(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte(res.Msg))
}

func GetSignupHandler(w http.ResponseWriter, r *http.Request) {
	// 返回注册页面
	data, err := os.ReadFile("./static/view/signup.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func GetSignInHandler(w http.ResponseWriter, r *http.Request) {
	data, err := os.ReadFile("./static/view/signin.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 登录接口
func SignInHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &userPb.SigninRequest{}
	r.ParseForm()
	username := r.Form.Get("username")
	password := r.Form.Get("password")

	req.UserName = username
	req.Password = password
	res, err := rpc.SignInHandlerPost(ctx, req)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	log.Println("rpc.SignInHandlerPost执行过了")
	log.Println(res)
	log.Println(res.Data)
	data := res.Data
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: data.Location,
			Username: username,
			Token:    data.Token,
		},
	}
	w.Write(resp.JSONBytes())
}

// 获取用户信息
func UserInfoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req := &userPb.UserInfoRequest{}
	r.ParseForm()
	username := r.Form.Get("username")
	req.UserName = username
	res, err := rpc.UserInfoHandler(ctx, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: res,
	}
	w.Write(resp.JSONBytes())
}
