package dispatch

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// agents 由 uint64 ID 构造一组在线客服（无序，用于验证排序）。
func agents(ids ...uint64) []Agent {
	out := make([]Agent, 0, len(ids))
	for _, id := range ids {
		out = append(out, NewAgentFromUint64(id))
	}
	return out
}

// fixedFetch 返回固定列表的拉取函数。
func fixedFetch(list []Agent) FetchAgentsFunc {
	return func(_ context.Context) ([]Agent, error) {
		return list, nil
	}
}

// newDispatcher 以 (startID, fetch, period) 构造分配器，
// 封装 NewDispatcher 的“配置对象”签名，避免每个用例重复写结构体字面量。
func newDispatcher(ctx context.Context, startID uint64, fetch FetchAgentsFunc, period time.Duration) (*Dispatcher, error) {
	return NewDispatcher(ctx, &Dispatcher{
		StartID: startID,
		Fetch:   fetch,
		Period:  period,
	})
}

// TestAssignRingNoDuplicate 验证成环且不重复：顺序经过所有客服后回到起点。
func TestAssignRingNoDuplicate(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20, 30, 40)), 0)
	if err != nil {
		t.Fatalf("init err: %v", err)
	}

	got := make([]uint64, 0, 4)
	for i := 0; i < 4; i++ {
		a := d.Assign()
		if a == nil {
			t.Fatalf("assign %d failed, no agent", i)
		}
		got = append(got, a.GetAgentId())
	}
	want := []uint64{10, 20, 30, 40}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("顺序错误: got=%v want=%v", got, want)
		}
	}

	if a := d.Assign(); a == nil || a.GetAgentId() != 10 {
		t.Fatalf("回绕错误: got=%v", a)
	}
}

// TestAssignUniform 验证单轮内每个客服恰好被分配一次（均匀）。
func TestAssignUniform(t *testing.T) {
	ids := []uint64{3, 1, 4, 2, 5}
	d, err := newDispatcher(context.Background(), 1, fixedFetch(agents(ids...)), 0)
	if err != nil {
		t.Fatal(err)
	}

	count := map[uint64]int{}
	for i := 0; i < len(ids); i++ {
		a := d.Assign()
		if a == nil {
			t.Fatalf("assign %d failed", i)
		}
		count[a.GetAgentId()]++
	}
	for _, id := range ids {
		if count[id] != 1 {
			t.Fatalf("客服 %d 被分配到 %d 次，期望 1 次: %v", id, count[id], count)
		}
	}
}

// TestDuplicateID 验证存在多个相同 ID 时，所有相同 ID 的实例都会被逐个轮到，不会被整体跳过。
func TestDuplicateID(t *testing.T) {
	// 两个相同 ID=10 的客服，加上 20、30。排序后：10,10,20,30
	ids := []uint64{10, 20, 10, 30}
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(ids...)), 0)
	if err != nil {
		t.Fatal(err)
	}

	// 一轮应轮到：10,10,20,30（两个 10 都出现一次）
	got := make([]uint64, 0, 4)
	for i := 0; i < 4; i++ {
		a := d.Assign()
		if a == nil {
			t.Fatalf("assign %d failed", i)
		}
		got = append(got, a.GetAgentId())
	}
	count10 := 0
	for _, v := range got {
		if v == 10 {
			count10++
		}
	}
	if count10 != 2 {
		t.Fatalf("相同ID=10的实例应被分配到 2 次，实际 %d 次: %v", count10, got)
	}
	if got[0] != 10 || got[1] != 10 || got[2] != 20 || got[3] != 30 {
		t.Fatalf("重复ID场景下顺序错误: %v", got)
	}

	// 下一轮继续成环：回绕到 10,10,20,30
	if a := d.Assign(); a == nil || a.GetAgentId() != 10 {
		t.Fatalf("第二轮首个应回绕到 10，实际 %v", a)
	}
}

