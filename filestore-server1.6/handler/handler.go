package handler

import (
	"encoding/json"
	dblayer "filestore-server/db"
	"filestore-server/meta"
	"filestore-server/store/oss"
	"filestore-server/util"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

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
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("Failed to get data: %v\n", err.Error())
			return
		}
		defer file.Close()

		// 文件信息存储
		fileMeta := meta.FileMeta{
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
		//meta.UpdateFileMeta(fileMeta)

		newFile.Seek(0, 0)

		// TODO 同时将文件写入 ceph 存储

		// TODO 同时将文件写入 oss 存储
		ossPath := "oss/" + fileMeta.FileSha1
		err = oss.Bucket().PutObject(ossPath, newFile)
		if err != nil {
			fmt.Println("Failed to upload file: " + err.Error())
			w.Write([]byte("Upload Failed:" + err.Error()))
			return
		}
		fileMeta.Location = ossPath

		_ = meta.UpdateFileMetaDB(fileMeta)

		// TODO 更新用户文件表
		r.ParseForm()
		username := r.Form.Get("username")
		suc := dblayer.OnUserFileUploadFinished(username, fileMeta.FileSha1, fileMeta.FileName, fileMeta.FileSize)
		if suc {
			http.Redirect(w, r, "/static/view/home.html", http.StatusFound)
		} else {
			w.Write([]byte("Upload Failed"))
		}
		return
		// TODO 后面可以直接注释掉
		// 提示成功信息
		//http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// 上传成功的处理函数
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "upload success!")
}

// 通过哈希值获取文件元信息的处理函数
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 下载文件
func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	fm := meta.GetFileMeta(fsha1)
	location := fm.Location
	// 打开文件
	file, err := os.Open(location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// 获取文件句柄后读取文件内容到内存
	// 如果读取文件很大，需要使用流的形式，读取部分后响应到客户端，然后刷新缓存再读取下一部分，直到文件读取完毕
	// 这里假设文件很小，直接全部读取到内存
	data, err := io.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 这里需要修改头信息，告诉浏览器是文件下载操作
	w.Header().Set("Content-Type", "application/octect-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+fm.FileName)
	w.Write(data)
}

// 更新元信息（这里只是重命名）
func FileMetaUpdateHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	opType := r.Form.Get("op")
	fileSha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")

	if opType != "0" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	curFileMeta := meta.GetFileMeta(fileSha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)

	data, err := json.Marshal(curFileMeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// 文件删除接口
func FileDeleteHandler(w http.ResponseWriter, r *http.Request) {
	// 删除内存和物理文件
	r.ParseForm()
	fileSha1 := r.Form.Get("filehash")
	// 删除物理文件
	fMeta := meta.GetFileMeta(fileSha1)
	os.Remove(fMeta.Location)
	// 删除内存文件
	meta.RemoveFileMeta(fileSha1)
	w.WriteHeader(http.StatusOK)
}

// 批量获取文件信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	//fileMetas, _ := meta.GetLastFileMetas(int(limitCnt))
	userfiles, err := dblayer.QueryUserFileMetas(username, limitCnt)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userfiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

// 尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	// 1. 解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))
	// 2. 从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// 3. 查不到记录则返回秒传失败
	if fileMeta == nil {
		resp := util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}
	// 4. 上传过则将文件信息写入用户文件表，返回成功
	suc := dblayer.OnUserFileUploadFinished(username, filehash, filename, int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	} else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，请稍后重试",
		}
		w.Write(resp.JSONBytes())
		return
	}
}

// 生成 oss 文件的下载地址
func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	// 从文件表查找记录
	fmeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// TODO 判断文件存在 oss 还是 ceph

	// 服务端签名后返回oss的下载地址
	signedURL := oss.DownloadURL(fmeta.Location)
	w.Write([]byte(signedURL))
}
