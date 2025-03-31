package cache

import (
	"context"
	"fmt"
	startupCfg "github.com/magic-lib/go-plat-startupcfg/startupcfg"
	"github.com/magic-lib/go-plat-utils/conv"
	"time"
)

type redisCache struct {
	redisCfg *startupCfg.RedisConfig //redis配置
	rc       *redisClient
}

func getRedisConfig(redisCfg ...*startupCfg.RedisConfig) *startupCfg.RedisConfig {
	if len(redisCfg) > 0 {
		for _, oneCfg := range redisCfg {
			if oneCfg != nil {
				cli, err := getRedisClient(getContext(nil), oneCfg)
				if cli != nil && err == nil {
					return oneCfg
				}
			}
		}
	}
	return defaultRedisCfg
}

// NewRedisCache 新建
func NewRedisCache(redisCfg ...*startupCfg.RedisConfig) (*redisCache, error) {
	oneCfg := getRedisConfig(redisCfg...)
	if oneCfg != nil {
		return &redisCache{
			redisCfg: oneCfg,
			rc:       NewRedisClient(oneCfg),
		}, nil
	}
	return nil, fmt.Errorf("redis NewRedisCache empty")
}

// Get 从缓存中取得一个值
func (co *redisCache) Get(ctx context.Context, key string) (string, error) {
	return co.rc.Get(getContext(ctx), key)
}

// Set timeout为秒
func (co *redisCache) Set(ctx context.Context, key string, val string, timeout time.Duration) (bool, error) {
	return co.rc.Set(getContext(ctx), key, conv.String(val), timeout)
}

// Del 从缓存中删除一个key
func (co *redisCache) Del(ctx context.Context, key string) (bool, error) {
	return co.rc.Del(getContext(ctx), key)
}
