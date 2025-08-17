package http

import (
	"context"
	"encoding/json"
	"filestore-server/app/gateway/rpc"
	"filestore-server/idl/es/esPb"
	"log"
	"net/http"
	"strconv"
)

type SearchRes struct {
	filename string
	path     string
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	// 获取参数：key（filename）、username、token（已经处理了）
	// 1. 根据 key 和 username 获取可能的文件，返回文件 hash 集合（es操作）
	// 2. 根据文件的 hash 集合获取对应的文件路径（本地、oss云存储）集合（mysql操作）
	// 3. 返回路径集合
	var res []*SearchRes

	// 解析表单数据（包括URL查询参数）
	if err := r.ParseForm(); err != nil {
		log.Println("解析参数失败:", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// 获取URL参数
	key := r.Form.Get("key")           // 文件名关键词
	username := r.Form.Get("username") // 用户名

	// 这里可以添加参数验证逻辑
	if key == "" || username == "" {
		log.Println("缺少必要参数")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	sizeStr := r.Form.Get("size")
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		log.Println(err)
	}
	startStr := r.Form.Get("start")
	start, err := strconv.Atoi(startStr)
	if err != nil {
		log.Println(err)
	}
	index := r.Form.Get("index")

	// conn 获取 hash 集合
	req := &esPb.SearchReq{
		Key:      key,
		Username: username,
		Size:     int64(size),
		Start:    int64(start),
		Index:    index,
	}
	hashList, err := rpc.GetFileHashList(context.Background(), req)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
	}
	// 返回结果
	ress := hashList.GetRes()
	for _, item := range ress {
		temp := &SearchRes{}
		temp.filename = item.Filename
		temp.path = item.Path
		res = append(res, temp)
	}
	w.WriteHeader(http.StatusOK)
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
