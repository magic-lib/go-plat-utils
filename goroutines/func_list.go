package goroutines

import (
	"sync"
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
	callback := func(key int, value func() (bool, error)) (bool, error) {
		return value()
	}
	return AsyncExecuteDataList(timeout, dataList, callback)
}

// AsyncExecuteDataList 异步执行数据列表，返回是否全部执行
// return: bool 是否完成循环。  error  执行过程中是否有错误
func AsyncExecuteDataList[T any](timeout time.Duration, dataList []T,
	callback func(key int, value T) (breakFlag bool, err error)) (complete bool, errExec error) {
	if dataList == nil || len(dataList) == 0 {
		return true, nil
	}
	waitGroupTemp := newWaitGroup(timeout)
	waitGroupTemp.add(len(dataList))

	breakDataList := false
	var errTotal error

	//如果dataList太长，这样会并发很多也不合理，所以分为二维数组会更合理一些，每50一组
	pageSize := 50
outLoop:
	for i := 0; i < len(dataList); {
		aw := sync.WaitGroup{}
		if len(dataList)-i > pageSize {
			aw.Add(pageSize)
		} else {
			aw.Add(len(dataList) - i)
		}
		for j := 0; j < pageSize; j++ {
			if i >= len(dataList) {
				break outLoop
			}
			if breakDataList {
				//如果循环中有跳出的指令以后，则后续的循环都直接全部完成
				waitGroupTemp.done()
				i++
				aw.Done()
				continue
			}
			GoAsync(func(params ...interface{}) {
				oneIndexTemp, ok0 := params[0].(int)
				oneValTemp, ok1 := params[1].(T)
				if ok0 && ok1 {
					if breakDataList {
						//如果循环中有跳出的指令以后，则后续的循环都直接全部完成
						waitGroupTemp.done()
						aw.Done()
						return
					}

					breakFlag, err := callback(oneIndexTemp, oneValTemp)
					if err != nil {
						errTotal = err
					}
					if breakFlag {
						//表示需要跳出后续循环
						breakDataList = true
					}
					//如果完成了，才能关闭
					waitGroupTemp.done()
				}
				aw.Done() //无论是否真正完成，都标记完成
			}, i, dataList[i])

			i++
		}
		aw.Wait()
	}

	err := waitGroupTemp.wait()
	if err != nil {
		//因为超时没有完成
		return false, err
	}
	//完成，但有执行的错误
	if errTotal != nil {
		return true, errTotal
	}
	return true, nil
}
