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
		got := d.AssignN(tc.n)
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

// TestAssignNInterleave 验证发放是交错均匀的，而不是先把一个人发满
func TestAssignNInterleave(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 3, 2: 1})
	got := idsOf(d.AssignN(4))
	want := []uint64{1, 2, 1, 1}
	if !equalUint64Slice(got, want) {
		t.Errorf("期望顺序 %v，实际 %v", want, got)
	}
}

// TestAssignNMultiBatch 多次连续调用后整体仍按权重分摊
func TestAssignNMultiBatch(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 3, 2: 1})
	total := map[uint64]int{}
	for i := 0; i < 10; i++ { // 每批 4 个，共 40 个
		for id, c := range countByID(d.AssignN(4)) {
			total[id] += c
		}
	}
	if total[1] != 30 || total[2] != 10 {
		t.Errorf("40 个按 3:1 期望 30/10，实际 %v", total)
	}
}

// TestAssignNContinueCursor 验证 AssignN 从当前游标之后开始，且会推进游标
func TestAssignNContinueCursor(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 1, 2: 1, 3: 1})
	d.SetLastAssigned(1)
	got := idsOf(d.AssignN(3))
	want := []uint64{2, 3, 1}
	if !equalUint64Slice(got, want) {
		t.Errorf("期望从 1 之后开始 %v，实际 %v", want, got)
	}

	// 游标已停在 1，下一批从 2 开始
	got2 := idsOf(d.AssignN(3))
	if !equalUint64Slice(got2, want) {
		t.Errorf("第二批期望 %v，实际 %v", want, got2)
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
	if d.TotalWeight() != 6 { // ID=1 两个实例各占 1，ID=2 权重 4
		t.Errorf("总权重期望 6，实际 %d", d.TotalWeight())
	}
	got := countByID(d.AssignN(6))
	if got[1] != 2 || got[2] != 4 {
		t.Errorf("期望 ID=1 分 2 个、ID=2 分 4 个，实际 %v", got)
	}
}

// TestAssignNEmpty 边界：空列表、n<=0
func TestAssignNEmpty(t *testing.T) {
	d := newDispatcherOf(nil)
	if got := d.AssignN(5); len(got) != 0 {
		t.Errorf("空列表期望返回空，实际 %v", got)
	}
	d2 := newWeightDispatcher(map[uint64]uint64{1: 1, 2: 1})
	if got := d2.AssignN(0); len(got) != 0 {
		t.Errorf("n=0 期望返回空，实际 %v", got)
	}
	if got := d2.AssignN(-1); len(got) != 0 {
		t.Errorf("n=-1 期望返回空，实际 %v", got)
	}
}

// TestTotalWeight 验证总权重
func TestTotalWeight(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 3, 2: 1, 3: 6})
	if d.TotalWeight() != 10 {
		t.Errorf("总权重期望 10，实际 %d", d.TotalWeight())
	}
	d2 := newDispatcherOf([]Agent{NewAgentFromUint64(1), NewAgentFromUint64(2), NewAgentFromUint64(3)})
	if d2.TotalWeight() != 3 { // 未设权重按 1 计
		t.Errorf("未设权重总权重期望 3，实际 %d", d2.TotalWeight())
	}
}

// TestAssignNNotAffectAssign 权重不影响单个 Assign 的环形轮询
func TestAssignNNotAffectAssign(t *testing.T) {
	d := newWeightDispatcher(map[uint64]uint64{1: 9, 2: 1})
	got := make([]uint64, 0, 4)
	for i := 0; i < 4; i++ {
		got = append(got, d.Assign().GetAgentId())
	}
	want := []uint64{1, 2, 1, 2}
	if !equalUint64Slice(got, want) {
		t.Errorf("单个 Assign 期望仍是 1:1 轮询 %v，实际 %v", want, got)
	}
}
