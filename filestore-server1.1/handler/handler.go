package handler

import (
	"filestore-server/meta"
	"filestore-server/util"
	"fmt"
	"io"
	"net/http"
	"os"
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

		// 计算文件的哈希值
		newFile.Seek(0, 0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		meta.UpdateFileMetaDB(fileMeta)

		// 提示成功信息
		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// 上传成功的处理函数
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "upload success!")
}
