package route

import (
	"filestore-server/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	r := gin.Default()

	// 静态资源处理(前者是，后者是当前目录下的相对路径)
	r.Static("/static/", "./static")

	// 公共访问接口
	r.GET("/user/signup", handler.SignupHandler)
	r.POST("/user/signup", handler.DoSignupHandler)
	r.GET("/user/signin", handler.SignInHandler)
	r.POST("/user/signin", handler.DoSignInHandler)

	// 验证token 中间件
	r.Use(handler.HTTPInterceptor())

	return r
}
