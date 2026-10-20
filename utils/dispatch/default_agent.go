package dispatch

import "hash/fnv"

// DefaultAgent 是 Agent 的默认实现，可直接以数字或字符串构造。
//
// 以字符串构造时（NewAgentFromString），会通过 FNV-64a 哈希稳定地生成一个 uint64，
// 使得“没有明确数字ID”的对象（如工号、姓名、任意字符串）也能参与均匀的环形分配，
// 应用场景更广。相同字符串始终映射到相同的 uint64。
type DefaultAgent struct {
	id     uint64
	weight uint64 // 分配权重，0 等价于 1
}

// NewAgentFromUint64 以数字直接构造（权重默认 1）。
func NewAgentFromUint64(id uint64) *DefaultAgent {
	return &DefaultAgent{id: id}
}

// NewAgentFromString 以字符串构造，自动哈希出稳定的 uint64 ID（权重默认 1）。
func NewAgentFromString(s string) *DefaultAgent {
	return &DefaultAgent{id: hashStringToUint64(s)}
}

// NewAgentFromUint64WithWeight 以数字 + 权重构造。
func NewAgentFromUint64WithWeight(id, weight uint64) *DefaultAgent {
	return &DefaultAgent{id: id, weight: weight}
}

// NewAgentFromStringWithWeight 以字符串 + 权重构造。
func NewAgentFromStringWithWeight(s string, weight uint64) *DefaultAgent {
	return &DefaultAgent{id: hashStringToUint64(s), weight: weight}
}

// GetAgentId 返回底层 uint64（字符串构造时即为其哈希值）。
func (a *DefaultAgent) GetAgentId() uint64 {
	return a.id
}

// GetWeight 返回权重，未设置时为 0（按 1 参与分配）。
func (a *DefaultAgent) GetWeight() uint64 {
	return a.weight
}

// SetWeight 设置权重，返回自身便于链式调用；传 0 表示恢复成默认权重 1。
func (a *DefaultAgent) SetWeight(weight uint64) *DefaultAgent {
	a.weight = weight
	return a
}

// hashStringToUint64 将字符串稳定地哈希为 uint64。
func hashStringToUint64(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}
