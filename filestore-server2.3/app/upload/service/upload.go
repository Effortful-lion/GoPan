package service

import (
	"context"
	"encoding/json"
	"errors"
	"filestore-server/app/transfer/mq"
	"filestore-server/app/upload/rpc"
	"filestore-server/common"
	cfg "filestore-server/config"
	"filestore-server/idl/dbproxy/dbproxyPb"
	"filestore-server/idl/upload/uploadPb"
	"filestore-server/util"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type UploadService struct {
	uploadPb.UnimplementedUploadServiceServer
}

func NewUploadService() *UploadService {
	return &UploadService{}
}

// 上传文件的处理函数
func (s *UploadService) UploadHandlerPost(ctx context.Context, req *uploadPb.UploadRequest) (*uploadPb.UploadResponse, error) {
	username := req.GetUsername()
	fileMeta_req := req.FileMeta
	fileMeta := common.FileMeta{
		FileName: fileMeta_req.FileName,
		FileSize: fileMeta_req.FileSize,
		Location: fileMeta_req.Location,
		UploadAt: fileMeta_req.UploadAt,
		FileSha1: fileMeta_req.FileHash,
	}

	// 根据存储类型进行存储操作
	if cfg.CurrentStoreType == common.StoreCeph {
		// TODO 同时将文件写入 ceph 存储
		ext := filepath.Ext(fileMeta_req.FileName)
		cephPath := "ceph/" + fileMeta.FileSha1 + ext
		data := mq.TransferData{
			FileHash:      fileMeta.FileSha1,
			CurLocation:   fileMeta.Location,
			DestLocation:  cephPath,
			DestStoreType: common.StoreCeph,
		}
		pubData, _ := json.Marshal(data)
		suc := mq.Publish(
			cfg.Config.RabbitMQConfig.TransExchangeName,
			cfg.Config.RabbitMQConfig.TransOSSRoutingKey,
			pubData)
		if !suc {
			// TODO 上传失败，重试

		}
	} else if cfg.CurrentStoreType == common.StoreOSS {
		// TODO 同时将文件写入 oss 存储
		ext := filepath.Ext(fileMeta_req.FileName)
		ossPath := "oss/" + fileMeta.FileSha1 + ext
		data := mq.TransferData{
			FileHash:      fileMeta.FileSha1,
			CurLocation:   fileMeta.Location,
			DestLocation:  ossPath,
			DestStoreType: common.StoreOSS,
		}
		pubData, _ := json.Marshal(data)
		suc := mq.Publish(
			cfg.Config.RabbitMQConfig.TransExchangeName,
			cfg.Config.RabbitMQConfig.TransOSSRoutingKey,
			pubData)
		if !suc {
			// TODO 上传失败，重试

		}
	}

	_, err := rpc.OnFileUploadFinished(ctx, &dbproxyPb.OnFileUploadFinishedRequest{
		Filehash: fileMeta.FileSha1,
		Filename: fileMeta.FileName,
		Filesize: fileMeta.FileSize,
		Fileaddr: fileMeta.Location,
	})

	// TODO 更新用户文件表
	_, err = rpc.OnUserFileUploadFinished(ctx, &dbproxyPb.OnUserFileUploadFinishedRequest{
		Username: username,
		Filename: fileMeta.FileName,
		Filehash: fileMeta.FileSha1,
		Filesize: fileMeta.FileSize,
	})
	if err == nil {
		url := "http://127.0.0.1:8080" + "/static/view/home.html"
		return &uploadPb.UploadResponse{
			Url: url,
		}, nil
	} else {
		return nil, errors.New("Upload Failed")
	}
}

func (s *UploadService) TryFastUploadHandler(ctx context.Context, req *uploadPb.TryFastUploadRequest) (*uploadPb.TryFastUploadResponse, error) {
	username := req.GetUsername()
	filehash := req.GetFileHash()
	filesize := req.GetFileSize()
	filename := req.GetFileName()
	// 2. 从文件表中查询相同hash的文件记录
	req2 := &dbproxyPb.GetFileMetaRequest{
		FileSha1: filehash,
	}
	fileMeta, err := rpc.GetFileMetaHandler(ctx, req2)
	if err != nil {
		fmt.Println(err.Error())
		return &uploadPb.TryFastUploadResponse{}, err
	}
	// 3. 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		return &uploadPb.TryFastUploadResponse{}, errors.New(resp.Msg)
	}
	// 4. 上传过则将文件信息写入用户文件表，返回成功
	req3 := &dbproxyPb.OnUserFileUploadFinishedRequest{
		Username: username,
		Filename: filename,
		Filehash: filehash,
		Filesize: filesize,
	}
	_, err = rpc.OnUserFileUploadFinished(ctx, req3)
	if err == nil {
		return &uploadPb.TryFastUploadResponse{}, nil
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		return &uploadPb.TryFastUploadResponse{}, errors.New(resp.Msg)
	}
}

