package goroutines_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/goroutines"
)

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

// 并发开启/关闭/提交任务，配合 -race 检查数据竞争；所有任务都要执行
func TestRoutinePoolConcurrent(t *testing.T) {
	defer goroutines.CloseRoutinePool()
	var done atomic.Int32
	var wg sync.WaitGroup
	const tasks = 200
	wg.Add(tasks)
	for i := 0; i < tasks; i++ {
		go func(i int) {
			switch i % 20 {
			case 0:
				goroutines.OpenRoutinePool(2)
			case 1:
				goroutines.CloseRoutinePool()
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
