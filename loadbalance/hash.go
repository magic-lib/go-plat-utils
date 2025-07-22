package loadbalance

import (
	"fmt"
	"hash/crc32"
	"sort"
)

// ConsistentHashSelector 一致性哈希选择器
type ConsistentHashSelector[T any] struct {
	replicas int          // 虚拟节点数量（解决数据倾斜）
	ring     []uint32     // 哈希环（存储虚拟节点的哈希值）
	hashMap  map[uint32]T // 虚拟节点哈希 -> 真实实例地址
}

// NewConsistentHashSelector 初始化一致性哈希选择器
func NewConsistentHashSelector[T any](instances []T, replicas int, f func(ins T) string) *ConsistentHashSelector[T] {
	ch := &ConsistentHashSelector[T]{
		replicas: replicas,
		hashMap:  make(map[uint32]T),
	}

	// 添加实例到哈希环（每个实例对应多个虚拟节点）
	for _, ins := range instances {
		keys := f(ins)
		for i := 0; i < replicas; i++ {
			// 计算虚拟节点哈希（实例地址+编号作为key）
			oneKey := fmt.Sprintf("%s-%d", keys, i)
			hash := crc32.ChecksumIEEE([]byte(oneKey))
			ch.ring = append(ch.ring, hash)
			ch.hashMap[hash] = ins
		}
	}

	// 排序哈希环
	sort.Slice(ch.ring, func(i, j int) bool {
		return ch.ring[i] < ch.ring[j]
	})

	return ch
}

// Select 根据请求key选择实例
func (c *ConsistentHashSelector[T]) Select(key string) T {
	var zero T
	if len(c.ring) == 0 {
		return zero
	}

	// 计算请求key的哈希
	hash := crc32.ChecksumIEEE([]byte(key))

	// 二分查找哈希环上第一个 >= 该哈希的虚拟节点
	idx := sort.Search(len(c.ring), func(i int) bool {
		return c.ring[i] >= hash
	})

	// 若索引超出范围，选第一个节点（哈希环是环形）
	if idx == len(c.ring) {
		idx = 0
	}

	return c.hashMap[c.ring[idx]]
}
