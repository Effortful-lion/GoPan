package http

import (
	"context"
	"encoding/json"
	"filestore-server/app/gateway/rpc"
	"filestore-server/idl/dbproxy/dbproxyPb"
	"net/http"
	"strconv"
)

// 通过哈希值获取文件元信息的处理函数
func GetFileMetaHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.Form["filehash"][0]
	data, err := rpc.GetFileMetaHandler(context.Background(), &dbproxyPb.GetFileMetaRequest{
		FileSha1: filehash,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	filemeta := data.GetFileMeta()
	res, err := json.Marshal(filemeta)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(res)
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

	res, err := rpc.GetFileMetaHandler(context.Background(), &dbproxyPb.GetFileMetaRequest{
		FileSha1: fileSha1,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	filemeta := res.GetFileMeta()
	filemeta.FileName = newFileName
	res2, err := rpc.FileMetaUpdateHandler(context.Background(), &dbproxyPb.FileMetaUpdateRequest{
		FileSha1: fileSha1,
		FileName: newFileName,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	filemeta2 := res2.GetFileMeta()

	data, err := json.Marshal(filemeta2)
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

	_, err := rpc.FileDeleteHandler(context.Background(), &dbproxyPb.FileDeleteRequest{
		FileSha1: fileSha1,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// 批量获取文件信息
func FileQueryHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	limitCnt, _ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	res, err := rpc.FileQueryHandler(context.Background(), &dbproxyPb.FileQueryRequest{
		UserName: username,
		Limit:    int32(limitCnt),
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(res.UserFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
