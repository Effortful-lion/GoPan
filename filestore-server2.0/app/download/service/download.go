package service

import (
	"context"
	"filestore-server/app/gateway/rpc"
	"filestore-server/idl/dbproxy/dbproxyPb"
	"filestore-server/idl/download/downloadPb"
	"filestore-server/store/oss"
	"fmt"
	"io"
	"os"
)

type DownloadService struct {
	downloadPb.UnimplementedDownloadServiceServer
}

func NewDownloadService() *DownloadService {
	return &DownloadService{}
}

// 下载文件
func (s *DownloadService) DownloadHandler(ctx context.Context, req *downloadPb.DownloadRequest) (*downloadPb.DownloadResponse, error) {
	fsha1 := req.GetFileHash()
	r := &dbproxyPb.GetFileMetaRequest{
		FileSha1: fsha1,
	}
	res, err := rpc.GetFileMetaHandler(ctx, r)
	fm := res.GetFileMeta()
	location := fm.Location
	// 打开文件
	file, err := os.Open(location)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	// 获取文件句柄后读取文件内容到内存
	// 如果读取文件很大，需要使用流的形式，读取部分后响应到客户端，然后刷新缓存再读取下一部分，直到文件读取完毕
	// 这里假设文件很小，直接全部读取到内存
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return &downloadPb.DownloadResponse{
		FileData: data,
	}, nil
}

// 生成 oss 文件的下载地址
func (s *DownloadService) DownloadURLHandler(ctx context.Context, req *downloadPb.DownloadURLRequest) (*downloadPb.DownloadURLResponse, error) {
	filehash := req.GetFileHash()
	// 从文件表查找记录
	r := &dbproxyPb.GetFileMetaRequest{
		FileSha1: filehash,
	}
	data, err := rpc.GetFileMetaHandler(ctx, r)
	fmeta := data.GetFileMeta()
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	// TODO 判断文件存在 oss 还是 ceph

	// 服务端签名后返回oss的下载地址
	signedURL := oss.DownloadURL(fmeta.Location)
	return &downloadPb.DownloadURLResponse{
		SignedUrl: signedURL,
	}, nil
}
