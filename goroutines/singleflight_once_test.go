package goroutines_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/magic-lib/go-plat-utils/goroutines"
)

// Once 成功后应缓存结果不再执行；失败时下次重新执行；并发调用无数据竞争
func TestSingleFlightOnce(t *testing.T) {
	sf := goroutines.NewSingleFlight()
	var calls atomic.Int32
	fn := func() (any, error) {
		calls.Add(1)
		return "v", nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if v, err := sf.Once("k", fn); err != nil || v != "v" {
				t.Errorf("Once = %v, %v", v, err)
			}
		}()
	}
	wg.Wait()
	before := calls.Load()
	_, _ = sf.Once("k", fn)
	if calls.Load() != before {
		t.Fatalf("成功后不应再执行, 调用次数 %d -> %d", before, calls.Load())
	}

	failCalls := 0
	failFn := func() (any, error) { failCalls++; return nil, errors.New("x") }
	_, _ = sf.Once("f", failFn)
	_, _ = sf.Once("f", failFn)
	if failCalls != 2 {
		t.Fatalf("失败时应重新执行, 调用 %d 次", failCalls)
	}
}
