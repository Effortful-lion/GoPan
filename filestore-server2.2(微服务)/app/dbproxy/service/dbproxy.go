package service

import (
	"context"
	"errors"
	"filestore-server/app/dbproxy/meta"
	dblayer "filestore-server/app/dbproxy/op"
	"filestore-server/app/dbproxy/redis"
	"filestore-server/idl/dbproxy/dbproxyPb"
	"log"
	"os"
	"time"
)

type DBProxyService struct {
	dbproxyPb.UnimplementedDBProxyServiceServer
}

func NewDBProxyService() *DBProxyService {
	return &DBProxyService{}
}

// 存储、缓存服务：提供存储、查询、缓存功能

// 通过哈希值获取文件元信息的处理函数
func (s *DBProxyService) GetFileMetaHandler(ctx context.Context, req *dbproxyPb.GetFileMetaRequest) (*dbproxyPb.GetFileMetaResponse, error) {
	filehash := req.GetFileSha1()
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		return nil, err
	}
	return &dbproxyPb.GetFileMetaResponse{
		FileMeta: &dbproxyPb.FileMeta{
			FileName: fMeta.FileName,
			FileSize: fMeta.FileSize,
			Location: fMeta.Location,
			UploadAt: fMeta.UploadAt,
			FileSha1: fMeta.FileSha1,
		},
	}, nil
}

// 更新元信息（这里只是重命名）
func (s *DBProxyService) FileMetaUpdateHandler(ctx context.Context, req *dbproxyPb.FileMetaUpdateRequest) (*dbproxyPb.FileMetaUpdateResponse, error) {
	opType := req.OpType
	fileSha1 := req.GetFileSha1()
	newFileName := req.GetFileName()
	fileaddr := req.GetLocation()
	fileSize := req.GetFileSize()
	uploadAt := req.GetUploadAt()

	if opType != "0" {
		return nil, nil
	}

	curFileMeta := meta.GetFileMeta(fileSha1)
	// 如果不为空就赋值：
	if newFileName != "" {
		curFileMeta.FileName = newFileName
	}
	if fileaddr != "" {
		curFileMeta.Location = fileaddr
	}
	if fileSize > 0 {
		curFileMeta.FileSize = fileSize
	}
	if uploadAt != "" {
		curFileMeta.UploadAt = uploadAt
	}

	return &dbproxyPb.FileMetaUpdateResponse{
		FileMeta: &dbproxyPb.FileMeta{
			FileName: newFileName,
			FileSha1: fileSha1,
			FileSize: curFileMeta.FileSize,
			Location: curFileMeta.Location,
			UploadAt: curFileMeta.UploadAt,
		}}, nil
}

// 文件删除接口
func (s *DBProxyService) FileDeleteHandler(ctx context.Context, req *dbproxyPb.FileDeleteRequest) (*dbproxyPb.FileDeleteResponse, error) {
	// 删除内存和物理文件
	fileSha1 := req.GetFileSha1()

	// 删除物理文件
	fMeta := meta.GetFileMeta(fileSha1)
	os.Remove(fMeta.Location)
	// 删除内存文件
	meta.RemoveFileMeta(fileSha1)
	return &dbproxyPb.FileDeleteResponse{}, nil
}

// 批量获取文件信息
func (s *DBProxyService) FileQueryHandler(ctx context.Context, req *dbproxyPb.FileQueryRequest) (*dbproxyPb.FileQueryResponse, error) {
	limitCnt := req.GetLimit()
	username := req.GetUserName()
	userfiles, err := dblayer.QueryUserFileMetas(username, int(limitCnt))
	if err != nil {
		return nil, err
	}
	userfilesMeta := make([]*dbproxyPb.UserFileMeta, 0, len(userfiles))
	for _, userfile := range userfiles {
		userfilesMeta = append(userfilesMeta, &dbproxyPb.UserFileMeta{
			UserName: userfile.UserName,
			FileMeta: &dbproxyPb.FileMeta{
				FileName: userfile.FileName,
				FileSize: userfile.FileSize,
				UploadAt: userfile.UploadAt,
				FileSha1: userfile.FileHash,
			},
			LastUpdated: userfile.LastUpdated,
		})
	}
	return &dbproxyPb.FileQueryResponse{
		UserFiles: userfilesMeta,
	}, nil
}

// 更新文件的meta（所有）
func (s *DBProxyService) FileMetaUpdateAll(ctx context.Context, req *dbproxyPb.UpdateFileMetaRequest) (*dbproxyPb.UpdateFileMetaResp, error) {
	req_meta := req.GetFileMeta()
	filehash := req_meta.GetFileSha1()
	filemeta := meta.GetFileMeta(filehash)
	filemeta.FileName = req_meta.FileName
	filemeta.FileSize = req_meta.FileSize
	filemeta.Location = req_meta.Location
	filemeta.UploadAt = req_meta.UploadAt
	filemeta.FileSha1 = req_meta.FileSha1
	meta.UpdateFileMetaDB(filemeta)
	return &dbproxyPb.UpdateFileMetaResp{
		FileMeta: &dbproxyPb.FileMeta{
			FileName: filemeta.FileName,
			FileSize: filemeta.FileSize,
			UploadAt: filemeta.UploadAt,
			FileSha1: filemeta.FileSha1,
			Location: filemeta.Location,
		},
	}, nil
}

