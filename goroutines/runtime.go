package goroutines

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/internal"
	"github.com/panjf2000/ants/v2"
	"github.com/timandy/routine"
	_ "go.uber.org/automaxprocs"
	"log"
	"runtime"
	"strings"
	"sync"
)

type asyncObj struct {
	poolMutex   sync.RWMutex
	panicMutex  sync.RWMutex
	antsPool    *ants.Pool
	panicHandle func(err error, retRecover any) //全局Panic后的处理方法
	oncePool    *internal.Once
}

var (
	defaultAsyncObj = asyncObj{
		oncePool: &internal.Once{},
	}
)

// SetPanicHandle panic的方法
func SetPanicHandle(c func(err error, retRecover any)) {
	if c != nil {
		defaultAsyncObj.panicMutex.Lock()
		defer defaultAsyncObj.panicMutex.Unlock()
		defaultAsyncObj.panicHandle = c
	}
}

// OpenRoutinePool 启动一个全局的goroutine的协程池，只会执行一次
func OpenRoutinePool(nums int) *ants.Pool {
	if defaultAsyncObj.antsPool != nil {
		if nums > 0 {
			defaultAsyncObj.antsPool.Tune(nums)
		}
		return defaultAsyncObj.antsPool
	}

	defaultAsyncObj.poolMutex.Lock()
	defer defaultAsyncObj.poolMutex.Unlock()

	err := defaultAsyncObj.oncePool.Do(func() error {
		if nums == 0 {
			nums = ants.DefaultAntsPoolSize
		}
		newPool, err := ants.NewPool(nums)
		if err == nil {
			defaultAsyncObj.antsPool = newPool
			return nil
		}
		return err
	})
	if err != nil {
		log.Println("OpenRoutinePool error:", err)
		return nil
	}

	return defaultAsyncObj.antsPool
}

// CloseRoutinePool 不用了，关闭池
func CloseRoutinePool() {
	if defaultAsyncObj.antsPool == nil || defaultAsyncObj.antsPool.IsClosed() {
		return
	}

	defaultAsyncObj.poolMutex.Lock()
	defer defaultAsyncObj.poolMutex.Unlock()

	defaultAsyncObj.antsPool.Release()
	defaultAsyncObj.antsPool = nil
}

// GoSync 同步方法
func GoSync(task func(params ...any), params ...any) {
	defer func() {
		if err := recover(); err != nil {
			//打印调用栈信息
			buf := make([]byte, 2048)
			n := runtime.Stack(buf, false)
			stackInfo := fmt.Sprintf("%s", buf[:n])
			stackInfo = strings.ReplaceAll(stackInfo, "\n", "|")
			errFormat := "panic_stack_info: %s ### %s"
			defaultAsyncObj.panicMutex.RLock()
			if defaultAsyncObj.panicHandle != nil {
				defaultAsyncObj.panicHandle(fmt.Errorf(errFormat, err, stackInfo), err)
				defaultAsyncObj.panicMutex.RUnlock()
			} else {
				defaultAsyncObj.panicMutex.RUnlock()
				log.Println(fmt.Sprintf(errFormat, err, stackInfo))
			}
			return
		}
	}()
	task(params...)
}

// GoAsync 异步方法
func GoAsync(task func(params ...any), params ...any) {
	taskFun := routine.WrapTask(func() {
		func(newTask func(params ...any), tempParams ...interface{}) {
			GoSync(newTask, tempParams...)
		}(task, params...)
	})
	if defaultAsyncObj.antsPool == nil {
		go taskFun.Run()
		return
	}
	defaultAsyncObj.poolMutex.RLock()
	defer defaultAsyncObj.poolMutex.RUnlock()
	err := defaultAsyncObj.antsPool.Submit(taskFun.Run)
	if err != nil {
		return
	}
}