// TestDuplicateIDUniform 验证大量重复 ID 时，每个实例在完整周期中恰好一次。
func TestDuplicateIDUniform(t *testing.T) {
	// 三个 10，两个 20
	ids := []uint64{10, 20, 10, 20, 10}
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(ids...)), 0)
	if err != nil {
		t.Fatal(err)
	}

	c10, c20 := 0, 0
	for i := 0; i < len(ids); i++ {
		a := d.Assign()
		if a == nil {
			t.Fatalf("assign %d failed", i)
		}
		switch a.GetAgentId() {
		case 10:
			c10++
		case 20:
			c20++
		default:
			t.Fatalf("出现意外ID: %d", a.GetAgentId())
		}
	}
	if c10 != 3 || c20 != 2 {
		t.Fatalf("重复ID周期分配不均: c10=%d c20=%d", c10, c20)
	}
}

// TestAddAgentNoDuplicate 验证运行过程中新增客服后，整体仍成环不重复。
func TestAddAgentNoDuplicate(t *testing.T) {
	var current = agents(10, 20, 30)
	d, err := newDispatcher(context.Background(), 10, func(_ context.Context) ([]Agent, error) {
		return current, nil
	}, 0)
	if err != nil {
		t.Fatal(err)
	}

	d.Assign()
	d.Assign()
	d.Assign()
	// 此时 cursor=(30,0)

	current = agents(10, 20, 25, 30)
	d.SetAgentList(current)

	seen := map[uint64]bool{}
	for i := 0; i < 4; i++ {
		a := d.Assign()
		if a == nil {
			t.Fatalf("assign %d failed", i)
		}
		if seen[a.GetAgentId()] {
			t.Fatalf("出现重复分配: %d", a.GetAgentId())
		}
		seen[a.GetAgentId()] = true
	}
}

// TestRemoveAgentNoDuplicate 验证下线客服被剔除后，其余仍成环不重复。
func TestRemoveAgentNoDuplicate(t *testing.T) {
	var current = agents(10, 20, 30, 40)
	d, err := newDispatcher(context.Background(), 10, func(_ context.Context) ([]Agent, error) {
		return current, nil
	}, 0)
	if err != nil {
		t.Fatal(err)
	}

	d.Assign() // 10, cursor=(10,0)
	d.Assign() // 20, cursor=(20,0)

	current = agents(10, 20, 40)
	d.SetAgentList(current)

	if a := d.Assign(); a == nil || a.GetAgentId() != 40 {
		t.Fatalf("期望 40，实际 %v", a)
	}
	if a := d.Assign(); a == nil || a.GetAgentId() != 10 {
		t.Fatalf("期望 10，实际 %v", a)
	}
	if a := d.Assign(); a == nil || a.GetAgentId() != 20 {
		t.Fatalf("期望 20，实际 %v", a)
	}
	if a := d.Assign(); a == nil || a.GetAgentId() != 40 {
		t.Fatalf("回绕期望 40，实际 %v", a)
	}
}

// TestRestartContinue 验证程序重启后通过 SetLastAssigned 续接，不重复。
func TestRestartContinue(t *testing.T) {
	d1, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20, 30, 40)), 0)
	if err != nil {
		t.Fatal(err)
	}
	d1.Assign()           //10
	d1.Assign()           //20
	d1.Assign()           //30
	lastDone := d1.lastID // =30

	d2, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20, 30, 40)), 0)
	if err != nil {
		t.Fatal(err)
	}
	d2.SetLastAssigned(lastDone) // 设置上一单完成的ID=30

	if a := d2.Assign(); a == nil || a.GetAgentId() != 40 {
		t.Fatalf("重启后期望下一个=40，实际 %v", a)
	}
	if a := d2.Assign(); a == nil || a.GetAgentId() != 10 {
		t.Fatalf("重启后续期望 10，实际 %v", a)
	}
}

// TestStartIDNotExists 验证 startID 不存在时，首次分配回退到最小 ID。
func TestStartIDNotExists(t *testing.T) {
	d, err := newDispatcher(context.Background(), 99, fixedFetch(agents(10, 20, 30)), 0)
	if err != nil {
		t.Fatal(err)
	}
	if a := d.Assign(); a == nil || a.GetAgentId() != 10 {
		t.Fatalf("startID 不存在时，期望最小 10，实际 %v", a)
	}
}

// TestEmptyAgents 验证无可用客服时 Assign 返回 nil。
func TestEmptyAgents(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents()), 0)
	if err != nil {
		t.Fatal(err)
	}
	if a := d.Assign(); a != nil {
		t.Fatalf("空列表应返回 nil，实际 %v", a)
	}
}

