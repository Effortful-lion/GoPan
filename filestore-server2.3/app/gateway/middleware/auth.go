package middleware

import (
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// http请求拦截器
//func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		r.ParseForm()
//		username := r.Form.Get("username")
//		token := r.Form.Get("token")
//		// 验证token是否有效
//		if len(username) < 3 || !IsTokenVaild(token) {
//			w.WriteHeader(http.StatusForbidden)
//			return
//		}
//		h(w, r)
//	})
//}

func AuthInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.Request.FormValue("username")
		token := c.Request.FormValue("token")
		// 验证token是否有效
		if len(username) < 3 || !util.IsTokenVaild(token) {
			resp := util.RespMsg{
				Code: -2,
				Msg:  "Invalid token",
				Data: nil,
			}
			c.JSON(http.StatusOK, resp)
			return
		}
		c.Next()
	}
}
