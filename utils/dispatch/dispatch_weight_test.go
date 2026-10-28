package dispatch

import (
	"context"
	"testing"
)

// newDispatcherOf 直接用给定的 Agent 列表构造分配器（内部会按 ID 升序排好）
func newDispatcherOf(agents []Agent) *Dispatcher {
	d, _ := NewDispatcher(context.Background(), &Dispatcher{})
	d.SetAgentList(agents)
	return d
}

// newWeightDispatcher 用 id -> weight 构造一个按权重分配的分配器
func newWeightDispatcher(weights map[uint64]uint64) *Dispatcher {
	agents := make([]Agent, 0, len(weights))
	for id, w := range weights {
		agents = append(agents, NewAgentFromUint64WithWeight(id, w))
	}
	return newDispatcherOf(agents)
}

// countByID 统计分配结果里每个 ID 分到的数量
func countByID(got []Agent) map[uint64]int {
	m := make(map[uint64]int)
	for _, a := range got {
		m[a.GetAgentId()]++
	}
	return m
}

// idsOf 按顺序取出分配结果的 ID 序列，便于断言交错顺序
func idsOf(got []Agent) []uint64 {
	ids := make([]uint64, 0, len(got))
	for _, a := range got {
		ids = append(ids, a.GetAgentId())
	}
	return ids
}

func equalUint64Slice(a, b []uint64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestAgentWeight 验证默认权重与自定义权重
func TestAgentWeight(t *testing.T) {
	if w := AgentWeight(NewAgentFromUint64(1)); w != 1 {
		t.Errorf("未设权重的 Agent 期望按 1，实际 %d", w)
	}
	if w := AgentWeight(NewAgentFromUint64WithWeight(1, 5)); w != 5 {
		t.Errorf("权重 5 实际 %d", w)
	}
	if w := AgentWeight(NewAgentFromUint64WithWeight(1, 0)); w != 1 {
		t.Errorf("权重 0 期望按 1，实际 %d", w)
	}
	if w := AgentWeight(NewAgentFromString("jack")); w != 1 {
		t.Errorf("字符串 Agent 未设权重期望 1，实际 %d", w)
	}
	if w := AgentWeight(NewAgentFromStringWithWeight("jack", 3)); w != 3 {
		t.Errorf("字符串 Agent 权重 3 实际 %d", w)
	}

	// SetWeight 链式调用
	a := NewAgentFromUint64(7).SetWeight(9)
	if a.GetAgentId() != 7 || a.GetWeight() != 9 {
		t.Errorf("SetWeight 后 id=%d weight=%d", a.GetAgentId(), a.GetWeight())
	}

	// 自定义 Agent 未实现 AgentWeighter 时按 1 处理
	if w := AgentWeight(&customAgent{id: 10}); w != 1 {
		t.Errorf("未实现 AgentWeighter 期望 1，实际 %d", w)
	}
}

// customAgent 只实现了 Agent 接口，未实现 AgentWeighter
type customAgent struct {
	id uint64
}

func (c *customAgent) GetAgentId() uint64 { return c.id }

// TestAssignNByWeight 验证按权重比例分配
func TestAssignNByWeight(t *testing.T) {
	testCases := []struct {
		name    string
		weights map[uint64]uint64
		n       int
		wantCnt map[uint64]int
	}{
		{
			name:    "权重3:1分4个",
			weights: map[uint64]uint64{1: 3, 2: 1},
			n:       4,
			wantCnt: map[uint64]int{1: 3, 2: 1},
		},
		{
			name:    "权重相等均分",
			weights: map[uint64]uint64{1: 1, 2: 1},
			n:       4,
			wantCnt: map[uint64]int{1: 2, 2: 2},
		},
		{
			name:    "权重2:1分10个",
			weights: map[uint64]uint64{1: 2, 2: 1},
			n:       10,
			wantCnt: map[uint64]int{1: 7, 2: 3}, // 6.67 -> 7，3.33 -> 3
		},
		{
			name:    "权重1:2:3分7个",
			weights: map[uint64]uint64{1: 1, 2: 2, 3: 3},
			n:       7,
			wantCnt: map[uint64]int{1: 1, 2: 2, 3: 4}, // 1.17/2.33/3.5，余下的 1 个给余数最大的 3
		},
		{
			name:    "订单数少于人数",
			weights: map[uint64]uint64{1: 5, 2: 3, 3: 2},
			n:       2,
			wantCnt: map[uint64]int{1: 1, 2: 1, 3: 0},
		},
		{
			name:    "全部未设权重则均分",
			weights: map[uint64]uint64{1: 0, 2: 0, 3: 0},
			n:       6,
			wantCnt: map[uint64]int{1: 2, 2: 2, 3: 2},
		},
		{
			name:    "权重悬殊",
			weights: map[uint64]uint64{1: 100, 2: 1},
			n:       10,
			wantCnt: map[uint64]int{1: 10, 2: 0}, // 0.099 -> 0，不会强行给人分 1 个
		},
		{
			name:    "n大于总数时按权重放大",
			weights: map[uint64]uint64{1: 3, 2: 1},
			n:       9,
			wantCnt: map[uint64]int{1: 7, 2: 2}, // 6.75 -> 7，2.25 -> 2
		},
	}

	for _, tc := range testCases {
		d := newWeightDispatcher(tc.weights)
		got := d.AssignN(context.Background(), tc.n)
		if len(got) != tc.n {
			t.Errorf("%s: 期望分配 %d 个，实际 %d 个", tc.name, tc.n, len(got))
			continue
		}
		gotCnt := countByID(got)
		for id, want := range tc.wantCnt {
			if gotCnt[id] != want {
				t.Errorf("%s: ID=%d 期望 %d 个，实际 %d 个（整体 %v）", tc.name, id, want, gotCnt[id], gotCnt)
			}
		}
	}
}

// TestAssignNInterleave 验证发放遵循“加权最少已分”：权重 3:1 时，权重大的因 ratio 更小而
// 先多分，追平后再轮询，而不是把一个人先发满。顺序为 [1,1,2,1]。
func TestAssignNInterleave(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 3, 2: 1})
	got := d.AssignN(context.Background(), 4)
	g := idsOf(got)
	want := []uint64{1, 1, 2, 1}
	if !equalUint64Slice(g, want) {
		t.Errorf("期望顺序 %v，实际 %v", want, g)
	}
}

