package goroutines

import (
	"github.com/hashicorp/go-multierror"
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
	waitGroupTemp := newWaitGroup(timeout)
	waitGroupTemp.add(len(dataList))

	var breakDataListFlag int64 = 0
	var errTotal *multierror.Error
	var mu sync.Mutex // 用于保护 errTotal

	pageSize := 50 //如果dataList太长，这样会并发很多也不合理，所以分为二维数组会更合理一些，每50一组
	for i := 0; i < len(dataList); {
		limit := i + pageSize
		if limit > len(dataList) {
			limit = len(dataList)
		}

		var aw sync.WaitGroup
		aw.Add(limit - i)

		for j := i; j < limit; j++ {
			i++
			if atomic.LoadInt64(&breakDataListFlag) > 0 {
				waitGroupTemp.done()
				aw.Done()
				continue
			}

			index := j
			value := dataList[index]
			GoAsync(func(_ ...any) {
				if atomic.LoadInt64(&breakDataListFlag) > 0 {
					waitGroupTemp.done()
					aw.Done()
					return
				}

				breakFlag, err := callback(value, index)
				if err != nil {
					mu.Lock()
					errTotal = multierror.Append(errTotal, err)
					mu.Unlock()
				}
				if breakFlag {
					atomic.AddInt64(&breakDataListFlag, 1)
				}
				waitGroupTemp.done()
				aw.Done()
			})
		}
		aw.Wait()
	}

	err := waitGroupTemp.wait()
	if err != nil {
		return false, err
	}
	if errTotal != nil {
		return true, errTotal
	}
	return true, nil
}
