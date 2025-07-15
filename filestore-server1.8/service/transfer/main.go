package main

import (
	"bufio"
	"encoding/json"
	"filestore-server/config"
	dblayer "filestore-server/db"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"fmt"
	"log"
	"os"
)

// 转移服务程序入口：
// 1. 监听转移任务消息队列
// 2. 消费任务

// 定义 callback 函数：处理消息
func ProcessTransfer(msg []byte) bool {
	// 1. 解析msg
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg, &pubData)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 2. 根据临时存储文件路径，创建文件句柄
	fd, err := os.Open(pubData.CurLocation)
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 3. 通过文件句柄将文件内容读出来并且上传到oss

	err = oss.Bucket().PutObject(
		pubData.DestLocation,
		bufio.NewReader(fd))
	if err != nil {
		log.Println(err.Error())
		return false
	}

	// 4. 更新文件的存储路径到文件表
	suc := dblayer.UpdateFileLocation(pubData.FileHash, pubData.DestLocation)
	if !suc {
		log.Println("ProcessTransfer: 更新文件的存储路径到文件表失败")
		return false
	}
	fmt.Println("文件异步转移成功")
	return true
}

func main() {
	log.Println("开始监听转移任务队列...")
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer-oss",
		ProcessTransfer,
	)
}
