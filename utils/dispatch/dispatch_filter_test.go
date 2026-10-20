package dispatch

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// TestAssignSkipFunc 验证设置了跳过条件后，被跳过的客服永远分不到单，其余仍成环。
func TestAssignSkipFunc(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20, 30)), 0)
	if err != nil {
		t.Fatal(err)
	}
	// 20 休假，永远跳过
	d.SetSkipFunc(func(_ context.Context, a Agent) bool {
		return a.GetAgentId() == 20
	})

	got := make([]uint64, 0, 4)
	for i := 0; i < 4; i++ {
		a, _ := d.Assign(context.Background())
		if a == nil {
			t.Fatalf("assign %d 返回 nil", i)
		}
		got = append(got, a.GetAgentId())
	}
	want := []uint64{10, 30, 10, 30}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("跳过后的顺序错误: got=%v want=%v", got, want)
		}
	}
}

// TestAssignSkipFuncDynamic 验证跳过条件是每次分配实时判断的：
// 同一个客服先不跳过、后跳过（模拟临时改状态），无需刷新列表即可生效。
func TestAssignSkipFuncDynamic(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20, 30)), 0)
	if err != nil {
		t.Fatal(err)
	}
	skip := false
	d.SetSkipFunc(func(_ context.Context, a Agent) bool {
		return skip && a.GetAgentId() == 20
	})

	if a, _ := d.Assign(context.Background()); a.GetAgentId() != 10 {
		t.Fatalf("期望 10，实际 %d", a.GetAgentId())
	}
	skip = true // 20 从现在起不接单
	if a, _ := d.Assign(context.Background()); a.GetAgentId() != 30 {
		t.Fatalf("20 被跳过后期望 30，实际 %d", a.GetAgentId())
	}
	// 此时 cursor 停在 30，下一单回绕到 10
	if a, _ := d.Assign(context.Background()); a.GetAgentId() != 10 {
		t.Fatalf("回绕期望 10，实际 %d", a.GetAgentId())
	}
	skip = false // 恢复，20 重新参与
	if a, _ := d.Assign(context.Background()); a.GetAgentId() != 20 {
		t.Fatalf("恢复后期望 20，实际 %d", a.GetAgentId())
	}
}

// TestAssignAllSkipped 验证所有人都不满足条件时 Assign 返回 nil，且不推进游标（之后恢复可继续）。
func TestAssignAllSkipped(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20)), 0)
	if err != nil {
		t.Fatal(err)
	}
	allSkip := true
	d.SetSkipFunc(func(_ context.Context, _ Agent) bool { return allSkip })

	if a, _ := d.Assign(context.Background()); a != nil {
		t.Fatalf("全部跳过时应返回 nil，实际 %d", a.GetAgentId())
	}
	allSkip = false
	if a, _ := d.Assign(context.Background()); a == nil || a.GetAgentId() != 10 {
		t.Fatalf("恢复后首次应回到 10，实际 %v", a)
	}
}

// TestAssignQuotaSkip 验证用跳过条件实现限量：达到上限的人不再分单，名额自然落到其他人头上。
func TestAssignQuotaSkip(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20, 30)), 0)
	if err != nil {
		t.Fatal(err)
	}
	limit := map[uint64]uint64{10: 1, 20: 1, 30: 1} // 每人最多 1 单
	used := map[uint64]uint64{}
	d.SetSkipFunc(func(_ context.Context, a Agent) bool {
		return used[a.GetAgentId()] >= limit[a.GetAgentId()]
	})

	// 第一轮：10,20,30 各一单
	for i, want := range []uint64{10, 20, 30} {
		a, _ := d.Assign(context.Background())
		if a == nil {
			t.Fatalf("assign %d 返回 nil", i)
		}
		if a.GetAgentId() != want {
			t.Fatalf("assign %d: got=%d want=%d", i, a.GetAgentId(), want)
		}
		used[a.GetAgentId()]++
	}
	// 第二轮：全部已达上限，无人可分
	if a, _ := d.Assign(context.Background()); a != nil {
		t.Fatalf("全部达限量时应返回 nil，实际 %d", a.GetAgentId())
	}
	// 30 释放一单后，只有 30 可分
	used[30] = 0
	if a, _ := d.Assign(context.Background()); a == nil || a.GetAgentId() != 30 {
		t.Fatalf("放量后期望 30，实际 %v", a)
	}
}

