package seqid

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-cache/cache"
	"github.com/magic-lib/go-plat-locker/lock"
	"github.com/orcaman/concurrent-map/v2"
	"golang.org/x/exp/constraints"
	"sync"
	"time"
)

// IDGenerator 基于MySQL的分布式ID生成器
type IDGenerator[T constraints.Integer] struct {
	LockFunc     func(ns, key string) lock.Locker // 分布式锁，动态获取，避免锁的颗粒度太大的问题，可以解决从IdStore里公共取数据的问题
	IdStore      cache.CommCache[T]               // 缓存的方法
	Timeout      time.Duration                    // 缓存最长时间
	Namespace    string                           // 命名空间
	BatchSize    int                              // 批量获取ID的数量
	currentIDMap cmap.ConcurrentMap[string, T]    // 当前ID
	maxIDMap     cmap.ConcurrentMap[string, T]    // 已预取的最大ID
	mu           sync.Mutex                       // 本地锁
}

// NewIDGenerator 创建新的ID生成器实例
func NewIDGenerator[T constraints.Integer](idGen *IDGenerator[T]) (*IDGenerator[T], error) {
	if idGen == nil {
		idGen = new(IDGenerator[T])
	}
	if idGen.BatchSize <= 0 {
		idGen.BatchSize = 1
	}
	if idGen.IdStore == nil {
		return nil, fmt.Errorf("ID存储不能为空")
	}
	if idGen.Namespace == "" {
		return nil, fmt.Errorf("命名空间不能为空,避免冲突覆盖")
	}
	if idGen.LockFunc == nil {
		setLocker(idGen)
	}

	if idGen.BatchSize <= 0 {
		idGen.BatchSize = 100
	}
	if idGen.Timeout <= 0 {
		idGen.Timeout = 7 * 24 * time.Hour //默认7天
	}
	idGen.currentIDMap = cmap.New[T]()
	idGen.maxIDMap = cmap.New[T]()

	return idGen, nil
}

// setLocker 设置锁
func setLocker[T constraints.Integer](idGen *IDGenerator[T]) {
	if idGen.LockFunc != nil {
		return
	}

	idGen.LockFunc = func(ns, key string) lock.Locker {
		return lock.NewLocker(fmt.Sprintf("%s/%s", ns, key))
	}
	return
}

// SetStartId 设置起始ID
func (g *IDGenerator[T]) SetStartId(key string, startId T) {
	var zero T
	if startId <= zero {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()

	hasSet := false
	if g.currentIDMap.Has(key) {
		if currId, ok1 := g.currentIDMap.Get(key); ok1 {
			if currId > zero {
				if currId < startId {
					hasSet = true
					g.currentIDMap.Set(key, startId)
				}
			}
		}
	}
	if !hasSet {
		g.currentIDMap.Set(key, startId)
	}

	hasSet = false
	if g.maxIDMap.Has(key) {
		if maxId, ok2 := g.maxIDMap.Get(key); ok2 {
			if startId > maxId {
				hasSet = true
				g.maxIDMap.Set(key, startId)
			}
		}
	}
	if !hasSet {
		g.maxIDMap.Set(key, startId)
	}
	return
}

// Generate 获取下一个ID
func (g *IDGenerator[T]) Generate(key string) (T, error) {
	//本地加锁
	g.mu.Lock()
	defer g.mu.Unlock()
	var zero T

	var currMaxID T
	if g.currentIDMap.Has(key) {
		if currId, ok1 := g.currentIDMap.Get(key); ok1 {
			//fmt.Println("id has key", keyPrefix, currId)
			if currId > zero {
				currMaxID = currId
			}
			if maxId, ok2 := g.maxIDMap.Get(key); ok2 {
				//fmt.Println("id has key maxId", keyPrefix, currId, maxId)
				if currId <= maxId {
					g.currentIDMap.Set(key, currId+1)
					//fmt.Println("cache currId:", currId)
					return currId, nil
				}
				if maxId > 0 {
					currMaxID = maxId + 1
				}
			}
		}
	}

	//需要从缓存中获取
	//fmt.Println("currMaxID:", currMaxID)
	startId, endId, err := g.fetchBatchFromCache(key, currMaxID)
	if err != nil {
		return 0, err
	}
	g.currentIDMap.Set(key, startId+1)
	g.maxIDMap.Set(key, endId)
	//fmt.Println("new currId:", startId, endId)
	return startId, nil
}

// fetchBatchFromDB 从数据库获取一批ID, currId 是当前最大值
func (g *IDGenerator[T]) fetchBatchFromCache(key string, currMaxId T) (T, T, error) {
	//fmt.Println("fetchBatchFromCache:", currMaxId)

	ctx := context.Background()

	// 使用分布式锁
	if g.LockFunc != nil {
		if err := g.LockFunc(g.Namespace, key).Lock(ctx); err != nil {
			return 0, 0, err
		}
		defer func() {
			_ = g.LockFunc(g.Namespace, key).UnLock(ctx)
		}()
	}

	curDataId, err := cache.NsGet[T](ctx, g.IdStore, g.Namespace, key)
	if err != nil {
		return 0, 0, err
	}

	//表名数据与缓存不同步了，则用最大的值为准
	if currMaxId > curDataId {
		curDataId = currMaxId
	} else {
		curDataId = curDataId + 1 //取下一个数
	}
	//fmt.Println("curDataId:", curDataId, currMaxId)

	//表明开始第一个，从1开始
	endDataId := curDataId + T(g.BatchSize) - 1
	_, err = cache.NsSet(ctx, g.IdStore, g.Namespace, key, endDataId, g.Timeout)
	if err != nil {
		return 0, 0, err
	}
	return curDataId, endDataId, nil
}
