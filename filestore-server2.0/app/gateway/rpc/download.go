package rpc

import (
	"context"
	"filestore-server/idl/download/downloadPb"
)

// 下载文件
func DownloadHandler(ctx context.Context, req *downloadPb.DownloadRequest) (*downloadPb.DownloadResponse, error) {
	res, err := DownloadClient.DownloadHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 生成 oss 文件的下载地址
func DownloadURLHandler(ctx context.Context, req *downloadPb.DownloadURLRequest) (*downloadPb.DownloadURLResponse, error) {
	res, err := DownloadClient.DownloadURLHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
