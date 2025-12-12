package goroutines

import (
	"fmt"
	"time"
)

// GoAsyncWithTimeout 执行一个方法带过期时间
func GoAsyncWithTimeout[T any](timeout time.Duration, fun func(params ...any) (T, error), params ...any) (t T, e error) {
	result := make(chan T)
	err := make(chan error)

	// 启动一个 goroutine 来执行耗时操作
	GoAsync(func(params ...any) {
		oneRet, oneErr := fun(params...)
		result <- oneRet
		err <- oneErr
	}, params...)

	// 使用 select 语句来等待结果或超时
	select {
	case res := <-result:
		return res, <-err
	case <-time.After(timeout):
		return t, fmt.Errorf("timeout")
	}
}
