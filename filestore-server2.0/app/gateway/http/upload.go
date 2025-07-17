package http

import (
	"filestore-server/app/gateway/rpc"
	"filestore-server/common"
	"filestore-server/idl/upload/uploadPb"
	"filestore-server/util"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"time"
)

// 上传文件的处理函数
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		data, err := os.ReadFile("./static/view/index.html")
		if err != nil {
			io.WriteString(w, "internel server error")
			return
		}
		io.WriteString(w, string(data))
	} else if r.Method == "POST" {
		// 接收文件并存储到本地目录
		username := r.Form.Get("username")
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data: %v\n", err.Error())
			return
		}
		defer file.Close()

		// 文件信息存储
		fileMeta := common.FileMeta{
			FileName: head.Filename,
			Location: "./tmp/" + head.Filename,
			UploadAt: time.Now().Format("2006-01-02 15:04:05"),
		}

		// 创建新文件接收上传的文件流
		// 如果不存在/tmp目录则创建
		if err := os.MkdirAll("./tmp", 0744); err != nil {
			fmt.Printf("Failed to create dir: %v\n", err.Error())
			return
		}
		newFile, err := os.Create("./tmp/" + head.Filename)
		if err != nil {
			fmt.Printf("Failed to create file: %v\n", err.Error())
			return
		}
		defer newFile.Close()
		fileMeta.FileSize, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data: %v\n", err.Error())
			return
		}

		// TODO ：这里如果文件特别大，同步计算哈希值太慢，后期改造为一个微服务进行异步计算
		// 计算文件的哈希值
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)

		newFile.Seek(0, 0)

		ctx := r.Context()

		req := &uploadPb.UploadRequest{
			Username: username,
			FileMeta: &uploadPb.FileMeta{
				FileName: fileMeta.FileName,
				Location: fileMeta.Location,
				UploadAt: fileMeta.UploadAt,
				FileHash: fileMeta.FileSha1,
				FileSize: fileMeta.FileSize,
			},
		}
		res, err := rpc.UploadHandlerPost(ctx, req)
		if err != nil {
			fmt.Printf("Failed to upload: %v\n", err.Error())
			return
		}
		w.Write([]byte(res.Url))
	}
}

// 上传成功的处理函数
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "upload success!")
}

// 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))
	// 2. 调用rpc服务
	res, err := rpc.TryFastUploadHandler(r.Context(), &uploadPb.TryFastUploadRequest{
		Username: username,
		FileName: filename,
		FileHash: filehash,
		FileSize: int64(filesize),
	})
	if err != nil {
		w.Write(util.NewRespMsg(-1, "TryFastUploadHandler failed: "+err.Error(), nil).JSONBytes())
		return
	}
	w.Write(util.NewRespMsg(0, "OK", res).JSONBytes())
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

	res, err := rpc.InitialMultipartUploadHandler(r.Context(), &uploadPb.InitMultipartUploadRequest{
		Username: username,
		FileHash: filehash,
		FileSize: int64(filesize),
	})
	if err != nil {
		w.Write(util.NewRespMsg(-1, "InitialMultipartUploadHandler failed: "+err.Error(), nil).JSONBytes())
		return
	}

	// 5. 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", res.InitInfo).JSONBytes())
}

// 上传文件分块
func UploadPartHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数
	r.ParseForm()
	//username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index") // 分块索引

	// 3. 获得文件分块数据
	// 确保文件目录存在
	fpath := "./data/" + uploadID + "/" + chunkIndex
	err := os.MkdirAll(path.Dir(fpath), 0744)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "MkdirAll failed: "+err.Error(), nil).JSONBytes())
		return
	}
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1, "Create failed: "+err.Error(), nil).JSONBytes())
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

	// 4. 调用rpc服务
	_, err = rpc.UploadPartHandler(r.Context(), &uploadPb.UploadPartRequest{
		UploadId:   uploadID,
		ChunkIndex: chunkIndex,
	})
	if err != nil {
		w.Write(util.NewRespMsg(-1, "UploadPartHandler failed: "+err.Error(), nil).JSONBytes())
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

	ctx := r.Context()
	rpc.CompleteUploadHandler(ctx, &uploadPb.CompleteUploadRequest{
		Username: username,
		FileHash: filehash,
		FileSize: int64(filesize),
		FileName: filename,
		UploadId: upid,
	})
	// 6. 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// TODO 取消上传
func CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数（上传id）
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")

	// 2. 调用rpc服务
	_, err := rpc.CancelUploadHandler(r.Context(), &uploadPb.CancelUploadRequest{
		UploadId: uploadID,
	})
	if err != nil {
		w.Write(util.NewRespMsg(-1, "CancelUploadHandler failed: "+err.Error(), nil).JSONBytes())
		return
	}

	// 5. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", nil).JSONBytes())
}

// TODO 查询上传状态（进度）
func QueryUploadStatusHandler(w http.ResponseWriter, r *http.Request) {
	// 1. 解析用户请求参数（上传id）
	r.ParseForm()
	uploadID := r.Form.Get("uploadid")

	// 2. 调用rpc服务
	res, err := rpc.QueryUploadStatusHandler(r.Context(), &uploadPb.QueryUploadStatusRequest{
		UploadId: uploadID,
	})
	if err != nil {
		w.Write(util.NewRespMsg(-1, "QueryUploadStatusHandler failed: "+err.Error(), nil).JSONBytes())
		return
	}

	res_data := map[string]interface{}{
		"totalCount": res.TotalCount,
		"chunkCount": res.ChunkCount,
		"progress":   res.Progress,
		"lacking":    res.Lacking,
	}

	// 3. 返回处理结果到客户端
	w.Write(util.NewRespMsg(0, "OK", res_data).JSONBytes())
}
