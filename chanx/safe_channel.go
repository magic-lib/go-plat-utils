package chanx

import "sync"

// SafeChannel 泛型安全通道
type SafeChannel[T any] struct {
	ch       chan T
	closeSig chan struct{}
	once     sync.Once
	closed   bool
	mu       sync.Mutex
}

func NewSafeChannel[T any](bufSize int) *SafeChannel[T] {
	return &SafeChannel[T]{
		ch:       make(chan T, bufSize),
		closeSig: make(chan struct{}),
	}
}

func (sc *SafeChannel[T]) Send(value T) bool {
	sc.mu.Lock()
	if sc.closed {
		sc.mu.Unlock()
		return false
	}
	sc.mu.Unlock()

	select {
	case <-sc.closeSig:
		return false
	case sc.ch <- value:
		return true
	}
}

func (sc *SafeChannel[T]) Close() {
	sc.once.Do(func() {
		sc.mu.Lock()
		sc.closed = true
		sc.mu.Unlock()

		close(sc.closeSig)
		close(sc.ch)
	})
}

func (sc *SafeChannel[T]) Chan() <-chan T {
	return sc.ch
}
