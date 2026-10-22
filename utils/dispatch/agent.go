package dispatch

import (
	"context"
)

// Agent 表示一个分配目标。环形分配以 GetUint64() 返回的 uint64 作为排序与定位依据。
type Agent interface {
	GetAgentId() uint64
}

// AgentWeighter 可选的权重接口。
// 实现了它的 Agent 会按权重参与分配（见 AssignN）；未实现、或返回 0 时按权重 1 处理。
// 用可选接口而不是把 GetWeight 塞进 Agent，是为了让已有的 Agent 实现零成本兼容。
type AgentWeighter interface {
	GetWeight() uint64
}

// AgentWeight 取 Agent 的权重：未实现 AgentWeighter、或权重为 0 时按 1 处理。
func AgentWeight(a Agent) uint64 {
	if a == nil {
		return 0
	}
	weighter, ok := a.(AgentWeighter)
	if !ok {
		return 1
	}
	if weight := weighter.GetWeight(); weight > 0 {
		return weight
	}
	return 1
}

// FetchAgentsFunc 拉取最新列表的函数。
type FetchAgentsFunc func(ctx context.Context) ([]Agent, error)

// SkipAgentsFunc 批量跳过条件：一次性给定当前全部 Agent，返回每个是否跳过（key=Agent.GetAgentId()，value=true 表示该 Agent 不分单）。
// 返回 nil 或为空 map 时视作“都不跳过”；只返回的某些 key 时，缺失的 Agent 视作不跳过，因此
// 不存在“返回切片长度和入参不一致导致下标错位/越界”的问题。
//
// 它在每次分配入口只调用一次，适合回调里需要查库/远程调用的场景：
// 一次查询拿到所有 Agent 的状态，远优于逐 Agent 查询（N 次串行 IO → 1 次批量）。
type SkipAgentsFunc func(ctx context.Context, agents []Agent) map[uint64]bool
