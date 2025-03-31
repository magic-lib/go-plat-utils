package lock

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"github.com/magic-lib/go-plat-utils/internal/gmlock"
	"github.com/magic-lib/go-plat-utils/internal/redislock"
	"log"
	"time"
)

var defaultRedisClient *redis.Client
var defaultExpiration = 30 * time.Second

// SetRedisClient 新建redis锁
func SetRedisClient(redisClient *redis.Client) {
	if redisClient != nil {
		if redislock.RedisPing(redisClient) != nil {
			return
		}
		defaultRedisClient = redisClient
	}
}

// Lock 加锁
func Lock(key string, callFunc func(), expiration ...time.Duration) (bool, error) {
	return lockWithLocker(nil, key, callFunc, false, expiration...)
}

// LockOnce 如果锁住了，就不执行Func了
func LockOnce(key string, callFunc func(), expiration ...time.Duration) (bool, error) {
	return lockWithLocker(nil, key, callFunc, true, expiration...)
}

// lockWithLocker 返回锁目前的状态
func lockWithLocker(locker Locker, key string, callFunc func(), runOnce bool, expiration ...time.Duration) (locked bool, err error) {
	if callFunc == nil {
		return false, fmt.Errorf("callFunc is nil")
	}

	useMemLock := false
	if locker == nil {
		//先使用redis
		if defaultRedisClient != nil {
			timeoutExp := defaultExpiration
			if expiration != nil || len(expiration) > 0 {
				timeoutExp = expiration[0]
			}
			locker1, err := redislock.NewRedSync(defaultRedisClient, key, timeoutExp)
			if err != nil {
				log.Println("newRedSync error:", err)
			} else {
				locker = locker1
			}
		}
		if locker == nil {
			useMemLock = true
			locker = gmlock.NewMemLock(key)
		}
	}

	ctx := context.Background()
	if _, ok := locker.(*gmlock.MemLock); ok {
		useMemLock = true
	}

	if useMemLock {
		if runOnce {
			if gmlock.TryLock(key) {
				goroutines.GoSync(func(params ...interface{}) {
					callFunc()
				}, nil)
			}
			return false, nil
		}
		gmlock.LockFunc(key, func() {
			goroutines.GoSync(func(params ...interface{}) {
				callFunc()
			}, nil)
		})
		return false, nil
	}

	if runOnce {
		if ok, err := locker.TryLock(ctx); ok && err == nil {
			defer locker.UnLock(ctx)
			goroutines.GoSync(func(params ...interface{}) {
				callFunc()
			}, nil)
		}
		return false, nil
	}

	ok, err := locker.Lock(ctx)
	if err == nil {
		if !ok { //返回false表示已经被别的线程锁住了，不能执行了
			return true, nil
		}
		defer locker.UnLock(ctx)
		goroutines.GoSync(func(params ...interface{}) {
			callFunc()
		}, nil)
		return false, nil
	}
	//如果有错误，则用内存锁
	locker = gmlock.NewMemLock(key)
	return lockWithLocker(locker, key, callFunc, runOnce, expiration...)
}
