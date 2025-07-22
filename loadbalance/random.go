package loadbalance

import (
	"math/rand"
)

// RandomSelector 随机选择器
type RandomSelector[T any] struct {
	instances []T
}

func NewRandomSelector[T any](ins []T) *RandomSelector[T] {
	return &RandomSelector[T]{
		instances: ins,
	}
}

// Select 随机选择实例
func (r *RandomSelector[T]) Select() T {
	if len(r.instances) == 0 {
		var zero T
		return zero
	}
	index := rand.Intn(len(r.instances))
	return r.instances[index]
}
