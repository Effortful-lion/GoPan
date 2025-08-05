package rpc

import (
	"context"
	"errors"
	"filestore-server/idl/dbproxy/dbproxyPb"
)

// 通过哈希值获取文件元信息的处理函数
func GetFileMetaHandler(ctx context.Context, req *dbproxyPb.GetFileMetaRequest) (*dbproxyPb.GetFileMetaResponse, error) {
	res, err := DBProxyClient.GetFileMetaHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 更新元信息（这里只是重命名）
func FileMetaUpdateHandler(ctx context.Context, req *dbproxyPb.FileMetaUpdateRequest) (*dbproxyPb.FileMetaUpdateResponse, error) {
	res, err := DBProxyClient.FileMetaUpdateHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 文件删除接口
func FileDeleteHandler(ctx context.Context, req *dbproxyPb.FileDeleteRequest) (*dbproxyPb.FileDeleteResponse, error) {
	res, err := DBProxyClient.FileDeleteHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 批量获取文件信息
func FileQueryHandler(ctx context.Context, req *dbproxyPb.FileQueryRequest) (*dbproxyPb.FileQueryResponse, error) {
	res, err := DBProxyClient.FileQueryHandler(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func OnUserFileUploadFinished(ctx context.Context, req *dbproxyPb.OnUserFileUploadFinishedRequest) (*dbproxyPb.OnUserFileUploadFinishedResp, error) {
	res, err := DBProxyClient.OnUserFileUploadFinished(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CacheChunk(ctx context.Context, req *dbproxyPb.CacheChunkRequest) (*dbproxyPb.CacheChunkResp, error) {
	res, err := DBProxyClient.CacheChunk(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CacheChunkUpdate(ctx context.Context, req *dbproxyPb.CacheChunkUpdateRequest) (*dbproxyPb.CacheChunkUpdateResp, error) {
	res, err := DBProxyClient.CacheChunkUpdate(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
func CacheChunkQuery(ctx context.Context, req *dbproxyPb.CacheChunkQueryRequest) (*dbproxyPb.CacheChunkQueryResp, error) {
	res, err := DBProxyClient.CacheChunkQuery(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func OnFileUploadFinished(ctx context.Context, req *dbproxyPb.OnFileUploadFinishedRequest) (*dbproxyPb.OnFileUploadFinishedResp, error) {
	res, err := DBProxyClient.OnFileUploadFinished(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CacheChunkDelete(ctx context.Context, req *dbproxyPb.CacheChunkDeleteRequest) (*dbproxyPb.CacheChunkDeleteResp, error) {
	res, err := DBProxyClient.CacheChunkDelete(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func CheckPassword(ctx context.Context, req *dbproxyPb.CheckPasswordRequest) (*dbproxyPb.CheckPasswordResponse, error) {
	if DBProxyClient == nil {
		return nil, errors.New("DBProxy client is nil")
	}
	res, err := DBProxyClient.CheckPassword(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func UserSignUp(ctx context.Context, req *dbproxyPb.SignupRequest) (*dbproxyPb.SignupResponse, error) {
	res, err := DBProxyClient.UserSignUp(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func UpdateToken(ctx context.Context, req *dbproxyPb.UpdateTokenRequest) (*dbproxyPb.UpdateTokenResp, error) {
	res, err := DBProxyClient.UpdateToken(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func GetUserInfo(ctx context.Context, req *dbproxyPb.GetUserInfoRequest) (*dbproxyPb.UserInfoResp, error) {
	res, err := DBProxyClient.GetUserInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func UpdateFileLocation(ctx context.Context, req *dbproxyPb.UpdateFileLocationRequest) (*dbproxyPb.UpdateFileLocationResp, error) {
	res, err := DBProxyClient.UpdateFileLocation(ctx, req)
	if err != nil {
		return nil, err
	}
	return res, nil
}
