package utils

import (
	"sync"
	"testing"
	"time"
)

func TestConsistentHashIncreaseReplicas(t *testing.T) {
	c := newConsistentHash(2)
	c.AddNode("a")
	c.AddNode("b")
	c.AddNode("a") // 重复添加不生效

	done := make(chan struct{})
	go func() {
		c.IncreaseReplicas(4)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("IncreaseReplicas 死锁")
	}
	if len(c.nodes) != 8 || c.virtualNodeMultiplier != 4 {
		t.Fatalf("虚拟节点数 = %d, 倍数 = %d; want 8, 4", len(c.nodes), c.virtualNodeMultiplier)
	}
	if node := c.GetNode("key"); node != "a" && node != "b" {
		t.Fatalf("GetNode = %q", node)
	}
}

// 并发添加节点，配合 -race 检查数据竞争
func TestConsistentHashConcurrentAdd(t *testing.T) {
	c := newConsistentHash(3)
	var wg sync.WaitGroup
	for _, node := range []string{"a", "b", "c", "a", "b", "c"} {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			c.AddNode(n)
			_ = c.GetNode("key-" + n)
		}(node)
	}
	wg.Wait()
	if len(c.nodes) != 9 {
		t.Fatalf("虚拟节点数 = %d, want 9", len(c.nodes))
	}
}
