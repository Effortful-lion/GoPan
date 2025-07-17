package router

import (
	handler2 "filestore-server/app/gateway/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	// 创建Gin引擎
	r := gin.Default()

	// 静态资源处理（使用Gin方式）
	r.Static("/static/", "./static")

	// 文件存取接口（转换为Gin风格）
	r.POST("/file/upload", gin.WrapF(handler2.UploadHandler))
	r.GET("/file/upload/suc", gin.WrapF(handler2.UploadSucHandler))
	r.GET("/file/meta", gin.WrapF(handler2.GetFileMetaHandler))
	r.POST("/file/query", gin.WrapF(handler2.FileQueryHandler))
	r.GET("/file/download", gin.WrapF(handler2.DownloadHandler))
	r.POST("/file/update", gin.WrapF(handler2.FileMetaUpdateHandler))
	r.GET("/file/delete", gin.WrapF(handler2.FileDeleteHandler))
	r.POST("/file/fastupload", gin.WrapF(handler2.TryFastUploadHandler))

	// TODO ceph

	// oss
	r.GET("/file/downloadurl", gin.WrapF(handler2.DownloadURLHandler))

	// 分块上传接口
	r.POST("/file/mpupload/init", gin.WrapF(handler2.InitialMultipartUploasdHandler))
	r.POST("/file/mpupload/uppart", gin.WrapF(handler2.UploadPartHandler))
	r.POST("/file/mpupload/complete", gin.WrapF(handler2.CompleteUploadHandler))
	r.POST("/file/mpupload/cancel", gin.WrapF(handler2.CancelUploadHandler))
	r.GET("/file/mpupload/status", gin.WrapF(handler2.QueryUploadStatusHandler))

	// 用户相关接口
	r.POST("/user/info", gin.WrapF(handler2.UserInfoHandler))
	r.POST("/user/signup", gin.WrapF(handler2.SignupHandler))
	r.POST("/user/signin", gin.WrapF(handler2.SignInHandler))

	return r
}
