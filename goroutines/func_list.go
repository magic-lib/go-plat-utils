package goroutines

import (
	"github.com/hashicorp/go-multierror"
	"github.com/samber/lo"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// AsyncExecuteFuncList 异步执行方法，返回是否全部执行
func AsyncExecuteFuncList(timeout time.Duration, calls ...func() (bool, error)) (complete bool, errExec error) {
	if calls == nil || len(calls) == 0 {
		return true, nil
	}
	dataList := make([]func() (bool, error), 0)
	for _, call := range calls {
		dataList = append(dataList, call)
	}
	callback := func(value func() (bool, error), key int) (bool, error) {
		return value()
	}
	return AsyncExecuteDataList(timeout, dataList, callback)
}

// AsyncExecuteDataList 异步执行数据列表，返回是否全部执行
// return: bool 是否完成循环。  error  执行过程中是否有错误
func AsyncExecuteDataList[T any](timeout time.Duration, dataList []T,
	callback func(value T, key int) (breakFlag bool, err error)) (complete bool, errExec error) {
	if dataList == nil || len(dataList) == 0 {
		return true, nil
	}
	return AsyncForEachWhile(dataList, func(value T, key int) (bool, error) {
		goonFlag, err := callback(value, key)
		return !goonFlag, err
	}, AsyncForEachWhileOptions{
		TotalTimeout: timeout,
	})
}

type AsyncForEachWhileOptions struct {
	TotalTimeout   time.Duration //总超时时间
	ChunkSize      int           //每组分配数量
	MaxConcurrency int           //最大并发数
}

func chunkSlice[T any](slice []T, chunkSize int) [][]T {
	var result [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}
	return result
}

// AsyncForEachWhile 异步执行方法，返回是否全部执行
func AsyncForEachWhile[T any](collection []T,
	iteratee func(item T, index int) (goon bool, err error), opt ...AsyncForEachWhileOptions) (complete bool, errExec error) {
	if len(collection) == 0 {
		return true, nil
	}
	option := AsyncForEachWhileOptions{
		TotalTimeout:   time.Minute,
		ChunkSize:      0, //如果dataList太长，这样会并发很多也不合理，所以分为二维数组会更合理一些
		MaxConcurrency: 0, //最大并发数
	}
	if len(opt) > 0 {
		if opt[0].TotalTimeout > 0 {
			option.TotalTimeout = opt[0].TotalTimeout
		}
		if opt[0].ChunkSize > 0 {
			option.ChunkSize = opt[0].ChunkSize
		}
		if opt[0].MaxConcurrency > 0 {
			option.MaxConcurrency = opt[0].MaxConcurrency
		}
	}

	{ //对ChunkSize和 MaxConcurrency 进行合理优化
		if option.ChunkSize == 0 {
			switch {
			case len(collection) > 1000:
				option.ChunkSize = 100
			case len(collection) > 100:
				option.ChunkSize = 50
			default:
				option.ChunkSize = 10
			}
			if len(collection) < option.ChunkSize {
				option.ChunkSize = 1
			}
		}
		if option.MaxConcurrency == 0 {
			option.MaxConcurrency = runtime.NumCPU() * 4
		}
		if option.MaxConcurrency == 0 {
			option.MaxConcurrency = 4
		}

		// 确保并发数不会超过总任务数，避免不必要的等待
		totalChunks := (len(collection) + option.ChunkSize - 1) / option.ChunkSize
		if option.MaxConcurrency > totalChunks {
			option.MaxConcurrency = totalChunks
		}
	}

	waitGroupTemp := newWaitGroup(option.TotalTimeout)
	waitGroupTemp.add(len(collection))

	var breakDataListFlag int64 = 0
	var errTotal *multierror.Error
	var mu sync.Mutex // 用于保护 errTotal

	// 使用带缓冲的 channel 实现最大并发控制
	semaphore := make(chan struct{}, option.MaxConcurrency)
	newCollection := chunkSlice(collection, option.ChunkSize)

	//起一个协程，避免阻塞主程序
	GoAsync(func(_ ...any) {
		lo.ForEach(newCollection, func(item []T, _ int) {
			chunkList := item // 避免闭包捕获问题

			if atomic.LoadInt64(&breakDataListFlag) > 0 {
				//直接退出
				waitGroupTemp.doneN(len(chunkList))
				return
			}

			// 获取信号量（如果已达最大并发，则阻塞）
			semaphore <- struct{}{}

			GoAsync(func(_ ...any) {
				defer func() {
					<-semaphore // 释放信号量
				}()
				lo.ForEach(chunkList, func(value T, index int) {
					defer func() {
						waitGroupTemp.done()
					}()
					if atomic.LoadInt64(&breakDataListFlag) > 0 {
						return
					}

					goonFlag, err := iteratee(value, index)
					if err != nil {
						mu.Lock()
						errTotal = multierror.Append(errTotal, err)
						mu.Unlock()
					}
					if !goonFlag {
						atomic.AddInt64(&breakDataListFlag, 1)
					}
					return
				})
			})
		})
	})

	err := waitGroupTemp.wait()
	if err != nil {
		return false, err
	}
	if errTotal != nil {
		return true, errTotal
	}
	return true, nil
}