// TestAssignNWithSkip 验证批量分配时被跳过的人权重被剔除，名额在剩余人之间按权重重新分摊。
func TestAssignNWithSkip(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20, 30)), 0)
	if err != nil {
		t.Fatal(err)
	}
	// 30 不参与分单
	d.SetSkipFunc(func(_ context.Context, a Agent) bool { return a.GetAgentId() == 30 })

	got := d.AssignN(context.Background(), 4)
	if len(got) != 4 {
		t.Fatalf("期望分配 4 个，实际 %d", len(got))
	}
	count := map[uint64]int{}
	for _, a := range got {
		count[a.GetAgentId()]++
	}
	if count[30] != 0 {
		t.Fatalf("被跳过的 30 不应分到单: %v", count)
	}
	if count[10] != 2 || count[20] != 2 {
		t.Fatalf("剩余两人应平分 4 单，实际 %v", count)
	}
}

// TestAssignNSkipRedistribute 验证有人被跳过时名额重新分摊：20 被跳过，3 单全给 10。
func TestAssignNSkipRedistribute(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20)), 0)
	if err != nil {
		t.Fatal(err)
	}
	d.SetSkipFunc(func(_ context.Context, a Agent) bool { return a.GetAgentId() == 20 })

	got := d.AssignN(context.Background(), 3)
	if len(got) != 3 {
		t.Fatalf("期望 3 个，实际 %d", len(got))
	}
	for _, a := range got {
		if a.GetAgentId() != 10 {
			t.Fatalf("限量后都应分给 10，实际 %d", a.GetAgentId())
		}
	}
}

// TestAssignNAllSkipped 验证所有人都不分单时 AssignN 返回空切片。
func TestAssignNAllSkipped(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20)), 0)
	if err != nil {
		t.Fatal(err)
	}
	d.SetSkipFunc(func(_ context.Context, _ Agent) bool { return true })
	got := d.AssignN(context.Background(), 3)
	if len(got) != 0 {
		t.Fatalf("全部跳过时期望空切片，实际 %v", got)
	}
}

// TestAssignWithCtx 验证 ctx 能透传给 Skip 回调。
func TestAssignWithCtx(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20)), 0)
	if err != nil {
		t.Fatal(err)
	}
	type ctxKey struct{}
	gotCtx := false
	d.SetSkipFunc(func(ctx context.Context, _ Agent) bool {
		if v, ok := ctx.Value(ctxKey{}).(string); ok && v == "trace" {
			gotCtx = true
		}
		return false
	})
	d.Assign(context.WithValue(context.Background(), ctxKey{}, "trace"))
	if !gotCtx {
		t.Fatal("ctx 未透传给 Skip 回调")
	}
}

// TestDynamicAgentList 验证人员变动（新增/下线）时整体替换列表后仍成环不重复。
func TestDynamicAgentList(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(10, 20, 30)), 0)
	if err != nil {
		t.Fatal(err)
	}
	d.Assign(context.Background()) // 10
	d.Assign(context.Background()) // 20

	// 20 下线、新增 25
	d.SetAgentList(agents(10, 25, 30))
	got := make([]uint64, 0, 3)
	for i := 0; i < 3; i++ {
		a, _ := d.Assign(context.Background())
		if a == nil {
			t.Fatalf("assign %d 返回 nil", i)
		}
		got = append(got, a.GetAgentId())
	}
	// cursor=(20,0)，20 已下线 → 跳到第一个大于 20 的 25，再回绕
	want := []uint64{25, 30, 10}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("人员变动后顺序错误: got=%v want=%v", got, want)
		}
	}
}

// TestAgentListGetter 验证 AgentList 返回按分配顺序排列的副本。
func TestAgentListGetter(t *testing.T) {
	d, err := newDispatcher(context.Background(), 10, fixedFetch(agents(30, 10, 20)), 0)
	if err != nil {
		t.Fatal(err)
	}
	list := d.AgentList()
	if len(list) != 3 {
		t.Fatalf("期望 3 个，实际 %d", len(list))
	}
	for i, want := range []uint64{10, 20, 30} {
		if list[i].GetAgentId() != want {
			t.Fatalf("列表顺序错误: got[%d]=%d want=%d", i, list[i].GetAgentId(), want)
		}
	}
}

