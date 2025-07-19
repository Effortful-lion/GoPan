package middleware

import (
	"context"
	"google.golang.org/grpc"
	"log"
	"runtime/debug"
)

// RecoveryInterceptor grpc 服务恢复中间件
func RecoveryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Recovered from panic in %s: %v\n%s", info.FullMethod, err, debug.Stack())
		}
	}()

	return handler(ctx, req)
}