// TestAssignNMultiBatch 多次连续调用后整体仍按权重分摊
func TestAssignNMultiBatch(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 3, 2: 1})
	total := map[uint64]int{}
	for i := 0; i < 10; i++ { // 每批 4 个，共 40 个
		batch := d.AssignN(context.Background(), 4)
		for id, c := range countByID(batch) {
			total[id] += c
		}
	}
	if total[1] != 30 || total[2] != 10 {
		t.Errorf("40 个按 3:1 期望 30/10，实际 %v", total)
	}
}

// TestAssignNCursorContinue 验证 AssignN 也跟游标成环：
// SetLastAssigned(1) 后从“>1 的第一个”(2) 续接、回绕到 1；每批整体仍按权重(此处均权)均分。
func TestAssignNCursorContinue(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 1, 2: 1, 3: 1})
	d.SetLastAssigned(1) // 游标影响 AssignN 的起始位置
	got := d.AssignN(context.Background(), 3)
	g := idsOf(got)
	want := []uint64{2, 3, 1}
	if !equalUint64Slice(g, want) {
		t.Errorf("游标续接期望 %v，实际 %v", want, g)
	}

	// 第二批从游标(已停在 1)之后继续，仍是 2,3,1 的环形续接
	got2 := d.AssignN(context.Background(), 3)
	g2 := idsOf(got2)
	if !equalUint64Slice(g2, want) {
		t.Errorf("第二批期望 %v，实际 %v", want, g2)
	}
}