// TestSlowFetchNoPileUp 验证查询变慢时定时任务不会堆积、不会并发刷新：
// 节拍是“上次结束后再等 Period”，因此慢查询只会导致刷新变稀疏，不会排队补跑。
func TestSlowFetchNoPileUp(t *testing.T) {
	var (
		mu          sync.Mutex
		inflight    int
		maxInflight int
		calls       int
	)
	slowFetch := func(_ context.Context) ([]Agent, error) {
		mu.Lock()
		inflight++
		if inflight > maxInflight {
			maxInflight = inflight
		}
		calls++
		mu.Unlock()

		time.Sleep(50 * time.Millisecond) // 比 Period 慢得多

		mu.Lock()
		inflight--
		mu.Unlock()
		return agents(10, 20), nil
	}

	d, err := newDispatcher(context.Background(), 10, slowFetch, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	_ = d

	time.Sleep(300 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if maxInflight != 1 {
		t.Fatalf("出现了并发刷新，最大并发数=%d", maxInflight)
	}
	// 每轮耗时 ≈ 20ms(等待) + 50ms(查询) = 70ms，300ms 内约 4 次；这里只卡上界防止堆积
	if calls == 0 || calls > 8 {
		t.Fatalf("刷新次数异常，疑似堆积或停摆: calls=%d", calls)
	}
}

// TestFetchTimeout 验证 FetchTimeout 会给 Fetch 套上超时，慢查询不会永久卡住刷新循环。
func TestFetchTimeout(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	d, err := newDispatcher(ctx, 10, func(ctx context.Context) ([]Agent, error) {
		select {
		case <-ctx.Done(): // 感知到超时/取消就立刻返回
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
			return agents(10, 20), nil
		}
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	d.Period = 10 * time.Millisecond
	d.FetchTimeout = 20 * time.Millisecond

	start := time.Now()
	err = d.Refresh(context.Background())
	if err == nil {
		t.Fatal("期望 Fetch 因超时返回错误，实际为 nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("期望 DeadlineExceeded，实际 %v", err)
	}
	if used := time.Since(start); used > time.Second {
		t.Fatalf("超时未生效，Refresh 耗时 %v", used)
	}
}

// TestInitWithAgentsField 验证不配 Fetch 时，可用 Agents 字段直接初始化。
func TestInitWithAgentsField(t *testing.T) {
	d, err := NewDispatcher(context.Background(), &Dispatcher{
		StartID: 20,
		Agents:  agents(10, 20, 30),
	})
	if err != nil {
		t.Fatal(err)
	}
	if a, _ := d.Assign(context.Background()); a == nil || a.GetAgentId() != 20 {
		t.Fatalf("期望从 20 开始，实际 %v", a)
	}
}

// TestRefreshManually 验证定时拉取之外，还能手动 Refresh 立即更新列表。
func TestRefreshManually(t *testing.T) {
	var current = agents(10, 20)
	d, err := newDispatcher(context.Background(), 10, func(_ context.Context) ([]Agent, error) {
		return current, nil
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	current = agents(10, 20, 30)
	if err = d.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if list := d.AgentList(); len(list) != 3 {
		t.Fatalf("刷新后期望 3 个，实际 %d", len(list))
	}
}

// TestPeriodAutoRefresh 验证设置了 Period 后会按间隔自动拉取最新列表。
func TestPeriodAutoRefresh(t *testing.T) {
	var (
		mu      sync.Mutex // 后台刷新协程会读 current，测试协程要写，需加锁避免测试自身的竞态
		current = agents(10, 20)
	)
	d, err := newDispatcher(context.Background(), 10, func(_ context.Context) ([]Agent, error) {
		mu.Lock()
		defer mu.Unlock()
		return current, nil
	}, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	mu.Lock()
	current = agents(10, 20, 30)
	mu.Unlock()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(d.AgentList()) == 3 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("定时刷新未生效，列表仍为 %d 个", len(d.AgentList()))
}
