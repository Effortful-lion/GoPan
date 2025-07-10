package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
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
		// 创建新文件接收上传的文件流
		newFile, err := os.Create("./tmp/" + head.Filename)
		if err != nil {
			fmt.Printf("Failed to create file: %v\n", err.Error())
			return
		}
		defer newFile.Close()
		_, err = io.Copy(newFile, file)
		if err != nil {
			fmt.Printf("Failed to save data: %v\n", err.Error())
			return
		}
		// 提示成功信息
		http.Redirect(w, r, "/file/upload/suc", http.StatusFound)
	}
}

// 上传成功的处理函数
func UploadSucHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "upload success!")
}