// TestAssignNStatefulProportional 验证有状态按比例：即便批次大小不一致、与单发 Assign 混用，
// 累计分布仍贴合权重（这是“每次独立按比例”做不到的）。
func TestAssignNStatefulProportional(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 3, 2: 1})

	// 模拟“已有订单”：在分配器启动前，ID=1 已经积了 30 单，ID=2 积了 10 单
	d.SetIssued(map[Agent]uint64{
		NewAgentFromUint64WithWeight(1, 3): 30,
		NewAgentFromUint64WithWeight(2, 1): 10,
	})

	// 之后每批 4 个，连续 10 批（共 40 个新单），权重 3:1 应让新增量也约 30:10
	total := map[uint64]int{}
	for i := 0; i < 10; i++ {
		batch := d.AssignN(context.Background(), 4)
		for id, c := range countByID(batch) {
			total[id] += c
		}
	}
	if total[1] != 30 || total[2] != 10 {
		t.Errorf("新增 40 单按 3:1 期望 30/10，实际 %v", total)
	}

	// 累计（含初始）应维持 3:1：ID=1 共 60，ID=2 共 20
	if d.issued[1] != 60 || d.issued[2] != 20 {
		t.Errorf("累计发放应为 60/20，实际 %d/%d", d.issued[1], d.issued[2])
	}
}

// TestAssignNOneByOneProportional 验证“一个一个来分”（AssignN(1)）也能贴权重，
// 不会像纯轮询那样失真。
func TestAssignNOneByOneProportional(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 3, 2: 1})
	for i := 0; i < 40; i++ {
		d.AssignN(context.Background(), 1)
	}
	if d.issued[1] != 30 || d.issued[2] != 10 {
		t.Errorf("40 单按 3:1 期望 30/10，实际 %d/%d", d.issued[1], d.issued[2])
	}
}

// TestAssignNEqualWeightRoundRobin 验证权重全相等时，分配为严格平均的循环（round-robin）：
// 既不会偏向某个 ID，也不会把一个人先发满，而是 ID 顺序交替（1,2,3,1,2,3…）。
func TestAssignNEqualWeightRoundRobin(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 1, 2: 1, 3: 1})
	var seq []uint64
	for i := 0; i < 12; i++ { // 4 整轮
		b := d.AssignN(context.Background(), 1)
		seq = append(seq, b[0].GetAgentId())
	}
	want := []uint64{1, 2, 3, 1, 2, 3, 1, 2, 3, 1, 2, 3}
	if !equalUint64Slice(seq, want) {
		t.Errorf("等权重期望严格循环 %v，实际 %v", want, seq)
	}
	// 每个人恰好 4 个，完全平均
	cnt := map[uint64]int{}
	for _, id := range seq {
		cnt[id]++
	}
	if cnt[1] != 4 || cnt[2] != 4 || cnt[3] != 4 {
		t.Errorf("等权重应每人 4 个，实际 %v", cnt)
	}
}

// TestAssignNSameIDInstances 同一 ID 多个实例时，每个实例至少占一个槽位
func TestAssignNSameIDInstances(t *testing.T) {
	agents := []Agent{
		NewAgentFromUint64WithWeight(1, 1),
		NewAgentFromUint64WithWeight(1, 1), // 同一 ID 第二个实例
		NewAgentFromUint64WithWeight(2, 4),
	}
	d := newDispatcherOf(agents)
	got := d.AssignN(context.Background(), 6)
	cnt := countByID(got)
	if cnt[1] != 2 || cnt[2] != 4 {
		t.Errorf("期望 ID=1 分 2 个、ID=2 分 4 个，实际 %v", got)
	}
}

// TestAssignNEmpty 边界：空列表、n<=0
func TestAssignNEmpty(t *testing.T) {
	d := newDispatcherOf(nil)
	// 空列表：返回空切片（不再返回 error）
	if got := d.AssignN(context.Background(), 5); len(got) != 0 {
		t.Errorf("空列表期望空切片，实际 got=%v", got)
	}
	d2 := newWeightDispatcher(map[uint64]uint64{1: 1, 2: 1})
	// n<=0：返回 nil（无需分配）
	if got := d2.AssignN(context.Background(), 0); got != nil {
		t.Errorf("n=0 期望 nil，实际 got=%v", got)
	}
	if got := d2.AssignN(context.Background(), -1); got != nil {
		t.Errorf("n=-1 期望 nil，实际 got=%v", got)
	}
}

