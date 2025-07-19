package main

import (
	"filestore-server/app/gateway/router"
	"filestore-server/app/gateway/rpc"
	"filestore-server/config"
	"filestore-server/util"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	log.Println("网关服务启动...")
	config.InitConfig()
	rpc.InitRpcClient()
	rpc.CheckClientHealthy()
	go startListening()
	select {}
}

func startListening() {
	r := router.Router()
	// 配置http服务
	server := &http.Server{
		Addr:           "127.0.0.1:8080",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	// 启动http服务：目前是 8080 端口
	if err := server.ListenAndServe(); err != nil {
		fmt.Println("gateway启动失败, err: ", err)
	}
	go func() {
		// 优雅关闭
		util.GracefullyShutdown(server)
	}()
}
