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

// SkipAgentFunc 跳过条件：返回 true 表示该 Agent 本次不参与分单（如休假、停用、不在班次等）。
// 每次分配时实时调用，因此人员状态变化无需刷新列表即可立即生效。
type SkipAgentFunc func(ctx context.Context, agent Agent) bool
