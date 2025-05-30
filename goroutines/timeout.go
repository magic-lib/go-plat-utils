package goroutines

import (
	"fmt"
	"time"
)

// RunWithTimeout 执行一个方法带过期时间
func RunWithTimeout[T any](timeout time.Duration, fun func() (T, error)) (t T, e error) {
	result := make(chan T)
	err := make(chan error)

	// 启动一个 goroutine 来执行耗时操作
	GoAsync(func(params ...interface{}) {
		oneRet, oneErr := fun()
		result <- oneRet
		err <- oneErr
	})

	// 使用 select 语句来等待结果或超时
	select {
	case res := <-result:
		return res, <-err
	case <-time.After(timeout):
		return t, fmt.Errorf("timeout")
	}
}
