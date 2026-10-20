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
	agents    []Agent           // 当前参与分配的客服，按 GetUint64() 升序排列
	lastID    uint64            // 上一次已分配的客服 ID
	lastIdx   int               // lastID 内已分配到的实例序号（0-based）
	assigned  bool              // 是否已分配过（区分首次）
	issued    map[uint64]uint64 // 每人（按 GetAgentId）已发放的单数累计，用于跨批次按比例分摊

	StartID      uint64  // 默认起始客服ID（首次分配从此客服开始）
	Agents       []Agent // 初始客服列表（也可事后用 SetAgentList 动态设置）
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
// 命中 Skip/Limited 的客服会被跳过；当没有可用客服时返回 (nil, sql.ErrNoRows)
// （cursor 不动，下次可重新尝试）。
func (d *Dispatcher) Assign(ctx context.Context) (Agent, error) {
	ctx = ctxOr(ctx)

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.agents) == 0 {
		return nil, sql.ErrNoRows
	}

	target, newID, newIdx := d.nextAssign(ctx)
	if target == nil {
		return nil, sql.ErrNoRows
	}
	// 同步累计发放计数，使 Assign 与 AssignN 共用同一本“账”，
	// AssignN 的按比例分摊才能把 Assign 发出去的单也纳入考虑。
	if d.issued == nil {
		d.issued = make(map[uint64]uint64)
	}
	d.issued[newID]++
	d.lastID = newID
	d.lastIdx = newIdx
	d.assigned = true
	return target, nil
}

