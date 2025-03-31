package redislock

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
	"time"
)

// RedSyncLock 	redis锁
type RedSyncLock struct {
	redisClient *redis.Client
	rs          *redsync.Redsync
	mx          *redsync.Mutex
	key         string
	expiration  time.Duration
}

// NewRedSync 新的锁
func NewRedSync(redisClient *redis.Client, key string, expiration time.Duration) (*RedSyncLock, error) {
	if redisClient == nil {
		return nil, fmt.Errorf("redis client is nil")
	}
	err := RedisPing(redisClient)
	if err != nil {
		return nil, err
	}

	rLock := new(RedSyncLock)
	rLock.redisClient = redisClient
	rLock.key = getLockerKeyName(key)
	rLock.expiration = expiration

	// implements the `redis.Pool` interface.
	client := goredislib.NewClient(&goredislib.Options{
		Addr:     redisClient.Options().Addr,
		Username: redisClient.Options().Username,
		Password: redisClient.Options().Password,
	})
	pool := goredis.NewPool(client) // or, pool := redigo.NewPool(...)

	// Create an instance of redisync to be used to obtain a mutual exclusion
	// lock.
	rs := redsync.New(pool)

	rLock.rs = rs

	//redis时间不能太短，避免大量的redis操作
	if expiration < DefaultExpireTime {
		expiration = DefaultExpireTime
	}

	// Obtain a new mutex by using the same name for all instances wanting the
	// same lock.
	mutex := rs.NewMutex(getLockerKeyName(key), redsync.WithExpiry(expiration))

	rLock.mx = mutex

	return rLock, nil
}

// Lock 上锁
func (l *RedSyncLock) Lock(ctx context.Context) (bool, error) {
	if err := l.mx.LockContext(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// UnLock 解锁
func (l *RedSyncLock) UnLock(ctx context.Context) (bool, error) {
	// Release the lock so other processes or threads can obtain a lock.
	if ok, err := l.mx.UnlockContext(ctx); !ok || err != nil {
		return false, err
	}
	return true, nil
}

// TryLock 尝试加锁，非阻塞
func (l *RedSyncLock) TryLock(ctx context.Context) (bool, error) {
	// Release the lock so other processes or threads can obtain a lock.
	if err := l.mx.TryLockContext(ctx); err != nil {
		return false, err
	}
	return true, nil
}
