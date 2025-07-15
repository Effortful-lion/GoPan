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
	"sort"
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
		w.Write(util.NewRespMsg(-1, "UploadPartHandler：create file failed", nil).JSONBytes())
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
	// 就是将每一块分块的内容进行 哈希计算（这里一般使用CRC32校验，快一点）
	// 前端检验结果和后端校验结果做比对

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
	// 步骤：
	// 1. 基于uploadid 在 ./data/uploadid/ 下获取所有文件并根据文件名排序
	// 2. 创建空文件到 ./tmp/ 下
	// 3. 将分块文件写入到空文件中
	// 4. 合并结束后删除 /data 下的分块文件 TODO （redis缓存文件删不删之后再说）
	filesPath := "./data/" + upid + "/"
	// 遍历目录下的所有文件
	files, err := os.ReadDir(filesPath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "read dir failed", nil).JSONBytes())
		return
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})
	tmpPath := "./tmp/" + strings.Split(filename, "/")[len(strings.Split(filename, "/"))-1]
	fmt.Println(tmpPath)
	newFile, err := os.Create(tmpPath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "CompleteUploadHandler：create file failed", nil).JSONBytes())
		return
	}
	defer newFile.Close()
	// 优化文件合并循环（确保文件正确关闭）
	for _, file := range files {
		chunkFile, err := os.Open(filesPath + file.Name())
		if err != nil {
			w.Write(util.NewRespMsg(-1, "open file failed: "+err.Error(), nil).JSONBytes())
			return
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
			w.Write(util.NewRespMsg(-1, "remove file failed", nil).JSONBytes())
			return
		}
	}()

	// TODO 将合并的文件保存到本地文件系统后，将文件信息保存到唯一文件表中
	fileaddr := "./tmp/" + filename

	// 5. 更新唯一文件表及用户文件表
	ok := dblayer.OnFileUploadFinished(filehash, filename, int64(filesize), fileaddr)
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
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	// 2. 强制删除本地的分块文件
	err := os.RemoveAll("./data/" + uploadID)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "remove file failed:"+err.Error(), nil).JSONBytes())
		return
	}
	// 3. 删除redis缓存的分块信息
	key := "MP_" + uploadID
	rConn := redis.RedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = rConn.Del(ctx, key).Err()
	if err != nil {
		w.Write(util.NewRespMsg(-1, "delete redis failed", nil).JSONBytes())
		return
	}
	// TODO 4. 更新mysql中用户上传记录（非必须）

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// TODO 查询上传状态（进度）
func QueryUploadStatusHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数（上传id）
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")
	// 2. 查询redis缓存的分块信息（chunkCount 和 chkidx_id个数）并计算
	rConn := redis.RedisClient()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	data, err := rConn.HGetAll(ctx, "MP_"+uploadID).Result()
	if err != nil {
		w.Write(util.NewRespMsg(-1, "get upload status failed", nil).JSONBytes())
		return
	}
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
				w.Write(util.NewRespMsg(-1, "get chunk count failed", nil).JSONBytes())
				return
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
	// 3. 返回处理结果到客户端
	mapData := map[string]interface{}{
		"totalCount": totalCount,
		"chunkcount": chunkCount,
		"progress":   fmt.Sprintf("%d%%", chunkCount*100/totalCount), // 进度
		"lacking":    res,                                            // 缺少的分块索引列表
	}
	w.Write(util.NewRespMsg(0, "OK", mapData).JSONBytes())
}