// TestAssignNNotAffectAssign 权重不影响单个 Assign 的环形轮询
func TestAssignNNotAffectAssign(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 9, 2: 1})
	got := make([]uint64, 0, 4)
	for i := 0; i < 4; i++ {
		a, _ := d.Assign(context.Background())
		got = append(got, a.GetAgentId())
	}
	want := []uint64{1, 2, 1, 2}
	if !equalUint64Slice(got, want) {
		t.Errorf("单个 Assign 期望仍是 1:1 轮询 %v，实际 %v", want, got)
	}
}

// TestAssignNPerAgentRatio 验证按个人（而非按权重值分组）计算比例：
// 同权重的 B 积压很多单，不应拖累同权重、零单的 C。
func TestAssignNPerAgentRatio(t *testing.T) {
	A := NewAgentFromUint64WithWeight(1, 2)
	B := NewAgentFromUint64WithWeight(2, 1)
	C := NewAgentFromUint64WithWeight(3, 1)
	d := newDispatcherOf([]Agent{A, B, C})
	d.SetIssued(map[Agent]uint64{B: 100})

	cnt := countByID(d.AssignN(context.Background(), 9))
	if cnt[1] != 6 || cnt[2] != 0 || cnt[3] != 3 {
		t.Errorf("A(w2):C(w1) 期望 6:3、B 不分，实际 %v", cnt)
	}
}

// TestAssignNNewcomerJoinsRotation 验证新加入的人直接加入轮转，而不是因历史欠账被连续灌单。
func TestAssignNNewcomerJoinsRotation(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 1, 2: 1})
	d.AssignN(context.Background(), 200)
	d.SetAgentList([]Agent{NewAgentFromUint64(1), NewAgentFromUint64(2), NewAgentFromUint64(3)})

	cnt := countByID(d.AssignN(context.Background(), 9))
	if cnt[1] != 3 || cnt[2] != 3 || cnt[3] != 3 {
		t.Errorf("新人加入后期望三人轮流各 3 单，实际 %v", cnt)
	}
}

// TestAssignNReturnerJoinsRotation 验证被跳过（下线）后回归的人直接加入轮转。
func TestAssignNReturnerJoinsRotation(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 1, 2: 1, 3: 1})
	ctx := context.Background()
	d.AssignN(ctx, 30)
	d.SetSkipBatchFunc(SkipFuncFromPred(func(a Agent) bool { return a.GetAgentId() == 3 }))
	d.AssignN(ctx, 100)
	d.SetSkipBatchFunc(nil)

	cnt := countByID(d.AssignN(ctx, 9))
	if cnt[1] != 3 || cnt[2] != 3 || cnt[3] != 3 {
		t.Errorf("回归后期望三人轮流各 3 单，实际 %v", cnt)
	}
}

// TestAssignNNewcomerWeighted 验证新人按自身权重折算对齐，之后按权重比例接单。
func TestAssignNNewcomerWeighted(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 1, 2: 1})
	ctx := context.Background()
	d.AssignN(ctx, 100) // 每人 50
	d.SetAgentList([]Agent{
		NewAgentFromUint64WithWeight(1, 1),
		NewAgentFromUint64WithWeight(2, 1),
		NewAgentFromUint64WithWeight(3, 2),
	})

	cnt := countByID(d.AssignN(ctx, 40))
	if cnt[1] != 10 || cnt[2] != 10 || cnt[3] != 20 {
		t.Errorf("新人权重 2 期望 10/10/20，实际 %v", cnt)
	}
}

// TestSetIssuedNotAligned 验证 SetIssued 显式设置过的人不被对齐，其历史欠账照常补齐。
func TestSetIssuedNotAligned(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 1, 2: 1})
	ctx := context.Background()
	d.AssignN(ctx, 20) // 每人 10
	C := NewAgentFromUint64(3)
	d.SetAgentList([]Agent{NewAgentFromUint64(1), NewAgentFromUint64(2), C})
	d.SetIssued(map[Agent]uint64{C: 4})

	cnt := countByID(d.AssignN(ctx, 6))
	if cnt[3] != 6 {
		t.Errorf("C 显式设置为 4 单，期望先补 6 单追平，实际 %v", cnt)
	}
}
