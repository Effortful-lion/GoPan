// 下载服务 8083/
package main

import (
	"filestore-server/app/download/rpc"
	"filestore-server/app/download/service"
	"filestore-server/app/gateway/middleware"
	"filestore-server/config"
	"filestore-server/discovery"
	"filestore-server/idl/download/downloadPb"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	config.InitConfig()

	// 注册服务
	go registerService()

	rpc.InitRpcClient()
	log.Println("DBProxyClient 初始化成功:", rpc.DBProxyClient)
	select {}
}

func registerService() {
	// 注册服务到 etcd
	registerEtcdService()
	// 注册服务到 grpc 服务器 并 启动监听
	registerGrpcService()
}

// 服务注册信息：服务地址 服务注册函数
type gRPCRegisterConfig struct {
	Addr         string
	RegisterFunc func(g *grpc.Server)
}

// 注册本服务到 grpc 服务器
func registerGrpcService() {
	// 1. 创建 gRPC 服务器
	// 创建带有恢复中间件的 gRPC 服务器
	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.RecoveryInterceptor),
	)
	defer s.Stop()

	// 2. TODO 不同： 向 grpc服务器 执行服务注册
	cfg := gRPCRegisterConfig{
		Addr: config.Config.ServiceConfig.DownloadServiceAddress,
		RegisterFunc: func(g *grpc.Server) {
			downloadPb.RegisterDownloadServiceServer(g, service.NewDownloadService())
		},
	}

	// 3. 执行服务注册函数
	cfg.RegisterFunc(s)

	fmt.Println("tcp端口：", cfg.Addr)

	// 4. 监听端口，监听服务端口地址：服务注册地址（开始监听事件）
	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		log.Println("cannot listen")
	}

	// 5. 启动 gRPC 服务器，开启监听端口的处理程序（开始处理监听到的事件）
	log.Printf("grpc server started as: %s \n", cfg.Addr)
	err = s.Serve(lis)
	if err != nil {
		log.Println("server started error", err)
		return
	}

	// TODO s 服务器资源优雅关闭
}

// 注册服务到 etcd
func registerEtcdService() {
	// 1. 获取 etcd 地址
	etcd_addr := fmt.Sprintf("%s:%d", config.Config.EtcdConfig.EtcdHost, config.Config.EtcdConfig.EtcdPort)
	// 2. 创建 etcd 注册器
	r := discovery.NewRegister([]string{etcd_addr}, logrus.New())
	defer r.Stop()

	// 3. TODO 不同：构造服务节点信息：服务名 + 服务地址
	info := discovery.Server{
		Name: config.Config.Domain.DownloadServiceDomain,
		Addr: config.Config.ServiceConfig.DownloadServiceAddress,
	}
	logrus.Println(info)

	// 4. 注册 服务到 etcd
	_, err := r.Register(info, 2)
	if err != nil {
		logrus.Fatalln(err)
	}
}
