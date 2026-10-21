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
	// 持锁期间执行发送：Close 同样需获取 mu，因此在本次 select 执行期间无法并发
	// close(sc.ch)，从而彻底避免「向已关闭的 channel 发送导致 panic」。
	// 代价：当 channel 已满且无接收方时调用 Close 会短暂等待当前发送完成（标准权衡）。
	select {
	case sc.ch <- value:
		return true
	case <-sc.closeSig:
		return false
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
