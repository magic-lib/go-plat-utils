package dispatch

import (
	"context"
	"sync"
	"time"
)

// Dispatcher 环形分配器
type Dispatcher struct {
	mu           sync.RWMutex
	refreshMu    sync.Mutex // 串行化刷新：手动 Refresh 与后台定时刷新互斥，避免并发 fetch / 旧数据覆盖新数据
	agents       []Agent    // 当前参与分配的客服，按 GetUint64() 升序排列
	lastID       uint64     // 上一次已分配的客服 ID
	lastIdx      int        // lastID 内已分配到的实例序号（0-based）
	assigned     bool       // 是否已分配过（区分首次）
	StartID      uint64     // 默认起始客服ID（首次分配从此客服开始）
	Agents       []Agent    // 初始客服列表（也可事后用 SetAgentList 动态设置）
	Fetch        FetchAgentsFunc
	Period       time.Duration // 定时拉取列表
	FetchTimeout time.Duration // 单次 Fetch 的超时（<=0 表示不限制），防止慢查询把刷新循环卡死
	Skip         SkipAgentFunc // 可选：跳过条件，返回 true 表示该客服不分单
}

// NewDispatcher 创建分配器。
//   - startID：默认值，表示环形队列的起始客服（首单分配给它）；
//   - fetch：拉取在线客服列表的回调函数；
//   - period：定时刷新间隔（>0 且 fetch != nil 时启动后台定时拉取）。
func NewDispatcher(ctx context.Context, dConfig *Dispatcher) (*Dispatcher, error) {
	if dConfig == nil {
		dConfig = new(Dispatcher)
	}
	ctx = ctxOr(ctx)
	err := dConfig.init(ctx)

	if dConfig.Period > 0 && dConfig.Fetch != nil {
		dConfig.start(ctx)
	}
	return dConfig, err
}

// SetSkipFunc 设置跳过条件（返回 true 表示该客服不分单），返回自身便于链式调用。
func (d *Dispatcher) SetSkipFunc(f SkipAgentFunc) *Dispatcher {
	d.mu.Lock()
	d.Skip = f
	d.mu.Unlock()
	return d
}

// AgentList 返回当前的客服列表副本（按分配顺序）。
func (d *Dispatcher) AgentList() []Agent {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]Agent, len(d.agents))
	copy(out, d.agents)
	return out
}

// SetLastAssigned 设置“上一次已完成的客服 ID”（用于程序重启后恢复进度）。
// 之后调用 Assign() 会返回该客服的下一个（含同 ID 的后续实例），保证不重复、且从断点继续成环。
func (d *Dispatcher) SetLastAssigned(id uint64) *Dispatcher {
	d.mu.Lock()
	d.lastID = id
	d.lastIdx = 0
	d.assigned = true
	d.mu.Unlock()
	return d
}

// SetAgentList 动态设置（整体替换）参与分配的客服列表，人员变动时调用即可立即生效。
// 列表会按 ID 升序重排，cursor 不变，因此下一次分配仍能正确成环。
func (d *Dispatcher) SetAgentList(agents []Agent) *Dispatcher {
	d.apply(agents)
	return d
}

// nextPos 在不加锁的前提下，依据当前 cursor 计算下一次应分配实例在 agents 中的位置。
// 调用方需自行加锁。
func (d *Dispatcher) nextPos() int {
	agents := d.agents

	if !d.assigned {
		// 首次分配：优先 startID，否则取最小 ID
		idx := indexOf(agents, d.StartID)
		if idx < 0 {
			idx = 0
		}
		return idx
	}

	cnt := occurrenceCount(agents, d.lastID)
	if cnt == 0 {
		// 该 ID 已不存在（全线下），跳到更大的 ID；没有则回退 startID，再兜底最小
		pos := nextGreaterFirstIndex(agents, d.lastID)
		if pos < 0 {
			if s := indexOf(agents, d.StartID); s >= 0 {
				pos = s
			} else {
				pos = 0
			}
		}
		return pos
	}

	if d.lastIdx+1 < cnt {
		// 同一 ID 还有后续实例：在本组内推进
		return indexOfNth(agents, d.lastID, d.lastIdx+1)
	}

	// 本组已轮完：跳到更大的 ID（没有则回绕到最小）
	pos := nextGreaterFirstIndex(agents, d.lastID)
	if pos < 0 {
		pos = 0
	}
	return pos
}

// available 判断某个 Agent 当前是否可分单：命中 Skip（跳过条件）则不可分。
// 回调在持锁状态下调用，因此回调里不要再反向调用 Dispatcher 的加锁方法，否则会死锁。
func (d *Dispatcher) available(ctx context.Context, a Agent) bool {
	if a == nil {
		return false
	}
	if d.Skip != nil && d.Skip(ctx, a) {
		return false
	}
	return true
}

// nextAvailablePos 从 nextPos() 开始沿环形向后找第一个“可分单”的实例位置，
// 全部被跳过或已达限量时返回 -1。调用方需持锁。
func (d *Dispatcher) nextAvailablePos(ctx context.Context) int {
	if len(d.agents) == 0 {
		return -1
	}
	start := d.nextPos()
	for k := 0; k < len(d.agents); k++ {
		pos := (start + k) % len(d.agents)
		if d.available(ctx, d.agents[pos]) {
			return pos
		}
	}
	return -1
}

// nextAssign 在不加锁的前提下，依据当前 cursor 计算下一次应分配的实例，
// 返回 (目标, 新的lastID, 新的lastIdx)。调用方需自行加锁。
func (d *Dispatcher) nextAssign(ctx context.Context) (Agent, uint64, int) {
	pos := d.nextAvailablePos(ctx)
	if pos < 0 {
		return nil, 0, 0
	}
	t := d.agents[pos]
	return t, t.GetAgentId(), instanceIndexAt(d.agents, pos)
}

