package loadbalance

import "math/rand"

// InstanceWithConn 带连接数的实例结构
type InstanceWithConn[T any] struct {
	Instance   T
	Connection int // 当前连接数
}

// LeastConnections 选择连接数最少的实例
func LeastConnections[T any](instances []InstanceWithConn[T]) T {
	var zero T
	if len(instances) == 0 {
		return zero
	}

	// 找到连接数最少的实例（若有多个，随机选一个）
	minConn := instances[0].Connection
	candidates := []InstanceWithConn[T]{instances[0]}

	for _, ins := range instances[1:] {
		if ins.Connection < minConn {
			minConn = ins.Connection
			candidates = []InstanceWithConn[T]{ins}
		} else if ins.Connection == minConn {
			candidates = append(candidates, ins)
		}
	}

	// 从候选中随机选一个（避免集中在单个实例）
	return candidates[rand.Intn(len(candidates))].Instance
}
