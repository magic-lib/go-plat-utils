package redislock

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"time"
)

const (
	DefaultExpireTime = 10 * time.Second // 默认过期时间10s
	DefaultKeyFront   = "{redis-lock}"
)

func getLockerKeyName(key string) string {
	return fmt.Sprintf("%s%s", DefaultKeyFront, key)
}

// RedisPing 测试redis连接
func RedisPing(redisClient *redis.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return nil
}
