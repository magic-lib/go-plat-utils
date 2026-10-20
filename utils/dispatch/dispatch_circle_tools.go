package dispatch

import (
	"context"
	"sort"
)

// weightQuota 用最大余数法把 n 个名额按 slots 分摊，返回每个位置分到的数量，总和恰好为 n。
func weightQuota(slots []uint64, total uint64, n int) []uint64 {
	quota := make([]uint64, len(slots))
	if n <= 0 || total == 0 || len(slots) == 0 {
		return quota
	}

	type remItem struct {
		idx uint64
		rem uint64
	}
	items := make([]remItem, 0, len(slots))
	used := uint64(0)
	for i, slot := range slots {
		prod := uint64(n) * slot
		quota[i] = prod / total
		used += quota[i]
		items = append(items, remItem{idx: uint64(i), rem: prod % total})
	}

	// 取整后剩下的名额，按余数从大到小依次补 1（余数相同则下标小的优先）
	rest := uint64(n) - used
	if rest > 0 {
		sort.SliceStable(items, func(a, b int) bool { return items[a].rem > items[b].rem })
		for k := uint64(0); k < rest; k++ {
			quota[items[int(k)%len(items)].idx]++
		}
	}
	return quota
}

// indexOf 返回列表中第一个等于 id 的实例位置，不存在返回 -1。
func indexOf(agents []Agent, id uint64) int {
	return indexOfNth(agents, id, 0)
}

// equals 统计列表中与给定 id 相等的实例个数。
func occurrenceCount(agents []Agent, id uint64) int {
	c := 0
	for _, a := range agents {
		if a.GetAgentId() == id {
			c++
		}
	}
	return c
}

// indexOfNth 返回列表中第 n 个（0-based）等于 id 的实例的位置，不存在返回 -1。
func indexOfNth(agents []Agent, id uint64, n int) int {
	c := 0
	for i, a := range agents {
		if a.GetAgentId() == id {
			if c == n {
				return i
			}
			c++
		}
	}
	return -1
}

// nextGreaterFirstIndex 返回第一个 id 严格大于给定 id 的实例位置；不存在返回 -1。
func nextGreaterFirstIndex(agents []Agent, id uint64) int {
	for i, a := range agents {
		if a.GetAgentId() > id {
			return i
		}
	}
	return -1
}

// instanceIndexAt 返回位置 pos 处实例在“相同 id 组”内的实例序号（0-based）。
func instanceIndexAt(agents []Agent, pos int) int {
	id := agents[pos].GetAgentId()
	c := 0
	for i := 0; i < pos; i++ {
		if agents[i].GetAgentId() == id {
			c++
		}
	}
	return c
}

// ctxOr 返回实际使用的 ctx：优先入参，其次分配器自身的 ctx，最后 Background。
func ctxOr(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
