package goroutines

import (
	"context"
	"errors"
	"fmt"
	"github.com/magic-lib/go-plat-utils/internal"
	"github.com/panjf2000/ants/v2"
	"github.com/timandy/routine"
	"go.uber.org/automaxprocs/maxprocs"
	"log"
	"runtime"
	"strings"
	"sync"
	"time"
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
	if task == nil {
		return
	}
	taskFun := routine.WrapTask(func() {
		GoSync(task, params...)
	})
	defaultAsyncObj.poolMutex.RLock()
	defer defaultAsyncObj.poolMutex.RUnlock()

	pool := defaultAsyncObj.antsPool
	var submitErr error
	if pool != nil {
		submitErr = pool.Submit(taskFun.Run)
	}
	if pool == nil || submitErr != nil {
		go taskFun.Run()
	}
}

type asyncResult[M any] struct {
	data M
	err  error
}

// GoAsyncTimeout 执行一个方法带过期时间
//   - ctx：父 context，传 nil 时等价于 context.Background()；
//   - timeout：>0 时同步等待结果，最多等 timeout；<=0 时纯异步，立即返回零值和 nil；
//   - fun：真正执行的任务，它的 ctx 入参已带上超时/取消能力，任务应自行响应；
//   - paramsOut：传给 fun 的可变参数。
//
// 注意两点：
//  1. 任务一定会被执行（协程池提交失败时退化成裸 goroutine）；
//  2. 超时或父 ctx 取消只是让调用方不再等待，已经跑起来的任务无法被强制终止。
func GoAsyncTimeout[T any](ctx context.Context, timeout time.Duration, fun func(ctx context.Context, paramsIn ...any) (T, error), paramsOut ...any) (t T, e error) {
	if fun == nil {
		return t, nil
	}
	resultChan := make(chan asyncResult[T], 1)

	var taskCtx context.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		taskCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	} else {
		taskCtx = context.WithoutCancel(ctx)
		// 如果超时时间为0，则表示没有超时时间，纯异步：任务生命周期不受父 ctx 影响
	}

	// 启动一个 goroutine 来执行耗时操作
	GoAsync(func(paramsInIn ...any) {
		defer func() {
			// 任务 panic 时也要回填结果，否则调用方只能傻等到超时，还拿不到 panic 原因
			if r := recover(); r != nil {
				resultChan <- asyncResult[T]{err: fmt.Errorf("GoAsyncTimeout panic: %v", r)}
				panic(r) // 继续抛给 GoSync 统一处理（打日志 / 回调 panicHandle）
			}
		}()
		oneRet, oneErr := fun(taskCtx, paramsInIn...)
		resultChan <- asyncResult[T]{data: oneRet, err: oneErr}
	}, paramsOut...)

	if timeout <= 0 {
		return t, nil // 直接异步，不等待结果
	}

	// 等待结果、超时或父 ctx 取消；用 taskCtx.Done() 而不是 time.After，
	// 既能少建一个 timer，也能在父 ctx 取消时立刻返回
	select {
	case res := <-resultChan:
		return res.data, res.err
	case <-taskCtx.Done():
		if errors.Is(taskCtx.Err(), context.DeadlineExceeded) {
			return t, fmt.Errorf("GoAsyncTimeout timeout after %v: %w", timeout, context.DeadlineExceeded)
		}
		return t, fmt.Errorf("GoAsyncTimeout canceled: %w", taskCtx.Err())
	}
}

func init() {
	_, _ = maxprocs.Set() // 手动设置，避免docker中会一直打印错误日志：maxprocs: Leaving GOMAXPROCS=4: CPU quota undefined
}
