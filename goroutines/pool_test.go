package goroutines_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/goroutines"
)

// 开启、关闭与提交并发执行，配合 -race 检查数据竞争和死锁。
// 关闭池时 ants 不等待已提交的任务，个别任务可能被丢弃，因此这里不校验任务全部完成
func TestRoutinePoolConcurrentOpenClose(t *testing.T) {
	defer goroutines.CloseRoutinePool()
	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 20 {
			case 0:
				goroutines.OpenRoutinePool(2)
			case 1:
				goroutines.CloseRoutinePool()
			}
			goroutines.GoAsync(func(...any) {})
		}(i)
	}
	ch := make(chan struct{})
	go func() { wg.Wait(); close(ch) }()
	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		t.Fatal("并发开启/关闭/提交出现阻塞")
	}
}

// 关闭后可以重新开启
func TestRoutinePoolReopen(t *testing.T) {
	if p := goroutines.OpenRoutinePool(4); p == nil {
		t.Fatal("第一次开启失败")
	}
	goroutines.CloseRoutinePool()
	p := goroutines.OpenRoutinePool(4)
	if p == nil {
		t.Fatal("关闭后再次开启应返回新的池")
	}
	goroutines.CloseRoutinePool()
}

// 并发开启与提交任务：所有任务都要执行
func TestRoutinePoolConcurrent(t *testing.T) {
	defer goroutines.CloseRoutinePool()
	var done atomic.Int32
	var wg sync.WaitGroup
	const tasks = 200
	wg.Add(tasks)
	for i := 0; i < tasks; i++ {
		go func(i int) {
			if i%20 == 0 {
				goroutines.OpenRoutinePool(2)
			}
			goroutines.GoAsync(func(...any) {
				defer wg.Done()
				done.Add(1)
			})
		}(i)
	}
	ch := make(chan struct{})
	go func() { wg.Wait(); close(ch) }()
	select {
	case <-ch:
	case <-time.After(10 * time.Second):
		t.Fatalf("任务未全部完成：%d/%d", done.Load(), tasks)
	}
}
