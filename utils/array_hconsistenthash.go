package utils

import "sync"

const (
	defaultNamespaceVirtualNode = 2 //外层的虚拟节点数量，倍数不要求太多
)

type ConsistentHash struct {
	coarseHash *consistentHash // 粗粒度哈希环
	replicas   int
	fineHash   map[string]*consistentHash // 细粒度哈希环（每个粗粒度节点对应一个细粒度环）
	mu         sync.RWMutex
}

func NewConsistentHash(replicas int) *ConsistentHash {
	return &ConsistentHash{
		coarseHash: newConsistentHash(defaultNamespaceVirtualNode),
		replicas:   replicas,
		fineHash:   make(map[string]*consistentHash),
	}
}

func (h *ConsistentHash) AddNode(node string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.coarseHash.AddNode(node)
	if _, exists := h.fineHash[node]; !exists {
		h.fineHash[node] = newConsistentHash(h.replicas) // 初始化细粒度环
	}
	h.fineHash[node].AddNode(node)
}

func (h *ConsistentHash) GetNode(key string) string {
	coarseNode := h.coarseHash.GetNode(key) // 快速定位粗粒度节点
	if coarseNode == "" {
		return ""
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	fineHash, exists := h.fineHash[coarseNode] // 获取对应的细粒度环
	if !exists || fineHash == nil {
		return ""
	}
	return fineHash.GetNode(key) // 在细粒度环中精确定位
}