// AssignN 一次性分配 n 个（如 n 个订单）。分配遵循“加权最少已分”的贪心 + 游标环形续接：
//
//   - 权重相同：先给已分单数最少的人；当其追平“倒数第二”后，二人开始轮询；
//     再追平“倒数第三”后，三人依次轮询……即“同分时取游标之后最先遇到的”，自然成环。
//   - 权重不同：ratio = (已分单数+1) / 权重，ratio 越小越“落后”越优先，权重大的自然多分；
//     长期累计严格贴合权重比例（与最大余数法等价），但发放顺序更平滑、且跟游标成环。
//   - 始终从“当前游标之后下一个可分单的 Agent”出发向后扫描，选最落后的一个，
//     再把游标推进到该实例之后——保证每单都从当前轮询的下一个续接，不回退、不成环断点。
//   - 命中 Skip 的客服不参与；其名额在剩余人中按上述规则重分。
//
// 返回：
//   - n<=0：返回 nil；
//   - 列表为空或所有人都被跳过：返回空切片；
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
	if d.issued == nil {
		d.issued = make(map[uint64]uint64)
	}

	// 一步到位：遍历 agents，直接收集可分单实例（环形位置）并就地按“权重值”分组。
	// 权重组 = 所有相同权重值的可分单实例；组权重 = 组内各实例权重之和；组已发 = 组内已发之和。
	type wgroup struct {
		weight  uint64
		members []int // 组内实例在 d.agents 中的位置
		issued  uint64
	}
	avail := make([]int, 0, len(d.agents))
	byWeight := make(map[uint64]*wgroup)
	var order []uint64 // 稳定的权重遍历顺序（升序）
	for i, a := range d.agents {
		if !d.available(ctx, a) {
			continue
		}
		avail = append(avail, i)
		w := AgentWeight(a)
		g, ok := byWeight[w]
		if !ok {
			g = &wgroup{}
			byWeight[w] = g
			order = append(order, w)
		}
		g.members = append(g.members, i)
		g.issued += d.issued[a.GetAgentId()]
		g.weight += w
	}
	if len(avail) == 0 {
		return nil
	}
	posIndex := make(map[int]int, len(avail)) // 位置 -> 在 avail 中的下标
	for ai, p := range avail {
		posIndex[p] = ai
	}
	wgroups := make([]*wgroup, 0, len(order))
	for _, w := range order {
		wgroups = append(wgroups, byWeight[w])
	}

	// 从游标续接：取“游标之后第一个可分单”的实例作为起点
	startPos := d.nextAvailablePos(ctx)
	cursor := posIndex[startPos] // 同分比较 / 推进的基准（avail 下标）

	findAgent := make([]Agent, 0, n)
	lastPos := -1
	excluded := make(map[int]bool) // 本批次内“所有 Agent 都无法分配”的权重组，直接跳过
	for step := 0; step < n; step++ {
		// 内层循环：选权重组 -> 选组内 Agent；若该组无任何可分单 Agent，
		// 则将其排除后重试，落到下一个待分配的权重组（不占用本次名额）。
		var g *wgroup
		var bestPos int
		for {
			// 第二步：判断本单落到哪个权重组。
			//   ratio = (组已发+1)/组权重：ratio 越小越“落后”越优先。
			//   —— 所有组都未达标时，按权重比例轮循（ratio 最小的组按权重频次出现）；
			//   —— 部分组已达标时，未达标组 ratio 更小，单直接落到“不满足的权重”上。
			//   同 ratio 时取“游标之后最先遇到的组”，从当前轮询的下一个 Agent 续接。
			bestG := -1
			var bestRatio float64
			bestDist := 0
			for gi, wg := range wgroups {
				if excluded[gi] || len(wg.members) == 0 {
					continue
				}
				ratio := float64(wg.issued+1) / float64(wg.weight)
				dist := len(avail)
				for _, p := range wg.members {
					dd := (posIndex[p] - cursor + len(avail)) % len(avail)
					if dd < dist {
						dist = dd
					}
				}
				if bestG == -1 || ratio < bestRatio || (ratio == bestRatio && dist < bestDist) {
					bestRatio = ratio
					bestG = gi
					bestDist = dist
				}
			}
			if bestG == -1 {
				g = nil
				break
			}
			g = wgroups[bestG]

			// 第三步：在该权重组内，按“已分最少”的 Agent 分配；已分相同时取游标之后最近的，
			// 实现组内轮询，且每次轮询起点都从上一次分完以后的 Agent 开始。
			// （“同一 Agent 分到超过 2 个后才轮询”由“始终选最少”自然保证：不会在有人更少时
			//   还把单堆给同一人，达到均匀轮询的效果。）
			bestPos = -1
			var bestIssued uint64
			bd := 0
			for _, p := range g.members {
				id := d.agents[p].GetAgentId()
				iss := d.issued[id]
				dist := (posIndex[p] - cursor + len(avail)) % len(avail)
				if bestPos == -1 || iss < bestIssued || (iss == bestIssued && dist < bd) {
					bestIssued = iss
					bestPos = p
					bd = dist
				}
			}
			if bestPos == -1 {
				// 该权重下所有 Agent 都无法分配：本批次后续直接跳过该权重，落到下一个权重组
				excluded[bestG] = true
				continue
			}
			break
		}
		if g == nil {
			break // 没有可用权重组了
		}
		pos := bestPos
		findAgent = append(findAgent, d.agents[pos])
		id := d.agents[pos].GetAgentId()
		d.issued[id]++
		g.issued++
		lastPos = pos
		cursor = (posIndex[pos] + 1) % len(avail) // 游标推进到本实例之后
	}

	if lastPos >= 0 {
		d.assigned = true
		d.lastID = d.agents[lastPos].GetAgentId()
		d.lastIdx = instanceIndexAt(d.agents, lastPos)
	}
	return findAgent
}

// SetIssued 初始化（或覆盖）每个人的“已发放单数”（即当前已有的订单数）。
// AssignN 会把它纳入比例计算，因此即便分配器启动前系统里已经存在订单，
// 之后按比例分摊也不会偏向“历史订单少”的人。传 nil 表示清空后重新统计。
func (d *Dispatcher) SetIssued(usedInitNum map[Agent]uint64) *Dispatcher {
	d.mu.Lock()
	defer d.mu.Unlock()
	if usedInitNum == nil {
		d.issued = make(map[uint64]uint64)
		return d
	}
	if d.issued == nil {
		d.issued = make(map[uint64]uint64)
	}
	for one, c := range usedInitNum {
		if one == nil {
			continue
		}
		d.issued[one.GetAgentId()] = c
	}
	return d
}
