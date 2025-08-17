package cache

import (
	"context"
	"filestore-server/cache/redis"
	"testing"
	"time"
)

func TestRedis(t *testing.T) {
	// 连接Redis
	client := redis.RedisClient()
	// 执行命令
	if err := client.Ping(context.Background()).Err(); err != nil {
		panic(err)
	}
	t.Log("Redis连接成功")
}

func TestHset(t *testing.T) {
	client := redis.RedisClient()
	data := map[string]interface{}{
		"name": "test",
		"age":  18,
	}
	if err := client.HMSet(context.Background(), "test", data).Err(); err != nil {
		panic(err)
	}
	t.Log("map参数，HSet成功")
	time.Sleep(10 * time.Second)
	// 删除key
	if err := client.Del(context.Background(), "test").Err(); err != nil {
		panic(err)
	}
	t.Log("删除成功")
}
