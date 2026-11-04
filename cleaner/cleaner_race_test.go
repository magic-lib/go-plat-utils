package cleaner_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/magic-lib/go-plat-utils/cleaner"
)

type testRes struct{ stopped *atomic.Int32 }

func (r testRes) Stop()        { r.stopped.Add(1) }
func (r testRes) Name() string { return "test" }

// Run 与并发 Register 无数据竞争，已登记的资源都会被停止
func TestRunConcurrentRegister(t *testing.T) {
	var stopped atomic.Int32
	for i := 0; i < 5; i++ {
		cleaner.Register(testRes{&stopped})
	}
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			cleaner.Register(testRes{&stopped})
		}
	}()
	cancel()
	cleaner.Run(ctx)
	wg.Wait()
	if stopped.Load() < 5 {
		t.Fatalf("已登记的资源应被停止, stopped=%d", stopped.Load())
	}
}
