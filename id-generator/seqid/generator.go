package seqid

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-cache/cache"
	"github.com/magic-lib/go-plat-locker/lock"
	"github.com/orcaman/concurrent-map/v2"
	"sync"
	"time"
)

// IDGenerator 基于MySQL的分布式ID生成器
type IDGenerator struct {
	IdStoreLocker func(ns, key string) lock.Locker  // 分布式锁，动态获取，避免锁的颗粒度太大的问题，可以解决从IdStore里公共取数据的问题
	IdStore       cache.CommCache[int64]            // 缓存的方法
	Timeout       time.Duration                     // 缓存最长时间
	Namespace     string                            // 命名空间
	BatchSize     int                               // 批量获取ID的数量
	currentIDMap  cmap.ConcurrentMap[string, int64] // 当前ID
	maxIDMap      cmap.ConcurrentMap[string, int64] // 已预取的最大ID
	mu            sync.Mutex                        // 本地锁
}

// NewIDGenerator 创建新的ID生成器实例
func NewIDGenerator(idGen *IDGenerator) (*IDGenerator, error) {
	if idGen == nil {
		idGen = new(IDGenerator)
	}
	if idGen.IdStore == nil {
		return nil, fmt.Errorf("ID存储不能为空")
	}
	if idGen.Namespace == "" {
		return nil, fmt.Errorf("命名空间不能为空")
	}
	if idGen.BatchSize <= 0 {
		idGen.BatchSize = 100
	}
	if idGen.Timeout <= 0 {
		idGen.Timeout = 7 * 24 * time.Hour //默认7天
	}
	idGen.currentIDMap = cmap.New[int64]()
	idGen.maxIDMap = cmap.New[int64]()

	return idGen, nil
}

// Generate 获取下一个ID
func (g *IDGenerator) Generate(keyPrefix string) (int64, error) {
	//本地加锁
	g.mu.Lock()
	defer g.mu.Unlock()

	var currID int64 = 0
	if g.currentIDMap.Has(keyPrefix) {
		if currId, ok1 := g.currentIDMap.Get(keyPrefix); ok1 {
			if currId > 0 {
				currID = currId
			}
			if maxId, ok2 := g.maxIDMap.Get(keyPrefix); ok2 {
				if maxId > 0 {
					currID = maxId
				}
				if currId <= maxId {
					g.currentIDMap.Set(keyPrefix, currId+1)
					return currId, nil
				}
			}
		}
	}

	//需要从缓存中获取
	startId, endId, err := g.fetchBatchFromCache(keyPrefix, currID)
	if err != nil {
		return 0, err
	}
	g.currentIDMap.Set(keyPrefix, startId)
	g.maxIDMap.Set(keyPrefix, endId)
	return startId, nil
}

// fetchBatchFromDB 从数据库获取一批ID
func (g *IDGenerator) fetchBatchFromCache(keyPrefix string, currId int64) (int64, int64, error) {
	ctx := context.Background()

	// 使用分布式锁
	if g.IdStoreLocker != nil {
		if err := g.IdStoreLocker(g.Namespace, keyPrefix).Lock(ctx); err != nil {
			return 0, 0, err
		}
		defer func() {
			_ = g.IdStoreLocker(g.Namespace, keyPrefix).UnLock(ctx)
		}()
	}

	curDataId, err := cache.NsGet[int64](ctx, g.IdStore, g.Namespace, keyPrefix)
	if err != nil {
		return 0, 0, err
	}

	//表名数据与缓存不同步了，则用最大的值为准
	if currId > curDataId {
		curDataId = currId
	}

	//表明开始第一个，从0开始
	if curDataId == 0 {
		endDataId := curDataId + int64(g.BatchSize) - 1
		_, err = cache.NsSet(ctx, g.IdStore, g.Namespace, keyPrefix, endDataId, g.Timeout)
		if err != nil {
			return 0, 0, err
		}
		return curDataId, endDataId, nil
	}
	// 表明当前号已经使用过了
	endDataId := curDataId + int64(g.BatchSize)
	_, err = cache.NsSet(ctx, g.IdStore, g.Namespace, keyPrefix, endDataId, g.Timeout)
	if err != nil {
		return 0, 0, err
	}
	return curDataId + 1, endDataId, nil
}
