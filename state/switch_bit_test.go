package state_test

import (
	"sync"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/state"
)

func TestBitSwitchSetAllStates(t *testing.T) {
	bs := state.NewBitSwitch()
	_ = bs.TurnOn(10)

	done := make(chan error, 1)
	go func() { done <- bs.SetAllStates(map[int]bool{0: true, 3: true, 5: false}) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("SetAllStates err: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("SetAllStates 死锁")
	}
	if got := bs.Uint64(); got != 0b1001 {
		t.Fatalf("Uint64 = %b, want 1001", got)
	}

	// 非法位置返回错误，原状态不变
	if err := bs.SetAllStates(map[int]bool{1: true, 64: true}); err == nil {
		t.Fatal("非法位置应返回错误")
	}
	if got := bs.Uint64(); got != 0b1001 {
		t.Fatalf("出错后状态被修改: %b", got)
	}
}

// 并发读写，配合 -race 检查数据竞争
func TestBitSwitchConcurrent(t *testing.T) {
	bs := state.NewBitSwitch()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(pos int) {
			defer wg.Done()
			for j := 0; j < 200; j++ {
				_ = bs.Toggle(pos)
				_, _ = bs.IsOn(pos)
				_ = bs.String()
				_ = bs.GetAllStates()
			}
		}(i)
	}
	wg.Wait()
}
