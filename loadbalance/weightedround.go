package loadbalance

import (
	"fmt"
	"math/rand"
)

// WeightedRoundSelector 带权重的实例结构
type WeightedRoundSelector[T any] struct {
	instances []T
	weights   []uint
}

func NewWeightedRoundSelector[T any](ins []T, weights []uint) (*WeightedRoundSelector[T], error) {
	if len(ins) != len(weights) {
		return nil, fmt.Errorf("实例数量和权重数量不一致")
	}

	return &WeightedRoundSelector[T]{
		instances: ins,
		weights:   weights,
	}, nil
}

// Select 轮询选择实例
func (r *WeightedRoundSelector[T]) Select() T {
	var zero T
	if len(r.instances) == 0 {
		return zero
	}

	var totalWeight uint = 0
	for _, ins := range r.weights {
		totalWeight += ins
	}
	if totalWeight == 0 {
		totalWeight = 100
	}

	// 生成随机数，落在哪个权重区间就选哪个实例
	randNum := rand.Intn(int(totalWeight))
	current := 0
	for k, ins := range r.instances {
		current += int(r.weights[k])
		if randNum < current {
			return ins
		}
	}

	return zero
}
