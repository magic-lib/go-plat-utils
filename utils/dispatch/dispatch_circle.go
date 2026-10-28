package dispatch

import (
	"context"
	"database/sql"
	"sync"
	"time"
)

// Dispatcher 环形分配器
type Dispatcher struct {
	mu        sync.RWMutex
	refreshMu sync.Mutex        // 串行化刷新：手动 Refresh 与后台定时刷新互斥，避免并发 fetch / 旧数据覆盖新数据
	agents    []Agent           // 当前参与分配的客服，按 GetAgentId() 升序排列
	lastID    uint64            // 上一次已分配的客服 ID
	lastIdx   int               // lastID 内已分配到的实例序号（0-based）
	assigned  bool              // 是否已分配过（区分首次）
	issued    map[uint64]uint64 // 每人（按 GetAgentId）已发放的单数累计，用于跨批次按比例分摊
	active    map[uint64]bool   // 上一次分配时可分单的 ID 集合，用于识别新加入/回归的人；nil 表示尚未分配过
	pinned    map[uint64]bool   // 经 SetIssued 显式设置过、尚未加入在岗集合的 ID，不参与对齐

	StartID      uint64  // 默认起始客服ID（首次分配从此客服开始）
	Agents       []Agent // 初始客服列表（也可事后用 SetAgentList 动态设置）
	Fetch        FetchAgentsFunc
	Period       time.Duration  // 定时拉取列表
	FetchTimeout time.Duration  // 单次 Fetch 的超时（<=0 表示不限制），防止慢查询把刷新循环卡死
	SkipBatch    SkipAgentsFunc // 可选：批量跳过条件，一次给定全部 Agent 并一次性判断（避免逐 Agent 查库）
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

// SetSkipBatchFunc 设置批量跳过条件：一次性给定全部 Agent 并一次性判断，返回自身便于链式调用。
// 适合跳过判断依赖查库/远程调用的场景（N 次 IO → 1 次批量）：一次查询拿到所有 Agent 的状态，
// 远优于逐 Agent 查询。返回 map[uint64]bool，key 为 Agent.GetAgentId()，value=true 表示该 Agent 不分单。
func (d *Dispatcher) SetSkipBatchFunc(f SkipAgentsFunc) *Dispatcher {
	d.mu.Lock()
	d.SkipBatch = f
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
		// 该 ID 已不存在（全线下），跳到更大的 ID；没有则回绕到最小（与正常成环一致）
		pos := nextGreaterFirstIndex(agents, d.lastID)
		if pos < 0 {
			pos = 0
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

// computeSkip 在持锁状态下一次性计算所有 Agent 的“跳过标记”（key=Agent.GetAgentId()，value=true 表示跳过）：
// 调用一次 SkipBatch 批量判断（适合查库/远程调用，N 次 IO → 1 次批量）。
//
// 用 map[uint64]bool 而非 []bool：避免“返回切片长度与 agents 不一致导致下标错位/越界”；
// 缺失的 Agent 一律视作“不跳过”。之后可用 available(skipMap, pos) 在内存中 O(1) 查表，
// 避免在分配循环里对每个 Agent 重复查库。
// 回调在持锁状态下调用，因此回调里不要再反向调用 Dispatcher 的加锁方法，否则会死锁。
func (d *Dispatcher) computeSkip(ctx context.Context) map[uint64]bool {
	skipMap := make(map[uint64]bool)
	if d.SkipBatch != nil {
		res := d.SkipBatch(ctx, d.agents)
		for id, skip := range res {
			if skip {
				skipMap[id] = true
			}
		}
	}
	return skipMap
}

// available 依据预先计算好的跳过标记（key 为 Agent.GetAgentId()）判断某个位置是否可分单。
// pos 越界或命中跳过条件时返回 false。skipMap 由 computeSkip 在分配入口一次性算出。
func (d *Dispatcher) available(skipMap map[uint64]bool, pos int) bool {
	if pos < 0 || pos >= len(d.agents) {
		return false
	}
	return !skipMap[d.agents[pos].GetAgentId()]
}

// nextAvailablePos 从 nextPos() 开始沿环形向后找第一个“可分单”的实例位置，
// 全部被跳过或已达限量时返回 -1。调用方需持锁。
func (d *Dispatcher) nextAvailablePos(skipMap map[uint64]bool) int {
	if len(d.agents) == 0 {
		return -1
	}
	start := d.nextPos()
	for k := 0; k < len(d.agents); k++ {
		pos := (start + k) % len(d.agents)
		if d.available(skipMap, pos) {
			return pos
		}
	}
	return -1
}

// nextAssign 在不加锁的前提下，依据当前 cursor 计算下一次应分配的实例，
// 返回 (目标, 新的lastID, 新的lastIdx)。调用方需自行加锁。
func (d *Dispatcher) nextAssign(skipMap map[uint64]bool) (Agent, uint64, int) {
	pos := d.nextAvailablePos(skipMap)
	if pos < 0 {
		return nil, 0, 0
	}
	t := d.agents[pos]
	return t, t.GetAgentId(), instanceIndexAt(d.agents, pos)
}

// Assign 分配一单：返回排序后“同一 ID 的下一个实例，否则大于 cursor 的第一个”客服，
// 并把 cursor 更新为该客服的（ID, 实例序号），使之成环。
// 命中 Skip/Limited 的客服会被跳过；当没有可用客服时返回 (nil, sql.ErrNoRows)
// （cursor 不动，下次可重新尝试）。
func (d *Dispatcher) Assign(ctx context.Context) (Agent, error) {
	ctx = ctxOr(ctx)

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.agents) == 0 {
		return nil, sql.ErrNoRows
	}

	flags := d.computeSkip(ctx)
	target, newID, newIdx := d.nextAssign(flags)
	if target == nil {
		return nil, sql.ErrNoRows
	}
	// 同步累计发放计数与在岗集合，使 Assign 与 AssignN 共用同一本“账”，
	// AssignN 的按比例分摊才能把 Assign 发出去的单也纳入考虑。Assign 本身不看权重。
	_, weights := d.availables(flags)
	d.alignNewcomers(weights)
	d.issued[newID]++
	d.lastID = newID
	d.lastIdx = newIdx
	d.assigned = true
	return target, nil
}

// AssignN 一次性分配 n 个（如 n 个订单）。分配遵循“个人加权最少已分”的贪心 + 游标环形续接：
//
//   - 每人 ratio = (已分单数+1) / 权重，ratio 越小越“落后”越优先，权重大的自然多分，
//     长期累计贴合权重比例；同一 ID 的多个实例共享一本账，有效权重为各实例权重之和。
//   - 权重相同：先给已分单数最少的人；追平后依次轮询，自然成环。
//   - ratio 相同时取“游标之后最先遇到的”，再把游标推进到该实例之后——每单都从当前轮询的下一个续接。
//   - 命中 Skip 的客服不参与；其名额在剩余人中按上述规则重分。
//   - 新加入、或下线/被跳过后回归的人，已分单数会对齐到在岗者的最低水平，直接加入轮转，
//     不会因历史欠账被连续灌单（见 alignNewcomers）。
//
// 返回：
//   - n<=0、列表为空或所有人都被跳过：返回 nil；
//   - 正常：返回长度恰好为 n 的切片。
func (d *Dispatcher) AssignN(ctx context.Context, n int) []Agent {
	if n <= 0 {
		return nil
	}
	ctx = ctxOr(ctx)

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.agents) == 0 {
		return nil
	}

	flags := d.computeSkip(ctx)
	avail, weights := d.availables(flags)
	if len(avail) == 0 {
		return nil
	}
	d.alignNewcomers(weights)

	posIndex := make(map[int]int, len(avail)) // 位置 -> 在 avail 中的下标
	for ai, p := range avail {
		posIndex[p] = ai
	}
	// 从游标续接：取“游标之后第一个可分单”的实例作为起点
	cursor := posIndex[d.nextAvailablePos(flags)]

	findAgent := make([]Agent, 0, n)
	lastPos := -1
	for step := 0; step < n; step++ {
		// 选 (已分+1)/权重 最小的实例；用交叉相乘比较，避免浮点误差。同值取游标之后最近的
		bestPos, bestDist := -1, 0
		var bestIss, bestW uint64
		for _, p := range avail {
			id := d.agents[p].GetAgentId()
			iss, w := d.issued[id]+1, weights[id]
			dist := (posIndex[p] - cursor + len(avail)) % len(avail)
			if bestPos == -1 || iss*bestW < bestIss*w || (iss*bestW == bestIss*w && dist < bestDist) {
				bestPos, bestDist, bestIss, bestW = p, dist, iss, w
			}
		}
		findAgent = append(findAgent, d.agents[bestPos])
		d.issued[d.agents[bestPos].GetAgentId()]++
		lastPos = bestPos
		cursor = (posIndex[bestPos] + 1) % len(avail) // 游标推进到本实例之后
	}

	d.assigned = true
	d.lastID = d.agents[lastPos].GetAgentId()
	d.lastIdx = instanceIndexAt(d.agents, lastPos)
	return findAgent
}

// availables 返回可分单实例在 agents 中的位置，以及每个 ID 的有效权重（同 ID 各实例权重之和）。
// 调用方需持锁。
func (d *Dispatcher) availables(skipMap map[uint64]bool) ([]int, map[uint64]uint64) {
	avail := make([]int, 0, len(d.agents))
	weights := make(map[uint64]uint64)
	for i, a := range d.agents {
		if !d.available(skipMap, i) {
			continue
		}
		avail = append(avail, i)
		weights[a.GetAgentId()] += AgentWeight(a)
	}
	return avail, weights
}

// alignNewcomers 把“新加入 / 下线或被跳过后回归”的人的已分单数，对齐到本次仍在岗老成员中的最低水平
// （按 已分/权重 折算），使其直接加入轮转，而不是因历史欠账被连续灌单。
// 以下情况不对齐：首次分配（尚无在岗基准）、经 SetIssued 显式设置过的人。
// weights 为本次可分单的 ID -> 有效权重；调用后它成为新的在岗集合。调用方需持锁。
func (d *Dispatcher) alignNewcomers(weights map[uint64]uint64) {
	if d.issued == nil {
		d.issued = make(map[uint64]uint64)
	}
	if d.active != nil {
		// 基准：本次仍在岗的老成员中 issued/weight 最小者
		var baseIss, baseW uint64
		for id, w := range weights {
			if !d.active[id] {
				continue
			}
			if iss := d.issued[id]; baseW == 0 || iss*baseW < baseIss*w {
				baseIss, baseW = iss, w
			}
		}
		if baseW > 0 {
			for id, w := range weights {
				if d.active[id] || d.pinned[id] {
					continue
				}
				d.issued[id] = baseIss * w / baseW
			}
		}
	}
	d.active = make(map[uint64]bool, len(weights))
	for id := range weights {
		d.active[id] = true
		delete(d.pinned, id)
	}
}

// SetIssued 初始化（或覆盖）每个人的“已发放单数”（即当前已有的订单数）。
// AssignN 会把它纳入比例计算，因此即便分配器启动前系统里已经存在订单，
// 之后按比例分摊也不会偏向“历史订单少”的人；这里显式设置过的人也不会被 alignNewcomers 对齐。
// 传 nil 表示清空后重新统计（同时清空在岗记录，下一次分配视作首次）。
func (d *Dispatcher) SetIssued(usedInitNum map[Agent]uint64) *Dispatcher {
	d.mu.Lock()
	defer d.mu.Unlock()
	if usedInitNum == nil {
		d.issued = make(map[uint64]uint64)
		d.active = nil
		d.pinned = nil
		return d
	}
	if d.issued == nil {
		d.issued = make(map[uint64]uint64)
	}
	if d.pinned == nil {
		d.pinned = make(map[uint64]bool)
	}
	for one, c := range usedInitNum {
		if one == nil {
			continue
		}
		d.issued[one.GetAgentId()] = c
		d.pinned[one.GetAgentId()] = true
	}
	return d
}
