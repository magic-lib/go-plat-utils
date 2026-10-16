package dispatch

import (
	"context"
	"github.com/magic-lib/go-plat-utils/goroutines"
	"hash/fnv"
	"log"
	"sort"
	"sync"
	"time"
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

// FetchAgentsFunc 拉取最新列表的函数。
type FetchAgentsFunc func(ctx context.Context) ([]Agent, error)

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

// Dispatcher 环形分配器。
//
// 设计要点：
//   - fetch 后按 GetUint64() 升序排序，保证轮询顺序稳定、可预期；
//   - cursor 记录（上一次已分配的客服 ID, 该 ID 内的实例序号）二元组，
//     因此即便列表中存在多个相同 ID 的客服，也能逐一轮到，不会因 ID 相同而被整体跳过；
//   - 列表增删、程序重启都不影响成环逻辑（cursor 基于 ID+实例序号，跨刷新稳定）；
//   - Assign() 优先在同一 ID 的后续实例间推进，没有则跳到更大的 ID（回绕到最小），
//     从而真正成环、不重复、分配均匀；
//   - 程序重启时，调用 SetLastAssigned(上一单完成的ID)，下次 Assign() 直接拿下一个，不会重复；
//   - Agent 可实现可选的 AgentWeighter 接口声明权重，AssignN(n) 会按权重比例一次性分摊 n 个，
//     权重未设置（0）或没实现该接口时按 1 处理，Assign() 仍是逐个的环形轮询。
type Dispatcher struct {
	mu       sync.RWMutex
	agents   []Agent // 仅包含在线客服，按 GetUint64() 升序排列
	lastID   uint64  // 上一次已分配的客服 ID
	lastIdx  int     // lastID 内已分配到的实例序号（0-based）
	assigned bool    // 是否已分配过（区分首次）
	StartID  uint64  // 默认起始客服ID（首次分配从此客服开始）
	Fetch    FetchAgentsFunc
	Period   time.Duration
}

// NewDispatcher 创建分配器。
//   - startID：默认值，表示环形队列的起始客服（首单分配给它）；
//   - fetch：拉取在线客服列表的回调函数；
//   - period：定时刷新间隔。
func NewDispatcher(ctx context.Context, dConfig *Dispatcher) (*Dispatcher, error) {
	if dConfig == nil {
		dConfig = new(Dispatcher)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	err := dConfig.init(ctx)

	if dConfig.Period > 0 && dConfig.Fetch != nil {
		dConfig.start(ctx)
	}

	return dConfig, err
}

func (d *Dispatcher) init(ctx context.Context) error {
	if d.Fetch == nil {
		return nil
	}

	agents, err := d.Fetch(ctx)
	if err != nil {
		return err
	}
	d.apply(agents)
	return nil
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
func (d *Dispatcher) SetAgentList(agents []Agent) *Dispatcher {
	d.apply(agents)
	return d
}

// apply 用拉取到的最新列表刷新环形队列：仅按 GetUint64() 升序排序，不改变 cursor。
// 由于 cursor 记录的是（ID, 实例序号），列表增删后下一次 Assign() 仍能正确取到下一个。
func (d *Dispatcher) apply(fetched []Agent) {
	// 复制后再排序，避免修改外部切片
	online := make([]Agent, len(fetched))
	copy(online, fetched)
	sort.Slice(online, func(i, j int) bool {
		return online[i].GetAgentId() < online[j].GetAgentId()
	})

	d.mu.Lock()
	d.agents = online
	d.mu.Unlock()
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

// nextAssign 在不加锁的前提下，依据当前 cursor 计算下一次应分配的实例，
// 返回 (目标, 新的lastID, 新的lastIdx)。调用方需自行加锁。
func (d *Dispatcher) nextAssign() (Agent, uint64, int) {
	pos := d.nextPos()
	t := d.agents[pos]
	return t, t.GetAgentId(), instanceIndexAt(d.agents, pos)
}

// Assign 分配一单：返回排序后“同一 ID 的下一个实例，否则大于 cursor 的第一个”客服，
// 并把 cursor 更新为该客服的（ID, 实例序号），使之成环。
// 当没有可用客服时返回 ok=false。
func (d *Dispatcher) Assign() Agent {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.agents) == 0 {
		return nil
	}

	target, newID, newIdx := d.nextAssign()
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
// 返回切片长度 <= n；Agent 为空或 n<=0 时返回 nil。
func (d *Dispatcher) AssignN(n int) []Agent {
	if n <= 0 {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.agents) == 0 {
		return nil
	}

	slots, total := d.weightSlots()
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

// TotalWeight 返回当前所有 Agent 的权重之和（未设置权重按 1 计，同一 ID 的多个实例权重叠加）。
func (d *Dispatcher) TotalWeight() uint64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	_, total := d.weightSlots()
	return total
}

// start 启动定时刷新：每隔 period 拉取一次最新在线客服列表并刷新环形队列。
// 刷新在独立的执行体中串行执行，不会与自身重叠；
// 若某次 fetch 耗时超过 period，期间到达的 tick 会被 time.Ticker 直接丢弃，
// 因此不会出现刷新任务堆积，也不会有并发的 fetch。
func (d *Dispatcher) start(ctx context.Context) {
	goroutines.GoAsync(func(params ...any) {
		ticker := time.NewTicker(d.Period)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				d.refreshOnce(ctx)
			}
		}
	})
}

// refreshOnce 拉取一次最新的客服列表并刷新环形队列，由定时循环调用。
func (d *Dispatcher) refreshOnce(ctx context.Context) {
	agents, err := d.Fetch(ctx)
	if err != nil {
		log.Println("[dispatch] 刷新客服列表失败:", err)
		return
	}
	d.apply(agents)
	log.Printf("[dispatch] 客服列表已刷新，在线客服数=%d\n", len(agents))
}

// indexOf 返回列表中第一个等于 id 的实例位置，不存在返回 -1。
func indexOf(agents []Agent, id uint64) int {
	return indexOfNth(agents, id, 0)
}
