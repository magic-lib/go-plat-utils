package goroutines

import (
	"context"
	"fmt"
	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/conc/stream"
)

func test() {
	var wg conc.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Go(func() {
			fmt.Println("goroutines test:", i)
		})
	}
	// 自动捕获 panic，Wait 时重新抛出
	wg.Wait()
}

func test2() {
	var wg conc.WaitGroup
	wg.Go(func() {
		panic("something went wrong")
	})
	wg.Wait() // 这里会重新 panic，带完整堆栈
}

func test3() {
	// 创建池
	p := pool.New()

	// 提交任务
	for i := 0; i < 100; i++ {
		p.Go(func() {
			fmt.Println("aaaaa")
		})
	}

	// 等待所有任务完成
	p.Wait()
}

func test4() {
	p := pool.New().WithMaxGoroutines(10)

	for i := 0; i < 1000; i++ {
		p.Go(func() {
			// 最多同时运行 10 个
			fmt.Println("aaaaa")
		})
	}

	p.Wait()
}

func test5() {
	p := pool.NewWithResults[int]().
		WithMaxGoroutines(4)

	inputs := []int{1, 2, 3, 4, 5, 6, 7, 8}

	for _, v := range inputs {
		v := v
		p.Go(func() int {
			return v * v
		})
	}

	results := p.Wait()
	// results: [1 4 9 16 25 36 49 64]（顺序不确定）
	fmt.Println(results)
}

func test6() {
	p := pool.New().WithErrors().
		WithMaxGoroutines(4)

	for i := 0; i < 10; i++ {
		i := i
		p.Go(func() error {
			if i == 5 {
				return fmt.Errorf("task %d failed", i)
			}
			return nil
		})
	}

	err := p.Wait()
	// err 是所有错误的聚合
	fmt.Println(err)
}

func test7() {
	ctx := context.Background()

	p := pool.New().WithContext(ctx).
		WithMaxGoroutines(4)

	for i := 0; i < 10; i++ {
		p.Go(func(ctx context.Context) error {
			select {
			case <-ctx.Done():
				return ctx.Err() // 其他任务失败了，被取消
			default:
				return func() error {
					return nil
				}()
			}
		})
	}

	err := p.Wait()
	fmt.Println(err)
}

func mapStream(in, out chan int, f func(int) int) {
	s := stream.New().WithMaxGoroutines(10)

	for elem := range in {
		elem := elem
		s.Go(func() stream.Callback {
			result := f(elem)
			// 这个回调会串行执行，保序
			return func() { out <- result }
		})
	}

	s.Wait()
}