// TestRefreshNeverOverlaps 验证定时刷新始终串行：
//  1. 任意时刻最多只有一个 fetch 在执行（刷新不自我重叠）；
//  2. period 远小于 fetch 耗时时，期间到达的 tick 被 time.Ticker 丢弃而非排队补跑，
//     因此 fetch 总次数远小于 ctx 时长 / period。
func TestRefreshNeverOverlaps(t *testing.T) {
	var (
		mu      sync.Mutex
		running int // 当前并发执行的 fetch 数
		maxRun  int // 观测到的最大并发 fetch 数
		total   int // fetch 总执行次数
	)
	slowFetch := func(_ context.Context) ([]Agent, error) {
		mu.Lock()
		running++
		if running > maxRun {
			maxRun = running
		}
		total++
		mu.Unlock()

		time.Sleep(120 * time.Millisecond) // 模拟耗时远超 period 的拉取

		mu.Lock()
		running--
		mu.Unlock()
		return agents(10, 20), nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	// period=10ms，远小于 fetch 的 120ms
	if _, err := newDispatcher(ctx, 10, slowFetch, 10*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	<-ctx.Done()

	mu.Lock()
	defer mu.Unlock()
	if maxRun > 1 {
		t.Fatalf("出现重叠刷新：最大并发 fetch=%d，期望始终不超过 1", maxRun)
	}
	// 300ms / 10ms 理论上有 30 次触发；串行 + 丢弃 tick 后只应跑 3 次左右
	if total > 5 {
		t.Fatalf("刷新次数异常：fetch 共执行 %d 次，期望因串行与丢弃 tick 远小于 30", total)
	}
}

// TestDefaultAgentFromString 验证以字符串构造时，可稳定生成 uint64，
// 且相同字符串得到相同 ID，不同字符串大概率不同（可参与均匀成环）。
func TestDefaultAgentFromString(t *testing.T) {
	a1 := NewAgentFromString("cs-alice")
	a2 := NewAgentFromString("cs-alice")
	a3 := NewAgentFromString("cs-bob")

	if a1.GetAgentId() != a2.GetAgentId() {
		t.Fatal("相同字符串应生成相同 uint64")
	}
	if a1.GetAgentId() == a3.GetAgentId() {
		t.Fatal("不同字符串应生成不同的 uint64")
	}

	d, err := newDispatcher(context.Background(), a1.GetAgentId(),
		fixedFetch([]Agent{a1, a3, NewAgentFromString("cs-carol")}), 0)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[uint64]bool{}
	for i := 0; i < 3; i++ {
		a := d.Assign()
		if a == nil {
			t.Fatalf("assign %d failed", i)
		}
		if seen[a.GetAgentId()] {
			t.Fatalf("字符串客服出现重复分配: %d", a.GetAgentId())
		}
		seen[a.GetAgentId()] = true
	}
}

func TestCurrentAgentFromString(t *testing.T) {
	a1 := NewAgentFromUint64(9)
	a2 := NewAgentFromUint64(2)
	a3 := NewAgentFromUint64(3)

	agentList := []Agent{a1, a2, a3}
	dispatchInstance, _ := NewDispatcher(context.Background(), &Dispatcher{StartID: uint64(2)})
	dispatchInstance.SetAgentList(agentList)

	kk := dispatchInstance.Assign()
	fmt.Println(kk.GetAgentId())
	kk = dispatchInstance.Assign()
	fmt.Println(kk.GetAgentId())
	kk = dispatchInstance.Assign()
	fmt.Println(kk.GetAgentId())
	kk = dispatchInstance.Assign()
	fmt.Println(kk.GetAgentId())

	fmt.Println("--------------------")

	dispatchInstance2, _ := NewDispatcher(context.Background(), &Dispatcher{})
	dispatchInstance2.SetAgentList(agentList)
	dispatchInstance2.SetLastAssigned(uint64(2))

	kk = dispatchInstance2.Assign()
	fmt.Println(kk.GetAgentId())
	kk = dispatchInstance2.Assign()
	fmt.Println(kk.GetAgentId())
	kk = dispatchInstance2.Assign()
	fmt.Println(kk.GetAgentId())
	kk = dispatchInstance2.Assign()
	fmt.Println(kk.GetAgentId())

}
