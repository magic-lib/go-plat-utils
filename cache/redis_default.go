package cache

import (
	startupCfg "github.com/magic-lib/go-plat-startupcfg/startupcfg"
)

var (
	defaultRedisCfg *startupCfg.RedisConfig
)

// SetDefaultRedisConfig 切换默认的redis连接
func SetDefaultRedisConfig(con *startupCfg.RedisConfig) {
	if con != nil {
		defaultRedisCfg = con
	}
}
