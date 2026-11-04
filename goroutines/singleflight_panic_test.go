package goroutines

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 并发 Once 只执行一次 fn
func TestSingleFlightOnceConcurrentRounds(t *testing.T) {
	for round := 0; round < 50; round++ {
		g := NewSingleFlight()
		var count int32
		var wg sync.WaitGroup
		for i := 0; i < 16; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				v, err := g.Once("k", func() (any, error) {
					atomic.AddInt32(&count, 1)
					return 1, nil
				})
				if err != nil || v != 1 {
					t.Errorf("got %v %v", v, err)
				}
			}()
		}
		wg.Wait()
		if count != 1 {
			t.Fatalf("round %d: fn executed %d times", round, count)
		}
	}
}

// fn panic 时等待方拿到错误，而不是 (nil, nil)
func TestSingleFlightPanicWaiterGetsError(t *testing.T) {
	g := NewSingleFlight()
	release := make(chan struct{})
	started := make(chan struct{})
	go func() {
		defer func() { _ = recover() }()
		_, _, _ = g.Do("k", func() (any, error) {
			close(started)
			<-release
			panic("boom")
		})
	}()
	<-started
	errCh := make(chan error, 1)
	go func() {
		_, _, err := g.Do("k", func() (any, error) { return 1, nil })
		errCh <- err
	}()
	time.Sleep(50 * time.Millisecond) // 让第二个调用进入等待
	close(release)
	if err := <-errCh; !errors.Is(err, errFnPanicked) {
		t.Fatalf("waiter err = %v, want errFnPanicked", err)
	}
}
