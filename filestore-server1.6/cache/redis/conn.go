package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

var (
	client    *redis.Client
	redisHost = "127.0.0.1:6379"
	redisPass = "123456"
)

func newRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:            redisHost,
		Password:        redisPass,
		DB:              0,
		PoolSize:        50,                // 连接池大小
		MaxIdleConns:    50,                // 最大空闲连接数：避免频繁创建连接 或者 创建连接过多，消耗资源
		MinIdleConns:    10,                // 最小空闲连接数:应对突发请求
		MaxActiveConns:  30,                // 最大连接数
		ConnMaxIdleTime: time.Second * 300, // 连接最大空闲时间
	})
}

func init() {
	client = newRedisClient()
	//client.AddHook(heartBeatHook{})
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
}

func RedisClient() *redis.Client {
	return client
}

//type heartBeatHook struct{}
//
//// 这个实际上并没有编写逻辑作用，用来实现Hook接口的占位方法
//func (heartBeatHook) DialHook(next redis.DialHook) redis.DialHook {
//	return func(ctx context.Context, network, addr string) (net.Conn, error) {
//		// before dial...
//		// 记录日志等前置逻辑
//		// 执行下一个钩子函数
//		conn, err := next(ctx, network, addr)
//		if err != nil {
//			return nil, err
//		}
//		// 执行自定义的逻辑
//		// after dial...
//		return conn, nil
//	}
//}
//
//func (heartBeatHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
//	return func(ctx context.Context, cmd redis.Cmder) error {
//		// before process...
//		if cmd.Name() == "PING" {
//			return next(ctx, cmd)
//		}
//		// 执行命令前检查连接是否存活
//		if err := client.Ping(ctx).Err(); err != nil {
//			// 连接已断开，重新连接
//			client = newRedisClient()
//		}
//		// 执行下一个钩子函数
//		err := next(ctx, cmd)
//		if err != nil {
//			return err
//		}
//		// 执行自定义的逻辑
//		// after process...
//		return nil
//	}
//}
//
//func (heartBeatHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
//	return func(ctx context.Context, cmds []redis.Cmder) error {
//		// before process pipeline...
//		// 执行命令前检查连接是否存活
//		if err := client.Ping(ctx).Err(); err != nil {
//			// 连接已断开，重新连接
//			client = newRedisClient()
//		}
//		// 执行下一个钩子函数
//		err := next(ctx, cmds)
//		if err != nil {
//			return err
//		}
//		// 执行自定义的逻辑
//		// after process pipeline...
//		return nil
//	}
//}
