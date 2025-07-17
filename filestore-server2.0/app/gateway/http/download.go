package http

import (
	"context"
	"filestore-server/app/gateway/rpc"
	"filestore-server/idl/download/downloadPb"
	"net/http"
)

func DownloadHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.FormValue("filehash")
	// 调用rpc
	res, err := rpc.DownloadHandler(context.Background(), &downloadPb.DownloadRequest{
		FileHash: filehash,
	})
	if err != nil || res == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filehash+"\"")
	w.Write(res.FileData)
}

func DownloadURLHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	filehash := r.FormValue("filehash")
	// 调用rpc
	res, err := rpc.DownloadURLHandler(context.Background(), &downloadPb.DownloadURLRequest{
		FileHash: filehash,
	})
	if err != nil || res == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(res.SignedUrl))
}