// 初始化分块上传
func (s *UploadService) InitialMultipartUploadHandler(ctx context.Context, req *uploadPb.InitMultipartUploadRequest) (*uploadPb.InitMultipartUploadResponse, error) {
	// 1. 解析用户请求参数
	username := req.GetUsername()
	filehash := req.GetFileHash()
	filesize := int(req.GetFileSize())

	// 3. 生成分块上传的初始化信息
	chunkSize := 5 * 1024 * 1024 // 5MB
	chunkcount := int(math.Ceil(float64(filesize) / float64(chunkSize)))
	upInfo := common.MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username + fmt.Sprintf("%x", time.Now().UnixNano()), // 生成唯一标识符：用户名+时间戳（十六进制字符串）
		ChunkSize:  5 * 1024 * 1024,                                     // 5MB
		ChunkCount: chunkcount,                                          // 分块数量: 向上取整
	}
	// 4. 将初始化信息写入到redis缓存
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // 设置 5 秒超时时间
	defer cancel()                                                          // 确保在函数结束时释放上下文资源

	// HSet 方法需要传入键、字段和值
	// 将哈希表的字段和值存储在一个 map 中
	//data := map[string]interface{}{
	//	"filehash":   upInfo.FileHash,
	//	"filesize":   upInfo.FileSize,
	//	"chunksize":  upInfo.ChunkSize,
	//	"chunkcount": upInfo.ChunkCount,
	//	"username":   username,
	//}

	// 使用 map 作为参数调用 HSet 方法
	key := "MP_" + upInfo.UploadID

	_, err := rpc.CacheChunk(ctx, &dbproxyPb.CacheChunkRequest{
		Key: key,
		Data: &dbproxyPb.CacheChunkData{
			Filehash:   upInfo.FileHash,
			Filesize:   int64(upInfo.FileSize),
			Chunksize:  int32(upInfo.ChunkSize),
			Chunkcount: int32(upInfo.ChunkCount),
			Username:   username,
		},
	})

	if err != nil {
		return &uploadPb.InitMultipartUploadResponse{}, err
	}
	// 5. 将响应初始化数据返回到客户端
	return &uploadPb.InitMultipartUploadResponse{
		InitInfo: &uploadPb.MultipartUploadInitInfo{
			FileHash:   upInfo.FileHash,
			FileSize:   int64(upInfo.FileSize),
			UploadId:   upInfo.UploadID,
			ChunkSize:  int32(upInfo.ChunkSize),
			ChunkCount: int32(upInfo.ChunkCount),
		},
	}, nil
}

// 上传文件分块
func (s *UploadService) UploadPartHandler(ctx context.Context, req *uploadPb.UploadPartRequest) (*uploadPb.UploadPartResponse, error) {
	uploadID := req.GetUploadId()
	chunkIndex := req.GetChunkIndex()
	// 4. 更新redis缓存状态
	req2 := &dbproxyPb.CacheChunkUpdateRequest{
		Hashkey: "MP_" + uploadID,
		Key:     "chkidx_" + chunkIndex,
		Value:   1,
	}
	_, err := rpc.CacheChunkUpdate(ctx, req2)
	if err != nil {
		return &uploadPb.UploadPartResponse{}, err
	}
	return &uploadPb.UploadPartResponse{}, nil
}

