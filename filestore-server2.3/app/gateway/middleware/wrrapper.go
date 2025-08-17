package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"
	"net/http"
	"time"
)

// 创建熔断器实例
var cb = gobreaker.NewCircuitBreaker(gobreaker.Settings{
	Name:    "FileStoreService",
	Timeout: 30 * time.Second, // 熔断器打开持续时间
	ReadyToTrip: func(counts gobreaker.Counts) bool {
		// 当连续失败次数达到 5 次时，熔断器打开
		return counts.ConsecutiveFailures > 5
	},
	OnStateChange: func(name string, from, to gobreaker.State) {
		// 熔断器状态变化时记录日志
		switch to {
		case gobreaker.StateOpen:
			// 熔断器打开
		case gobreaker.StateHalfOpen:
			// 熔断器半开
		case gobreaker.StateClosed:
			// 熔断器关闭
		}
	},
})

// CircuitBreakerMiddleware 熔断降级中间件
func CircuitBreakerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, err := cb.Execute(func() (interface{}, error) {
			// 执行后续的中间件和处理函数
			c.Next()
			// 检查响应状态码，如果状态码 >= 500，认为请求失败
			if c.Writer.Status() >= http.StatusInternalServerError {
				return nil, fmt.Errorf("request failed with status code %d", c.Writer.Status())
			}
			return nil, nil
		})

		if err != nil {
			if cb.State() == gobreaker.StateOpen {
				// 熔断器打开，返回降级响应
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"code": 503,
					"msg":  "服务暂时不可用，请稍后再试",
				})
				c.Abort()
				return
			}
			// 其他错误处理
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  "服务器内部错误",
			})
			c.Abort()
			return
		}
	}
}
