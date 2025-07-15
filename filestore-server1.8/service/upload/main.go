// package main
//
// import (
//
//	"filestore-server/handler"
//	"filestore-server/route"
//	"fmt"
//	"github.com/gin-gonic/gin"
//	"log"
//	"net/http"
//
// )
//
//	func main() {
//		log.Println("文件上传服务启动，监听端口：8080...")
//
//		mux := http.NewServeMux()
//
//		r := route.Router()
//
//		//// 静态资源处理
//		//http.Handle("/static/",
//		//	http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
//
//		// 文件存取接口
//		mux.HandleFunc("/file/upload", handler.UploadHandler)
//		mux.HandleFunc("/file/upload/suc", handler.UploadSucHandler)
//		mux.HandleFunc("/file/meta", handler.GetFileMetaHandler)
//		mux.HandleFunc("/file/query", handler.FileQueryHandler)
//		mux.HandleFunc("/file/download", handler.DownloadHandler)
//		mux.HandleFunc("/file/update", handler.FileMetaUpdateHandler)
//		mux.HandleFunc("/file/delete", handler.FileDeleteHandler)
//		mux.HandleFunc("/file/fastupload", handler.TryFastUploadHandler)
//
//		// oss
//		mux.HandleFunc("/file/downloadurl", handler.DownloadURLHandler)
//
//		// 分块上传接口
//		mux.HandleFunc("/file/mpupload/init", handler.InitialMultipartUploasdHandler)
//		mux.HandleFunc("/file/mpupload/uppart", handler.UploadPartHandler)
//		mux.HandleFunc("/file/mpupload/complete", handler.CompleteUploadHandler)
//		mux.HandleFunc("/file/mpupload/cancel", handler.CancelUploadHandler)
//		mux.HandleFunc("/file/mpupload/status", handler.QueryUploadStatusHandler)
//
//		// 用户相关接口
//		//mux.HandleFunc("/user/signup", handler.SignupHandler)
//		//mux.HandleFunc("/user/signin", handler.SignInHandler)
//		mux.HandleFunc("/user/info", handler.UserInfoHandler)
//		//r.GET("/user/info", gin.WrapF(handler.UserInfoHandler))
//
//		mux.Handle("/", r)
//
//		err := http.ListenAndServe(":8080", mux)
//		if err != nil {
//			fmt.Printf("Failed to start server: %v\n", err.Error())
//		} else {
//			log.Printf("Server started on port 8080\n")
//		}
//	}
package main

import (
	"filestore-server/handler"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
)

func main() {
	log.Println("文件上传服务启动，监听端口：8080...")

	// 创建Gin引擎
	r := gin.Default()

	// 静态资源处理（使用Gin方式）
	r.Static("/static/", "./static")

	// 文件存取接口（转换为Gin风格）
	r.POST("/file/upload", gin.WrapF(handler.UploadHandler))
	r.GET("/file/upload/suc", gin.WrapF(handler.UploadSucHandler))
	r.GET("/file/meta", gin.WrapF(handler.GetFileMetaHandler))
	r.POST("/file/query", gin.WrapF(handler.FileQueryHandler))
	r.GET("/file/download", gin.WrapF(handler.DownloadHandler))
	r.POST("/file/update", gin.WrapF(handler.FileMetaUpdateHandler))
	r.GET("/file/delete", gin.WrapF(handler.FileDeleteHandler))
	r.POST("/file/fastupload", gin.WrapF(handler.TryFastUploadHandler))

	// oss
	r.GET("/file/downloadurl", gin.WrapF(handler.DownloadURLHandler))

	// 分块上传接口
	r.POST("/file/mpupload/init", gin.WrapF(handler.InitialMultipartUploasdHandler))
	r.POST("/file/mpupload/uppart", gin.WrapF(handler.UploadPartHandler))
	r.POST("/file/mpupload/complete", gin.WrapF(handler.CompleteUploadHandler))
	r.POST("/file/mpupload/cancel", gin.WrapF(handler.CancelUploadHandler))
	r.GET("/file/mpupload/status", gin.WrapF(handler.QueryUploadStatusHandler))

	// 用户相关接口
	r.POST("/user/info", gin.WrapF(handler.UserInfoHandler))

	// 启动Gin服务器
	if err := r.Run(":8080"); err != nil {
		fmt.Printf("Failed to start server: %v\n", err.Error())
	} else {
		log.Printf("Server started on port 8080\n")
	}
}
