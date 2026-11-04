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
	defer sc.mu.Unlock()
	if sc.closed {
		return false
	}
	// 持锁期间执行发送：Close 关闭 sc.ch 前同样需获取 mu，因此在本次 select 执行期间无法并发
	// close(sc.ch)，从而彻底避免「向已关闭的 channel 发送导致 panic」。
	// channel 已满且无接收方时，Close 会先关闭 closeSig 让这里返回 false 并释放锁。
	select {
	case sc.ch <- value:
		return true
	case <-sc.closeSig:
		return false
	}
}

func (sc *SafeChannel[T]) Close() {
	sc.once.Do(func() {
		// 先关闭 closeSig 唤醒阻塞在 Send 中的发送方；若先加锁，发送方持锁阻塞时 Close 会永久等待
		close(sc.closeSig)

		sc.mu.Lock()
		defer sc.mu.Unlock()
		sc.closed = true
		close(sc.ch)
	})
}

func (sc *SafeChannel[T]) Chan() <-chan T {
	return sc.ch
}