// 通知上传合并
func (s *UploadService) CompleteUploadHandler(ctx context.Context, req *uploadPb.CompleteUploadRequest) (*uploadPb.CompleteUploadResponse, error) {
	upid := req.GetUploadId()
	username := req.GetUsername()
	filehash := req.GetFileHash()
	filesize := req.GetFileSize()
	filename := req.GetFileName()

	res1, err := rpc.CacheChunkQuery(ctx, &dbproxyPb.CacheChunkQueryRequest{
		Hashkey: "MP_" + upid,
	})
	if err != nil {
		return &uploadPb.CompleteUploadResponse{}, err
	}
	data := res1.GetData()

	totalCount := 0
	chunkCount := 0
	for k, v := range data {
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			// 存在分块索引，且值为1，表示分块上传完成
			chunkCount++
		}
	}

	if totalCount != chunkCount {
		return &uploadPb.CompleteUploadResponse{}, errors.New("分块上传不完整")
	}
	// 4. TODO 如果所有分块上传完成，合并分块
	// 步骤：
	// 1. 基于uploadid 在 ./data/uploadid/ 下获取所有文件并根据文件名排序
	// 2. 创建空文件到 ./tmp/ 下
	// 3. 将分块文件写入到空文件中
	// 4. 合并结束后删除 /data 下的分块文件 TODO （redis缓存文件删不删之后再说）
	filesPath := "./data/" + upid + "/"
	// 遍历目录下的所有文件
	files, err := os.ReadDir(filesPath)
	if err != nil {
		return &uploadPb.CompleteUploadResponse{}, err
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	// 确保./tmp/ 目录存在
	if err := os.MkdirAll("./tmp", 0744); err != nil {
		return &uploadPb.CompleteUploadResponse{}, err
	}

	tmpPath := "./tmp/" + strings.Split(filename, "/")[len(strings.Split(filename, "/"))-1]
	fmt.Println(tmpPath)
	newFile, err := os.Create(tmpPath)
	if err != nil {
		return &uploadPb.CompleteUploadResponse{}, err
	}
	defer newFile.Close()
	// 优化文件合并循环（确保文件正确关闭）
	for _, file := range files {
		chunkFile, err := os.Open(filesPath + file.Name())
		if err != nil {
			return &uploadPb.CompleteUploadResponse{}, err
		}

		// 使用闭包确保及时关闭文件
		func(f *os.File) {
			defer f.Close()
			buf := make([]byte, 1024*1024)
			for {
				n, err := f.Read(buf)
				if n > 0 {
					if _, err := newFile.Write(buf[:n]); err != nil {
						fmt.Printf("[ERROR] 写入失败: %v\n", err)
					}
				}
				if err != nil {
					break
				}
			}
		}(chunkFile) // 传入文件对象
	}
	defer func() {
		// 合并结束后删除本地临时文件
		err = os.RemoveAll(filesPath)
		if err != nil {
			fmt.Printf("[ERROR] 删除目录失败: %v\n", err)
		}
	}()

	// TODO 将合并的文件保存到本地文件系统后，将文件信息保存到唯一文件表中
	fileaddr := "./tmp/" + filename

	// 5. 更新唯一文件表及用户文件表
	_, err = rpc.OnFileUploadFinished(ctx, &dbproxyPb.OnFileUploadFinishedRequest{
		Filehash: filehash,
		Filename: filename,
		Filesize: int64(filesize),
		Fileaddr: fileaddr,
	})
	if err != nil {
		return &uploadPb.CompleteUploadResponse{}, err
	}
	_, err = rpc.OnUserFileUploadFinished(ctx, &dbproxyPb.OnUserFileUploadFinishedRequest{
		Filehash: filehash,
		Filename: filename,
		Filesize: filesize,
		Username: username,
	})
	if err != nil {
		return &uploadPb.CompleteUploadResponse{}, err
	}
	// 6. 将响应初始化数据返回到客户端
	return &uploadPb.CompleteUploadResponse{}, nil
}

// TODO 取消上传
func (s *UploadService) CancelUploadHandler(ctx context.Context, req *uploadPb.CancelUploadRequest) (*uploadPb.CancelUploadResponse, error) {
	uploadID := req.GetUploadId()
	// 2. 强制删除本地的分块文件
	err := os.RemoveAll("./data/" + uploadID)
	if err != nil {
		return &uploadPb.CancelUploadResponse{}, err
	}
	// 3. 删除redis缓存的分块信息
	key := "MP_" + uploadID
	_, err = rpc.CacheChunkDelete(ctx, &dbproxyPb.CacheChunkDeleteRequest{
		Hashkey: key,
	})
	if err != nil {
		return &uploadPb.CancelUploadResponse{}, err
	}
	// TODO 4. 更新mysql中用户上传记录（非必须）

	// 5. 返回处理结果到客户端
	return &uploadPb.CancelUploadResponse{}, nil
}

// TODO 查询上传状态（进度）
func (s *UploadService) QueryUploadStatusHandler(ctx context.Context, req *uploadPb.QueryUploadStatusRequest) (*uploadPb.QueryUploadStatusResponse, error) {
	uploadID := req.GetUploadId()
	// 2. 查询redis缓存的分块信息（chunkCount 和 chkidx_id个数）并计算
	res1, err := rpc.CacheChunkQuery(ctx, &dbproxyPb.CacheChunkQueryRequest{
		Hashkey: "MP_" + uploadID,
	})
	if err != nil {
		return &uploadPb.QueryUploadStatusResponse{}, err
	}
	data := res1.GetData()

	totalCount := 0
	chunkCount := 0
	res := make([]int, 0)
	temp := make([]int, 0)
	for k, v := range data {
		if k == "chunkcount" {
			totalCount, _ = strconv.Atoi(v)
		} else if strings.HasPrefix(k, "chkidx_") && v == "1" {
			chunkCount++
			// 拆解出分块索引
			var idx int
			_, err := fmt.Sscanf(k, "chkidx_%d", &idx)
			if err != nil {
				return &uploadPb.QueryUploadStatusResponse{}, err
			}
			temp = append(temp, idx)
		}
	}
	// 检查temp，查出缺失的序号：
	sort.Ints(temp)
	for i := 0; i < len(temp); i++ {
		// eg: temp = [1, 2, 3, 4, 5, 6, 8, 9] , 查出7不在
		if temp[i] != i+1 {
			res = append(res, i+1)
		}
	}
	// 将res转换为int32类型的数组
	lacking := make([]int32, 0)
	for _, v := range res {
		lacking = append(lacking, int32(v))
	}
	// 3. 返回处理结果到客户端
	return &uploadPb.QueryUploadStatusResponse{
		TotalCount: int32(totalCount),
		ChunkCount: int32(chunkCount),
		Progress:   fmt.Sprintf("%d%%", chunkCount*100/totalCount), // 进度
		Lacking:    lacking,                                        // 缺少的分块索引列表
	}, nil
}