func (s *DBProxyService) OnUserFileUploadFinished(ctx context.Context, req *dbproxyPb.OnUserFileUploadFinishedRequest) (*dbproxyPb.OnUserFileUploadFinishedResp, error) {
	username := req.GetUsername()
	filehash := req.GetFilehash()
	filename := req.GetFilename()
	filesize := req.GetFilesize()
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, filesize)
	if suc {
		return &dbproxyPb.OnUserFileUploadFinishedResp{}, nil
	} else {
		return &dbproxyPb.OnUserFileUploadFinishedResp{}, errors.New("OnUserFileUploadFinished: file upload failed")
	}
}

func (s *DBProxyService) CacheChunk(ctx context.Context, req *dbproxyPb.CacheChunkRequest) (*dbproxyPb.CacheChunkResp, error) {
	key := req.GetKey()
	data := req.GetData()
	rConn := redis.RedisClient()
	err := rConn.HMSet(ctx, key, data).Err()
	if err != nil {
		return nil, err
	}
	return &dbproxyPb.CacheChunkResp{}, nil
}

func (s *DBProxyService) CacheChunkUpdate(ctx context.Context, req *dbproxyPb.CacheChunkUpdateRequest) (*dbproxyPb.CacheChunkUpdateResp, error) {
	hashkey := req.GetHashkey()
	key := req.GetKey()
	data := req.GetValue()
	rConn := redis.RedisClient()
	// 4. 更新redis缓存状态
	err := rConn.HSet(context.Background(), hashkey, key, data).Err()
	if err != nil {
		return nil, err
	}
	return &dbproxyPb.CacheChunkUpdateResp{}, nil
}

func (s *DBProxyService) CacheChunkQuery(ctx context.Context, req *dbproxyPb.CacheChunkQueryRequest) (*dbproxyPb.CacheChunkQueryResp, error) {
	hashkey := req.GetHashkey()
	rConn := redis.RedisClient()
	data, err := rConn.HGetAll(ctx, hashkey).Result()
	if err != nil {
		return nil, err
	}
	return &dbproxyPb.CacheChunkQueryResp{
		Data: data,
	}, nil
}

func (s *DBProxyService) OnFileUploadFinished(ctx context.Context, req *dbproxyPb.OnFileUploadFinishedRequest) (*dbproxyPb.OnFileUploadFinishedResp, error) {
	ok := dblayer.OnFileUploadFinished(req.GetFilehash(), req.GetFilename(), req.GetFilesize(), req.GetFileaddr())
	if ok {
		return &dbproxyPb.OnFileUploadFinishedResp{}, nil
	}
	return &dbproxyPb.OnFileUploadFinishedResp{}, errors.New("OnFileUploadFinished: file upload failed")
}

func (s *DBProxyService) CacheChunkDelete(ctx context.Context, req *dbproxyPb.CacheChunkDeleteRequest) (*dbproxyPb.CacheChunkDeleteResp, error) {
	key := req.GetHashkey() // 3. 删除redis缓存的分块信息
	rConn := redis.RedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err := rConn.Del(ctx, key).Err()
	if err != nil {
		return nil, err
	}
	return &dbproxyPb.CacheChunkDeleteResp{}, nil
}

func (s *DBProxyService) CheckPassword(ctx context.Context, req *dbproxyPb.CheckPasswordRequest) (*dbproxyPb.CheckPasswordResponse, error) {
	log.Println("CheckPassword")
	username := req.GetUsername()
	password := req.GetPassword()
	ok := dblayer.UserSignin(username, password)
	if ok {
		return &dbproxyPb.CheckPasswordResponse{}, nil
	}
	return &dbproxyPb.CheckPasswordResponse{}, errors.New("CheckPassword: password check failed")
}

func (s *DBProxyService) UserSignUp(ctx context.Context, req *dbproxyPb.SignupRequest) (*dbproxyPb.SignupResponse, error) {
	username := req.GetUsername()
	password := req.GetPassword()
	ok := dblayer.UserSignup(username, password)
	if ok {
		return &dbproxyPb.SignupResponse{}, nil
	}
	return &dbproxyPb.SignupResponse{}, errors.New("UserSignUp: user signup failed")
}

func (s *DBProxyService) UpdateToken(ctx context.Context, req *dbproxyPb.UpdateTokenRequest) (*dbproxyPb.UpdateTokenResp, error) {
	username := req.GetUsername()
	token := req.GetToken()
	ok := dblayer.UpdateToken(username, token)
	if ok {
		return &dbproxyPb.UpdateTokenResp{}, nil
	}
	return &dbproxyPb.UpdateTokenResp{}, errors.New("UpdateToken: token update failed")
}

func (s *DBProxyService) GetUserInfo(ctx context.Context, req *dbproxyPb.GetUserInfoRequest) (*dbproxyPb.UserInfoResp, error) {
	username := req.GetUsername()
	userinfo, err := dblayer.GetUserInfo(username)
	if err != nil {
		return nil, err
	}
	return &dbproxyPb.UserInfoResp{
		Username:     userinfo.UserName,
		Email:        userinfo.Email,
		Phone:        userinfo.Phone,
		SignupAt:     userinfo.SignupAt,
		LastActiveAt: userinfo.LastActiveAt,
		Status:       int32(userinfo.Status),
	}, nil
}

func (s *DBProxyService) UpdateFileLocation(ctx context.Context, req *dbproxyPb.UpdateFileLocationRequest) (*dbproxyPb.UpdateFileLocationResp, error) {
	filehash := req.GetFilehash()
	fileaddr := req.GetFileaddr()
	ok := dblayer.UpdateFileLocation(filehash, fileaddr)
	if ok {
		return &dbproxyPb.UpdateFileLocationResp{}, nil
	}
	return &dbproxyPb.UpdateFileLocationResp{}, errors.New("UpdateFileLocation: file location update failed")
}
