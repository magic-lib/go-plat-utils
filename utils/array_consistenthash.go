package utils

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

const (
	defaultVirtualNodeMultiplier = 5
	maxVirtualNodeMultiplier     = 100
)

type consistentHash struct {
	hashFunc              func(data []byte) uint32
	virtualNodeMultiplier int               // 虚拟节点倍数，一个真实的节点对应多个虚拟节点
	nodes                 []uint32          // 哈希环
	nodeMap               map[uint32]string // 虚拟节点哈希 -> 真实节点名称
	mu                    sync.RWMutex
}

func newConsistentHash(replicas int) *consistentHash {
	if replicas <= 0 {
		replicas = defaultVirtualNodeMultiplier
	}
	if replicas > maxVirtualNodeMultiplier {
		replicas = maxVirtualNodeMultiplier // 超出上限时使用默认最大值
	}
	return &consistentHash{
		hashFunc:              crc32.ChecksumIEEE,
		virtualNodeMultiplier: replicas,
		nodeMap:               make(map[uint32]string),
	}
}

func (c *consistentHash) IncreaseReplicas(newReplicas int) {
	if newReplicas > maxVirtualNodeMultiplier {
		newReplicas = maxVirtualNodeMultiplier // 超出上限时使用默认最大值
	}
	if newReplicas <= c.virtualNodeMultiplier {
		return // 新值不大于当前值，无需调整
	}
	c.mu.Lock()
	defer c.mu.Unlock()

	// 记录原有节点
	existingNodes := make(map[string]bool)
	for _, node := range c.nodeMap {
		existingNodes[node] = true
	}

	// 清空原有虚拟节点
	c.nodes = []uint32{}
	c.nodeMap = make(map[uint32]string)

	// 更新 replicas 值
	c.virtualNodeMultiplier = newReplicas

	// 重新添加所有节点（包括虚拟节点）
	for node := range existingNodes {
		c.AddNode(node)
	}
}

func (c *consistentHash) AddNode(node string) {
	// 检查节点是否已存在
	for _, existingNode := range c.nodeMap {
		if existingNode == node {
			return // 节点已存在，直接返回
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	for i := 0; i < c.virtualNodeMultiplier; i++ {
		// 创建虚拟节点名称，例如 "node1#0", "node1#1"
		virtualKey := node + "#" + strconv.Itoa(i)
		hash := c.hashFunc([]byte(virtualKey))
		c.nodes = append(c.nodes, hash)
		c.nodeMap[hash] = node // 映射回真实节点
	}

	sort.Slice(c.nodes, func(i, j int) bool {
		return c.nodes[i] < c.nodes[j]
	})
}

func (c *consistentHash) GetNode(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.nodes) == 0 {
		return ""
	}

	hash := c.hashFunc([]byte(key))

	idx := sort.Search(len(c.nodes), func(i int) bool {
		return c.nodes[i] >= hash
	})

	if idx == len(c.nodes) {
		idx = 0
	}

	return c.nodeMap[c.nodes[idx]]
}
