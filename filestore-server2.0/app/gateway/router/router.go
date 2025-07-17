package router

import (
	"filestore-server/app/gateway/http"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	// 创建Gin引擎
	r := gin.Default()

	// 静态资源处理（使用Gin方式）
	r.Static("/static/", "./static")

	// 文件存取接口（转换为Gin风格）
	r.POST("/file/upload", gin.WrapF(http.UploadHandler))
	r.GET("/file/upload/suc", gin.WrapF(http.UploadSucHandler))
	r.GET("/file/meta", gin.WrapF(http.GetFileMetaHandler))
	r.POST("/file/query", gin.WrapF(http.FileQueryHandler))
	r.GET("/file/download", gin.WrapF(http.DownloadHandler))
	r.POST("/file/update", gin.WrapF(http.FileMetaUpdateHandler))
	r.GET("/file/delete", gin.WrapF(http.FileDeleteHandler))
	r.POST("/file/fastupload", gin.WrapF(http.TryFastUploadHandler))

	// TODO ceph

	// oss
	r.GET("/file/downloadurl", gin.WrapF(http.DownloadURLHandler))

	// 分块上传接口
	r.POST("/file/mpupload/init", gin.WrapF(http.InitialMultipartUploasdHandler))
	r.POST("/file/mpupload/uppart", gin.WrapF(http.UploadPartHandler))
	r.POST("/file/mpupload/complete", gin.WrapF(http.CompleteUploadHandler))
	r.POST("/file/mpupload/cancel", gin.WrapF(http.CancelUploadHandler))
	r.GET("/file/mpupload/status", gin.WrapF(http.QueryUploadStatusHandler))

	// 用户相关接口
	r.POST("/user/info", gin.WrapF(http.UserInfoHandler))
	r.POST("/user/signup", gin.WrapF(http.SignupHandler))
	r.POST("/user/signin", gin.WrapF(http.SignInHandler))
	r.GET("/user/signin", gin.WrapF(http.GetSignInHandler))
	r.GET("/user/signup", gin.WrapF(http.GetSignupHandler))

	return r
}