// Assign 分配一单：返回排序后“同一 ID 的下一个实例，否则大于 cursor 的第一个”客服，
// 并把 cursor 更新为该客服的（ID, 实例序号），使之成环。
// 命中 Skip/Limited 的客服会被跳过；当没有可用客服时返回 nil（cursor 不动，下次可重新尝试）。
func (d *Dispatcher) Assign(ctx context.Context) Agent {
	return d.AssignWithCtx(ctx)
}

// AssignWithCtx 同 Assign，但显式指定传给 Skip/Limited 回调的 ctx；ctx 为 nil 时回退到默认 ctx。
func (d *Dispatcher) AssignWithCtx(ctx context.Context) Agent {
	ctx = ctxOr(ctx)

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.agents) == 0 {
		return nil
	}

	target, newID, newIdx := d.nextAssign(ctx)
	if target == nil {
		return nil
	}
	d.lastID = newID
	d.lastIdx = newIdx
	d.assigned = true
	return target
}

// AssignN 一次性分配 n 个（如 n 个订单），按各 Agent 的权重比例分摊。
//
// 权重来源：Agent 实现 AgentWeighter（DefaultAgent 已实现），未实现或权重为 0 时按 1 处理。
// 配额用“最大余数法”计算：先取 floor(n*权重/总权重)，再把因取整剩下的名额按余数从大到小补发，
// 保证总数恰好是 n，不会因取整丢名额。
// 发放顺序沿用环形顺序（从当前游标之后开始），多轮轮询发放，使结果均匀交错而不是先给一个人发满。
//
// 例：两个 Agent 权重为 3 和 1，AssignN(4) 得到 [A, B, A, A]。
// 分配完成后游标停在最后被分配的实例上，连续多次调用整体仍是按权重分摊的。
//
// ErrNoAvailableAgent 表示当前没有任何可分单的客服（列表为空，或所有人被 Skip/Limited 剔除）。
var ErrNoAvailableAgent = errors.New("dispatch: 没有可分单的客服")

// AssignN 批量分配 n 个：按权重把 n 个名额分摊到各 Agent，返回分配结果切片。
// 命中 Skip/Limited 的客服不参与分摊（其权重被剔除后再按剩余权重重新分摊）。
//
// 返回：
//   - n<=0：返回 (nil, nil)（无需分配）；
//   - 列表为空或所有人都被跳过/限量（一圈下来没人可分）：返回 (nil, ErrNoAvailableAgent)；
//   - 正常：返回长度恰好为 n 的切片与 nil。
func (d *Dispatcher) AssignN(ctx context.Context, n int) ([]Agent, error) {
	return d.AssignNWithCtx(ctx, n)
}

// AssignNWithCtx 同 AssignN，但显式指定传给 Skip/Limited 回调的 ctx。
func (d *Dispatcher) AssignNWithCtx(ctx context.Context, n int) ([]Agent, error) {
	if n <= 0 {
		return nil, nil
	}
	ctx = ctxOr(ctx)

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.agents) == 0 {
		return nil, ErrNoAvailableAgent
	}

	slots, _ := d.weightSlots()
	// 被跳过/已达限量的客服槽位清零，剩余名额在其他人之间按权重重新分摊
	total := uint64(0)
	for i, a := range d.agents {
		if !d.available(ctx, a) {
			slots[i] = 0
		}
		total += slots[i]
	}
	if total == 0 {
		return nil, ErrNoAvailableAgent
	}
	quota := weightQuota(slots, total, n)

	// 从游标之后开始，按环形顺序多轮发放，直到配额发完
	order := d.ringOrder()
	out := make([]Agent, 0, n)
	remain := make([]uint64, len(quota))
	copy(remain, quota)
	lastPos := -1
	for len(out) < n {
		moved := false
		for _, pos := range order {
			if remain[pos] == 0 {
				continue
			}
			out = append(out, d.agents[pos])
			remain[pos]--
			lastPos = pos
			moved = true
			if len(out) >= n {
				break
			}
		}
		if !moved { // 配额已发完（正常不会走到这里）
			break
		}
	}
	if lastPos >= 0 {
		d.assigned = true
		d.lastID = d.agents[lastPos].GetAgentId()
		d.lastIdx = instanceIndexAt(d.agents, lastPos)
	}
	return out
}

// weightSlots 计算每个实例占的“槽位数”与总槽位数。调用方需持锁。
// 槽位数就是该实例的权重（未设置按 1）；同一 ID 有多个实例时，该 ID 的权重自然叠加。
func (d *Dispatcher) weightSlots() ([]uint64, uint64) {
	slots := make([]uint64, len(d.agents))
	total := uint64(0)
	for i, a := range d.agents {
		slot := AgentWeight(a)
		slots[i] = slot
		total += slot
	}
	return slots, total
}

// ringOrder 返回按环形顺序排列的下标序列：从游标之后的那个位置开始，到末尾再回绕到开头。
// 调用方需持锁。
func (d *Dispatcher) ringOrder() []int {
	start := d.nextPos()
	order := make([]int, 0, len(d.agents))
	for i := start; i < len(d.agents); i++ {
		order = append(order, i)
	}
	for i := 0; i < start; i++ {
		order = append(order, i)
	}
	return order
}

// TotalWeight 返回当前所有 Agent 的权重之和（未设置权重按 1 计，同一 ID 的多个实例权重叠加）。
func (d *Dispatcher) TotalWeight() uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, total := d.weightSlots()
	return total
}
