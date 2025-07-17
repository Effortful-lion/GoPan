package service

import (
	"context"
	dblayer "filestore-server/app/dbproxy/op"
	"filestore-server/app/gateway/rpc"
	"filestore-server/idl/dbproxy/dbproxyPb"
	"filestore-server/idl/user/userPb"
	"filestore-server/util"
	"fmt"
)

const (
	pwd_salt = "*#890"
)

type UserService struct {
	userPb.UnimplementedUserServiceServer
}

func NewUserService() *UserService {
	return &UserService{}
}

// 处理注册逻辑
func (s *UserService) SignupHandlerPost(ctx context.Context, req *userPb.SignupRequest) (*userPb.SignupResponse, error) {
	username := req.GetUserName()
	password := req.GetPassword()
	if len(username) < 3 || len(password) < 5 {
		return &userPb.SignupResponse{
			Code: -1,
			Msg:  "username or password error",
		}, fmt.Errorf("username or password error")
	}
	// 加密处理
	enc_pwd := util.Sha1([]byte(password + pwd_salt))
	_, err := rpc.UserSignUp(ctx, &dbproxyPb.SignupRequest{
		Username: username,
		Password: enc_pwd,
	})
	if err != nil {
		return &userPb.SignupResponse{
			Code: 0,
			Msg:  "SUCCESS",
		}, nil
	} else {
		return &userPb.SignupResponse{
			Code: -1,
			Msg:  "FAILED",
		}, nil
	}
}

// 登录接口
func (s *UserService) SignInHandlerPost(ctx context.Context, req *userPb.SigninRequest) (*userPb.SigninResponse, error) {
	username := req.GetUserName()
	password := req.GetPassword()
	encPwd := util.Sha1([]byte(password + pwd_salt))
	// 校验密码
	//pwdChecked := dblayer.UserSignin(username, encPwd)
	_, err := rpc.CheckPassword(ctx, &dbproxyPb.CheckPasswordRequest{
		Username: username,
		Password: encPwd,
	})
	if err != nil {
		return &userPb.SigninResponse{
			Code: -1,
			Msg:  "FAILED:" + err.Error(),
		}, nil
	}
	// 生成token
	token := util.GenToken(username)
	upRes := dblayer.UpdateToken(username, token)
	if !upRes {
		return &userPb.SigninResponse{
			Code: -1,
			Msg:  "FAILED" + err.Error(),
		}, nil
	}
	// 重定向到首页
	location := "/static/view/home.html"
	return &userPb.SigninResponse{
		Code: 0,
		Msg:  "OK",
		Data: &userPb.SigninData{
			Location: location,
			UserName: username,
			Token:    token,
		},
	}, nil
}

// 获取用户信息
func (s *UserService) UserInfoHandler(ctx context.Context, req *userPb.UserInfoRequest) (*userPb.UserInfoResponse, error) {
	// 1. 解析请求参数
	username := req.GetUserName()
	// 3. 查询用户信息
	// 4. 组装并返回响应数据
	user, err := dblayer.GetUserInfo(username)
	if err != nil {
		return nil, err
	}
	return &userPb.UserInfoResponse{
		Code: 0,
		Msg:  "OK",
		User: &userPb.User{
			UserName:     user.UserName,
			Email:        user.Email,
			Phone:        user.Phone,
			SignupAt:     user.SignupAt,
			LastActiveAt: user.LastActiveAt,
			Status:       int32(user.Status),
		},
	}, nil
}
