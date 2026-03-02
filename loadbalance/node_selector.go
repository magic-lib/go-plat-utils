package loadbalance

import (
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

//通过一致性hash算法，可以快速从一个string key获取到唯一的一个节点

type HashCodeGenerator func(key string) uint     //将一个string转换为一个整数
type NodeKeyGenerator[N any] func(data N) string //一个节点转换为名称

// NodeConsistentSelector 一致性哈希选择器 N:node
type NodeConsistentSelector[N any] struct {
	mu                sync.RWMutex // 保护 keys 和 hashMap 的读写安全
	hashCodeGenerator HashCodeGenerator
	nodeKeyGenerator  NodeKeyGenerator[N] // 实例 -> 实例名称
	replicas          int                 // 虚拟节点数量（解决数据倾斜）
	rings             []uint              // 哈希环（存储虚拟节点的哈希值）
	hashMap           map[uint]N          // 虚拟节点哈希 -> 真实实例地址
}

func NewNodeConsistentSelector[N any](replicas int, nameGenerator NodeKeyGenerator[N], fnList ...HashCodeGenerator) (*NodeConsistentSelector[N], error) {
	if replicas <= 0 {
		return nil, fmt.Errorf("replicas must be positive")
	}
	if nameGenerator == nil {
		return nil, fmt.Errorf("nameGenerator must be not nil")
	}
	// 默认使用 CRC32，性能优异
	var fn = func(key string) uint {
		return uint(crc32.ChecksumIEEE([]byte(key)))
	}

	if len(fnList) > 0 {
		fn = fnList[0]
	}

	return &NodeConsistentSelector[N]{
		replicas:          replicas,
		hashCodeGenerator: fn,
		nodeKeyGenerator:  nameGenerator,
		rings:             make([]uint, 0),
		hashMap:           make(map[uint]N),
	}, nil
}

// IsEmpty 检查哈希环是否为空
func (m *NodeConsistentSelector[N]) IsEmpty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rings) == 0
}

// AddNode 添加真实节点到哈希环
// 资深建议：真实环境节点名称应包含 IP+端口，保证唯一
func (m *NodeConsistentSelector[N]) AddNode(nodeList ...N) {
	if len(nodeList) == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, node := range nodeList {
		key := m.nodeKeyGenerator(node)
		if key == "" {
			continue
		}
		for i := 0; i < m.replicas; i++ {
			// 计算虚拟节点哈希（实例地址+编号作为key）
			hashCode := m.getVirtualNodeHashCode(node, i)
			// 边界情况：冲突处理（虽然概率极低，但生产需处理）
			if _, exists := m.hashMap[hashCode]; exists {
				// 资深做法：如果冲突，可以使用加盐重哈希
				// 这里简略处理
				continue
			}
			m.rings = append(m.rings, hashCode)
			m.hashMap[hashCode] = node
		}
	}
	sort.Slice(m.rings, func(i, j int) bool {
		return m.rings[i] < m.rings[j]
	})
}

func (m *NodeConsistentSelector[N]) getVirtualNodeHashCode(node N, currIdx int) uint {
	key := m.nodeKeyGenerator(node)
	if key == "" {
		return 0
	}
	oneKey := fmt.Sprintf("%s-%d", key, currIdx)
	hashCode := m.hashCodeGenerator(oneKey)
	return hashCode
}

// GetNode 根据 key 获取最近的真实节点名称
// 改进：增加 error 返回，处理空环情况
func (m *NodeConsistentSelector[N]) GetNode(anyKey string) (t N, err error) {
	if anyKey == "" {
		return t, fmt.Errorf("empty key provided")
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.rings) == 0 {
		return t, fmt.Errorf("rings is empty")
	}

	hash := m.hashCodeGenerator(anyKey)
	// 二分查找第一个匹配的虚拟节点
	idx := sort.Search(len(m.rings), func(i int) bool {
		return m.rings[i] >= hash
	})

	// 若索引超出范围，选第一个节点（哈希环是环形）
	if idx == len(m.rings) {
		idx = 0
	}

	hashCode := m.rings[idx%len(m.rings)]

	// 完善：确保 hashMap 中一定存在该虚拟节点的映射
	if node, ok := m.hashMap[hashCode]; ok {
		return node, nil
	}
	return t, nil
}

// RemoveNode 移除节点（生产环境必备）
func (m *NodeConsistentSelector[N]) RemoveNode(nodeList ...N) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, node := range nodeList {
		key := m.nodeKeyGenerator(node)
		if key == "" {
			continue
		}
		for i := 0; i < m.replicas; i++ {
			hashCode := m.getVirtualNodeHashCode(node, i)
			delete(m.hashMap, hashCode)
		}
	}

	// 重建 rings 切片
	m.rings = m.rings[:0]
	for hashCode := range m.hashMap {
		m.rings = append(m.rings, hashCode)
	}
	sort.Slice(m.rings, func(i, j int) bool {
		return m.rings[i] < m.rings[j]
	})
}

// GetAllNodes 获取所有节点（用于健康检查）
func (m *NodeConsistentSelector[N]) GetAllNodes() []N {
	m.mu.RLock()
	defer m.mu.RUnlock()

	nodeList := make([]N, 0)
	for _, node := range m.hashMap {
		nodeList = append(nodeList, node)
	}
	return nodeList
}
