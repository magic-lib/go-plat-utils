package syncx

import (
	"errors"
	"testing"
	"time"
)

// fn panic 时等待方拿到错误，而不是 (nil, nil)
func TestSingleFlightPanicWaiterGetsError(t *testing.T) {
	g := NewSingleFlight()
	release := make(chan struct{})
	started := make(chan struct{})
	go func() {
		defer func() { _ = recover() }()
		_, _ = g.Do("k", func() (any, error) {
			close(started)
			<-release
			panic("boom")
		})
	}()
	<-started
	errCh := make(chan error, 1)
	go func() {
		_, err := g.Do("k", func() (any, error) { return 1, nil })
		errCh <- err
	}()
	time.Sleep(50 * time.Millisecond) // 让第二个调用进入等待
	close(release)
	if err := <-errCh; !errors.Is(err, errFnPanicked) {
		t.Fatalf("waiter err = %v, want errFnPanicked", err)
	}
}
