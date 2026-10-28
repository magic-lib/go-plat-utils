package dispatch

import (
	"context"
)

// indexOf 返回列表中第一个等于 id 的实例位置，不存在返回 -1。
func indexOf(agents []Agent, id uint64) int {
	return indexOfNth(agents, id, 0)
}

// occurrenceCount 统计列表中与给定 id 相等的实例个数。
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

// ctxOr 返回实际使用的 ctx：入参为 nil 时取 Background。
func ctxOr(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
