// 文件转移服务 8084
package main

import (
	"bufio"
	"encoding/json"
	dblayer "filestore-server/app/transfer/db"
	mq "filestore-server/app/transfer/mq"
	"filestore-server/config"
	"filestore-server/store/oss"
	"fmt"
	"log"
	"os"
)

func main() {
	log.Println("开始监听转移任务队列...")
	config.InitConfig()
	mq.InitMq()
	//// 注册服务
	//registerService()

	mq.StartConsume(
		config.Config.RabbitMQConfig.TransOSSQueueName,
		"transfer-oss",
		ProcessTransfer,
	)
}

//func registerService() {
//	// 注册服务到 etcd
//	registerEtcdService()
//	// 注册服务到 grpc 服务器 并 启动监听
//	registerGrpcService()
//}
//
//// 服务注册信息：服务地址 服务注册函数
//type gRPCRegisterConfig struct {
//	Addr         string
//	RegisterFunc func(g *grpc.Server)
//}
//
//// 注册本服务到 grpc 服务器
//func registerGrpcService() {
//	// 1. 创建 gRPC 服务器
//	s := grpc.NewServer()
//	defer s.Stop()
//
//	// 2. TODO 不同： 向 grpc服务器 执行服务注册
//	cfg := gRPCRegisterConfig{
//		Addr: config.Config.ServiceConfig.TransferServiceAddress,
//		RegisterFunc: func(g *grpc.Server) {
//			transferPb.
//		},
//	}
//
//	// 3. 执行服务注册函数
//	cfg.RegisterFunc(s)
//
//	fmt.Println("tcp端口：", cfg.Addr)
//
//	// 4. 监听端口，监听服务端口地址：服务注册地址（开始监听事件）
//	lis, err := net.Listen("tcp", cfg.Addr)
//	if err != nil {
//		log.Println("cannot listen")
//	}
//
//	// 5. 启动 gRPC 服务器，开启监听端口的处理程序（开始处理监听到的事件）
//	log.Printf("grpc server started as: %s \n", cfg.Addr)
//	err = s.Serve(lis)
//	if err != nil {
//		log.Println("server started error", err)
//		return
//	}
//
//	// TODO s 服务器资源优雅关闭
//}
//
//// 注册服务到 etcd
//func registerEtcdService() {
//	// 1. 获取 etcd 地址
//	etcd_addr := fmt.Sprintf("%s:%d", config.Config.EtcdConfig.EtcdHost, config.Config.EtcdConfig.EtcdPort)
//	// 2. 创建 etcd 注册器
//	r := discovery.NewRegister([]string{etcd_addr}, logrus.New())
//	defer r.Stop()
//
//	// 3. TODO 不同：构造服务节点信息：服务名 + 服务地址
//	info := discovery.Server{
//		Name: config.Config.Domain.DownloadServiceDomain,
//		Addr: config.Config.ServiceConfig.DownloadServiceAddress,
//	}
//	logrus.Println(info)
//
//	// 4. 注册 服务到 etcd
//	_, err := r.Register(info, 2)
//	if err != nil {
//		logrus.Fatalln(err)
//	}
//}

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
