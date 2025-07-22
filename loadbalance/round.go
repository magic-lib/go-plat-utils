package loadbalance

import "sync/atomic"

// RoundSelector 轮询选择器（线程安全）
type RoundSelector[T any] struct {
	instances []T
	index     uint64 // 当前索引（原子操作保证并发安全）
}

func NewRoundSelector[T any](ins []T, inx uint64) *RoundSelector[T] {
	return &RoundSelector[T]{
		instances: ins,
		index:     inx,
	}
}

// Select 轮询选择实例
func (r *RoundSelector[T]) Select() T {
	if len(r.instances) == 0 {
		var zero T
		return zero
	}
	// 原子自增并取模，避免索引越界
	current := atomic.AddUint64(&r.index, 1) - 1
	index := current % uint64(len(r.instances))
	return r.instances[index]
}
