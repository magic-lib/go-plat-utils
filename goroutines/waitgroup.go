package goroutines

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type group struct {
	gc    chan bool
	tk    *time.Ticker
	cap   int
	mutex sync.Mutex
}

// newWaitGroup 新建一个等待实例
func newWaitGroup(timeout time.Duration) *group {
	return &group{
		gc:  make(chan bool),
		cap: 0,
		tk:  time.NewTicker(timeout),
	}
}

// add 新增N个协程
func (w *group) add(index int) {
	if index == 0 {
		return
	}
	w.mutex.Lock()
	w.cap = w.cap + index
	w.mutex.Unlock()

	if index > 0 {
		GoAsync(func(params ...any) {
			for i := 0; i < index; i++ {
				w.gc <- true
			}
		})
	} else {
		for i := index; i < 0; i++ {
			<-w.gc
		}
	}
}

// done 关闭一个协程
func (w *group) done() {
	w.doneN(1)
}

// done 关闭N个协程
func (w *group) doneN(n int) {
	if n <= 0 {
		return
	}
	// 先消费 channel 中的信号
	for i := 0; i < n; i++ {
		<-w.gc
	}
	w.mutex.Lock()
	w.cap -= n
	if w.cap < 0 {
		w.cap = 0 // 防止负值
	}
	w.mutex.Unlock()
}

// wait 等待
func (w *group) wait() error {
	defer w.tk.Stop()
	for {
		select {
		case <-w.tk.C:
			log.Println("goroutines wait: timeout exec over")
			return fmt.Errorf("timeout exec over")
		default:
			w.mutex.Lock()
			if w.cap <= 0 {
				w.mutex.Unlock()
				return nil
			}
			w.mutex.Unlock()
		}
	}
}
