package rpc

import (
	"context"
	"filestore-server/idl/upload/uploadPb"
)

type UploadService struct {
	uploadPb.UnimplementedUploadServiceServer
}

// 上传文件的处理函数
func UploadHandlerPost(ctx context.Context, req *uploadPb.UploadRequest) (*uploadPb.UploadResponse, error) {
	res, err := UploadClient.UploadHandlerPost(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 尝试秒传接口
func TryFastUploadHandler(ctx context.Context, req *uploadPb.TryFastUploadRequest) (*uploadPb.TryFastUploadResponse, error) {
	res, err := UploadClient.TryFastUploadHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 初始化分块上传
func InitialMultipartUploadHandler(ctx context.Context, req *uploadPb.InitMultipartUploadRequest) (*uploadPb.InitMultipartUploadResponse, error) {
	res, err := UploadClient.InitMultipartUpload(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 上传文件分块
func UploadPartHandler(ctx context.Context, req *uploadPb.UploadPartRequest) (*uploadPb.UploadPartResponse, error) {
	res, err := UploadClient.UploadPartHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 通知上传合并
func CompleteUploadHandler(ctx context.Context, req *uploadPb.CompleteUploadRequest) (*uploadPb.CompleteUploadResponse, error) {
	res, err := UploadClient.CompleteUploadHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO 取消上传
func CancelUploadHandler(ctx context.Context, req *uploadPb.CancelUploadRequest) (*uploadPb.CancelUploadResponse, error) {
	res, err := UploadClient.CancelUploadHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// TODO 查询上传状态（进度）
func QueryUploadStatusHandler(ctx context.Context, req *uploadPb.QueryUploadStatusRequest) (*uploadPb.QueryUploadStatusResponse, error) {
	res, err := UploadClient.QueryUploadStatusHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
