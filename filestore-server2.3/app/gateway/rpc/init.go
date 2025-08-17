package rpc

import (
	"context"
	"filestore-server/config"
	"filestore-server/discovery"
	"filestore-server/idl/dbproxy/dbproxyPb"
	"filestore-server/idl/download/downloadPb"
	"filestore-server/idl/es/esPb"
	"filestore-server/idl/upload/uploadPb"
	"filestore-server/idl/user/userPb"
	"fmt"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"log"
	"time"
)

// 初始化rpc客户端，连接服务端，以便进行rpc调用（对外暴露客户端api）

var (
	Resolver   *discovery.Resolver
	ctx        context.Context
	CancelFunc context.CancelFunc

	// TODO 这里进行 rpc 客户端全局声明
	DBProxyClient  dbproxyPb.DBProxyServiceClient
	DownloadClient downloadPb.DownloadServiceClient
	UploadClient   uploadPb.UploadServiceClient
	UserClient     userPb.UserServiceClient
	ESClient       esPb.ESServiceClient
)

func InitRpcClient() {
	// 注册etcd解析器到grpc
	EtcdAddress := fmt.Sprintf("%s:%d", config.Config.EtcdConfig.EtcdHost, config.Config.EtcdConfig.EtcdPort)
	Resolver = discovery.NewResolver([]string{EtcdAddress}, logrus.New())
	resolver.Register(Resolver)
	ctx, CancelFunc = context.WithTimeout(context.Background(), 3*time.Second)
	defer CancelFunc()

	defer Resolver.Close()
	// TODO 这里进行 rpc客户端 初始化
	initClient(config.Config.Domain.DownloadServiceDomain, &DownloadClient)
	initClient(config.Config.Domain.UserServiceDomain, &UserClient)
	initClient(config.Config.Domain.UploadServiceDomain, &UploadClient)
	initClient(config.Config.Domain.DBProxyServiceDomain, &DBProxyClient)
	initClient(config.Config.Domain.ESSServiceDomain, &ESClient)
}

func CheckClientHealthy() {
	if DBProxyClient == nil {
		panic("DBProxyClient is nil")
	}
	if DownloadClient == nil {
		panic("DownloadClient is nil")
	}
	if UploadClient == nil {
		panic("UploadClient is nil")
	}
	if UserClient == nil {
		panic("UserClient is nil")
	}
	if ESClient == nil {
		panic("ESClient is nil")
	}
	log.Println("CheckClientHealthy success")
	log.Println("DBProxyClient: ", DBProxyClient)
	log.Println("DownloadClient: ", DownloadClient)
	log.Println("UploadClient: ", UploadClient)
	log.Println("UserClient: ", UserClient)
	log.Println("ESClient: ", ESClient)
}

// TODO 初始化客户端
func initClient(serviceName string, client interface{}) {
	conn, err := connectServer(serviceName)

	if err != nil {
		panic(err)
	}

	// TODO 这里添加不同服务端的客户端变量初始化
	switch c := client.(type) {
	case *downloadPb.DownloadServiceClient:
		*c = downloadPb.NewDownloadServiceClient(conn)
	case *userPb.UserServiceClient:
		*c = userPb.NewUserServiceClient(conn)
	case *uploadPb.UploadServiceClient:
		*c = uploadPb.NewUploadServiceClient(conn)
	case *dbproxyPb.DBProxyServiceClient:
		*c = dbproxyPb.NewDBProxyServiceClient(conn)
	case *esPb.ESServiceClient:
		*c = esPb.NewESServiceClient(conn)
	default:
		panic("unsupported client type")
	}
}

// 初始化客户端连接服务端
func connectServer(serviceName string) (conn *grpc.ClientConn, err error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	}
	addr := fmt.Sprintf("%s:///%s", Resolver.Scheme(), serviceName)
	log.Printf("connectServer addr: %s", addr)
	// 调试信息
	log.Printf("Resolving address: %s", addr)

	// TODO 建立 gRPC 连接：使用上下文控制超时  【已弃用】（暂时使用）
	conn, err = grpc.Dial(addr, opts...)
	if err != nil {
		log.Printf("连接 %s 失败: %v", addr, err)
	}

	if err != nil {
		log.Printf("Failed to connect to %s: %v", addr, err)
	}
	return
}
