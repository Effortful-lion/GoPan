package rpc

import (
	"context"
	"filestore-server/idl/user/userPb"
	"log"
)

// 处理注册逻辑
func SignupHandlerPost(ctx context.Context, req *userPb.SignupRequest) (*userPb.SignupResponse, error) {
	res, err := UserClient.SignupHandlerPost(ctx, req)
	return res, err
}

// 登录接口
func SignInHandlerPost(ctx context.Context, req *userPb.SigninRequest) (*userPb.SigninResponse, error) {
	log.Println("SignInHandlerPost")
	res, err := UserClient.SignInHandlerPost(ctx, req)
	return res, err
}

// 获取用户信息
func UserInfoHandler(ctx context.Context, req *userPb.UserInfoRequest) (*userPb.UserInfoResponse, error) {
	res, err := UserClient.UserInfoHandler(ctx, req)
	return res, err
}
