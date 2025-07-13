package handler

import (
	"context"
	"filestore-server/cache/redis"
	dblayer "filestore-server/db"
	"filestore-server/util"
	"fmt"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

// mutiply upload:分块上传

// 分块上传初始化信息
type MultipartUploadInfo struct {
	FileHash   string // 文件hash值
	FileSize   int    // 文件总大小
	UploadID   string // 分块上传的唯一标识符
	ChunkSize  int    // 分块大小
	ChunkCount int    // 分块数量
}

// 初始化分块上传
func InitialMultipartUploasdHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}
	// 2. 获得redis的一个连接
	rConn := redis.RedisClient()

	// 3. 生成分块上传的初始化信息
	chunkSize := 5 * 1024 * 1024 // 5MB
	chunkcount := int(math.Ceil(float64(filesize) / float64(chunkSize)))
	upInfo := MultipartUploadInfo{
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
	data := map[string]interface{}{
		"filehash":   upInfo.FileHash,
		"filesize":   upInfo.FileSize,
		"chunksize":  upInfo.ChunkSize,
		"chunkcount": upInfo.ChunkCount,
		"username":   username,
	}

	// 使用 map 作为参数调用 HSet 方法
	key := "MP_" + upInfo.UploadID
	err = rConn.HMSet(ctx, key, data).Err()
	if err != nil {
		w.Write(util.NewRespMsg(-1, "write redis failed", err).JSONBytes())
		return
	}
	// 5. 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", upInfo).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	//username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index") // 分块索引
	// 2. 获得redis的一个连接
	rConn := redis.RedisClient()
	// 3. 获得文件分块数据
	// 确保文件目录存在
	fpath := "./data/" + uploadID + "/" + chunkIndex
	err := os.MkdirAll(path.Dir(fpath), 0744)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "mkdir failed", nil).JSONBytes())
		return
	}
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "create file failed", nil).JSONBytes())
		return
	}
	defer fd.Close()

	buf := make([]byte, 1024*1024)
	for {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}

	// TODO 哈希分块校验

	// 4. 更新redis缓存状态
	err = rConn.HSet(context.Background(), "MP_"+uploadID, "chkidx_"+chunkIndex, 1).Err()
	if err != nil {
		w.Write(util.NewRespMsg(-1, "set redis failed:"+err.Error(), nil).JSONBytes())
		return
	}
	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// 通知上传合并
func CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	upid := r.Form.Get("uploadid")
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize, err := strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1, "params invalid", nil).JSONBytes())
		return
	}
	filename := r.Form.Get("filename")
	// 2. 获得redis的一个连接
	rConn := redis.RedisClient()

	// 3. 通过uploadid查询redis并判断是否所有分块上传完成
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := rConn.HGetAll(ctx, "MP_"+upid).Result()
	if err != nil {
		w.Write(util.NewRespMsg(-1, "CompleteUploadHandler: upload failed", nil).JSONBytes())
		return
	}
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
		w.Write(util.NewRespMsg(-2, "CompleteUploadHandler: not all upload", nil).JSONBytes())
		return
	}
	// 4. TODO 如果所有分块上传完成，合并分块

	// 5. 更新唯一文件表及用户文件表
	ok := dblayer.OnFileUploadFinished(filehash, filename, int64(filesize), "")
	if !ok {
		w.Write(util.NewRespMsg(-2, "CompleteUploadHandler: OnFileUploadFinished failed", nil).JSONBytes())
		return
	}
	ok = dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if !ok {
		w.Write(util.NewRespMsg(-3, "CompleteUploadHandler: OnUserFileUploadFinished failed", nil).JSONBytes())
		return
	}
	// 6. 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// TODO 取消上传
func CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数（上传id）

	// 2. 删除本地的分块文件
	// 3. 删除redis缓存的分块信息
	// 4. 更新mysql中用户上传记录（非必须）
	// 5. 返回处理结果到客户端
}

// TODO 查询上传状态（进度）
func QueryUploadStatusHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数（上传id）
	// 2. 查询redis缓存的分块信息（chunkCount 和 chkidx_id个数）并计算
	// 3. 返回处理结果到客户端
}
