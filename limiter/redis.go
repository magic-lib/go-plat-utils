package limiter

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type RedisLimiter struct {
	client *redis.Client
	script *redis.Script
}

var rateLimiterLua = `
-- ratelimiter.lua
local key = KEYS[1]          -- 限流key
local now = tonumber(ARGV[1]) -- 当前时间戳
local window = tonumber(ARGV[2]) * 1000 -- 窗口时间(转毫秒)
local limit = tonumber(ARGV[3])         -- 阈值

-- 移除窗口外的数据
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- 获取当前请求数
local current = redis.call('ZCARD', key)

if current < limit then
    redis.call('ZADD', key, now, now .. '-' .. math.random())
    redis.call('PEXPIRE', key, window)
    return 1 -- 允许访问
end
return 0 -- 拒绝访问
`

func NewRedisLimiter(client *redis.Client) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		script: redis.NewScript(rateLimiterLua),
	}
}

func (rl *RedisLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now().UnixMilli()
	result, err := rl.script.Run(ctx, rl.client, []string{key},
		now,
		window.Milliseconds(),
		limit,
	).Int()

	if err != nil {
		return false, err
	}
	return result == 1, nil
}
