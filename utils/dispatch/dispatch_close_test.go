package dispatch

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// Close 后后台刷新停止（调用方传 nil ctx 时原来无法停止）
func TestDispatcherClose(t *testing.T) {
	var fetches atomic.Int32
	d, err := NewDispatcher(nil, &Dispatcher{
		Period: 20 * time.Millisecond,
		Fetch: func(ctx context.Context) ([]Agent, error) {
			fetches.Add(1)
			return []Agent{NewAgentFromUint64(1)}, nil
		},
	})
	if err != nil {
		t.Fatalf("NewDispatcher: %v", err)
	}
	time.Sleep(80 * time.Millisecond)
	d.Close()
	d.Close() // 可重复调用
	time.Sleep(30 * time.Millisecond)
	before := fetches.Load()
	time.Sleep(100 * time.Millisecond)
	if after := fetches.Load(); after != before {
		t.Fatalf("Close 后仍在刷新: %d -> %d", before, after)
	}
}

// SkipBatch 在锁外执行：回调内调用 Dispatcher 的加锁方法不会死锁
func TestSkipBatchOutsideLock(t *testing.T) {
	d, err := NewDispatcher(context.Background(), &Dispatcher{
		Agents: []Agent{NewAgentFromUint64(1), NewAgentFromUint64(2)},
	})
	if err != nil {
		t.Fatalf("NewDispatcher: %v", err)
	}
	defer d.Close()
	d.SetSkipBatchFunc(func(ctx context.Context, agents []Agent) map[uint64]bool {
		_ = d.AgentList() // 原来在持写锁时回调，这里会死锁
		return map[uint64]bool{1: true}
	})
	done := make(chan Agent, 1)
	go func() {
		a, _ := d.Assign(context.Background())
		done <- a
	}()
	select {
	case a := <-done:
		if a == nil || a.GetAgentId() != 2 {
			t.Fatalf("Assign = %v, want agent 2", a)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("SkipBatch 回调中调用 AgentList 死锁")
	}
	if got := d.AssignN(context.Background(), 2); len(got) != 2 || got[0].GetAgentId() != 2 || got[1].GetAgentId() != 2 {
		t.Fatalf("AssignN 应跳过 agent 1, got %v", got)
	}
}
