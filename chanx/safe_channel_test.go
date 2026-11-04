package chanx_test

import (
	"sync"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/chanx"
)

// 缓冲区已满且无消费者时，Close 不能被阻塞中的 Send 卡住
func TestSafeChannelCloseWhileSendBlocked(t *testing.T) {
	sc := chanx.NewSafeChannel[int](1)
	if !sc.Send(1) {
		t.Fatal("第一次 Send 应成功")
	}
	sendResult := make(chan bool, 1)
	go func() { sendResult <- sc.Send(2) }() // 缓冲区已满，阻塞
	time.Sleep(50 * time.Millisecond)

	closed := make(chan struct{})
	go func() {
		sc.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("Close 死锁")
	}
	if ok := <-sendResult; ok {
		t.Fatal("关闭后阻塞中的 Send 应返回 false")
	}
	if sc.Send(3) {
		t.Fatal("关闭后 Send 应返回 false")
	}
	// 已发送成功的数据仍可读出，之后 channel 关闭
	if v, ok := <-sc.Chan(); !ok || v != 1 {
		t.Fatalf("读取 = %d, %v; want 1, true", v, ok)
	}
	if _, ok := <-sc.Chan(); ok {
		t.Fatal("channel 应已关闭")
	}
}

// 并发 Send 与 Close 不能 panic
func TestSafeChannelConcurrentSendClose(t *testing.T) {
	for round := 0; round < 50; round++ {
		sc := chanx.NewSafeChannel[int](2)
		var wg sync.WaitGroup
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func(v int) {
				defer wg.Done()
				for j := 0; j < 20; j++ {
					sc.Send(v)
				}
			}(i)
		}
		go func() {
			for range sc.Chan() {
			}
		}()
		sc.Close()
		wg.Wait()
	}
}
