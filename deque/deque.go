package deque

import (
	"github.com/gammazero/deque"
	"slices"
	"sync"
)

// SafeDeque 安全队列
type SafeDeque[T any] struct {
	mu sync.RWMutex
	d  deque.Deque[T]
}

func New[T any]() *SafeDeque[T] {
	return &SafeDeque[T]{}
}
func (s *SafeDeque[T]) GetDeque() deque.Deque[T] {
	return s.d
}

func (s *SafeDeque[T]) Items() []T {
	return slices.Collect(s.d.Iter())
}

// 排序
//sorted := slices.Sorted(deque.Iter())
//q.SetBaseCap(1024) // 设置基础容量为 1024
//q.Grow(1000) // 预分配 1000 个元素的空间

func (s *SafeDeque[T]) PushFront(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.d.PushFront(v)
}
func (s *SafeDeque[T]) PushBack(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.d.PushBack(v)
}
func (s *SafeDeque[T]) PopFront() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d.Len() > 0 {
		return s.d.PopFront(), true
	}
	var zero T
	return zero, false
}
func (s *SafeDeque[T]) PopBack() (T, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.d.Len() > 0 {
		return s.d.PopBack(), true
	}
	var zero T
	return zero, false
}
func (s *SafeDeque[T]) Len() int {
	return s.d.Len()
}
func (s *SafeDeque[T]) Front() (T, bool) {
	if s.d.Len() > 0 {
		return s.d.Front(), true
	}
	var zero T
	return zero, false
}
func (s *SafeDeque[T]) Back() (T, bool) {
	if s.d.Len() > 0 {
		return s.d.Back(), true
	}
	var zero T
	return zero, false
}
func (s *SafeDeque[T]) At(i int) (T, bool) {
	if s.d.Len() > 0 {
		return s.d.At(i), true
	}
	var zero T
	return zero, false
}
