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
	} else if index < 0 {
		for i := index; i < 0; i++ {
			<-w.gc
		}
	}
}

// done 关闭一个协程
func (w *group) done() {
	<-w.gc
	w.mutex.Lock()
	w.cap--
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
				log.Println("goroutines: wait all  done")
				w.mutex.Unlock()
				return nil
			}
			w.mutex.Unlock()
		}
	}
}
