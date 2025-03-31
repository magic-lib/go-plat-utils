package cache

import (
	"context"
	"time"
)

type defaultCache struct {
	cCache    CommCache[string]
	isDefault bool //是否是默认的，避免重复提交
}

var (
	defaultMemCache CommCache[string] //本地默认缓存
)

// New 新建
func New(con ...CommCache[string]) *defaultCache {
	com := new(defaultCache)
	if len(con) > 0 {
		com.isDefault = false
		com.cCache = con[0]
		return com
	}
	com.isDefault = true
	defaultMemCache = NewMemGoCache[string](5*time.Minute, 10*time.Minute)
	com.cCache = defaultMemCache
	return com
}

// Get 从缓存中取得一个值，如果没有redis则从本地缓存
func (co *defaultCache) Get(ctx context.Context, key string) (string, error) {
	ret, err := co.cCache.Get(ctx, key)
	if err == nil {
		return ret, nil
	}
	if co.isDefault {
		return "", err
	}
	ret2, err2 := defaultMemCache.Get(ctx, key)
	if ret2 == "" || err2 != nil {
		return "", err
	}
	return ret2, nil
}

// Set timeout为秒
func (co *defaultCache) Set(ctx context.Context, key, val string, timeout time.Duration) (bool, error) {
	ret, err := co.cCache.Set(ctx, key, val, timeout)
	if err == nil {
		return ret, nil
	}
	if co.isDefault {
		return false, err
	}
	ret2, err2 := defaultMemCache.Set(ctx, key, val, timeout)
	if !ret2 || err2 != nil {
		return false, err
	}
	return true, nil
}

// Del 从缓存中删除一个key
func (co *defaultCache) Del(ctx context.Context, key string) (bool, error) {
	ret, err := co.cCache.Del(ctx, key)
	if err == nil {
		return ret, nil
	}
	if co.isDefault {
		return false, err
	}
	ret2, err2 := defaultMemCache.Del(ctx, key)
	if !ret2 || err2 != nil {
		return false, err
	}
	return true, nil
}
