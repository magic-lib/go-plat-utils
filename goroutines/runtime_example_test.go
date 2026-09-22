package goroutines_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/goroutines"
)

// context.WithValue 的 key 必须用自定义类型，直接用 string 会触发 staticcheck SA1029
type ctxKeyFish struct{}
type ctxKeyPig struct{}

// asyncCtxResult 收集异步任务里看到的 context 内容，统一回到主 goroutine 再断言，
// 避免在 goroutine 里直接打印（顺序随机、含指针地址，结果不可比对）
type asyncCtxResult struct {
	idx  int
	fish any
	pId  string
	ok   bool // 是否取到了 context
}

// TestGoAsyncExample 演示 goroutine-local context 在 GoAsync 中的传递。
//
// 原本写成 ExampleGoAsync：Example 函数只要带 `// Output:` 注释，go test 就会把
// 该函数运行期间写到 stdout 的全部内容与期望串比对。而这里的输出既依赖 goroutine
// 调度顺序（i=8/i=1/i=3…），又打印了 context 指针地址（每次不同），根本无法写死，
// 所以改成普通 Test：不看 stdout，改为收集结果后统一断言。
func TestGoAsyncExample(t *testing.T) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, ctxKeyFish{}, "章鱼")

	goroutines.SetContext(&ctx)
	defer goroutines.DelContext()

	goroutines.OpenRoutinePool(9)
	defer goroutines.CloseRoutinePool()

	_, pId, _ := goroutines.GetContext()
	t.Logf("start: %s", pId)

	const total = 10
	resultChan := make(chan asyncCtxResult, total)

	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		goroutines.GoAsync(func(params ...any) {
			defer wg.Done()
			idx := params[0].(int)
			c, pidIn, _ := goroutines.GetContext()
			if c == nil {
				// 原实现在这里无条件解引用 *c，c 为 nil 时会 panic
				resultChan <- asyncCtxResult{idx: idx}
				return
			}
			resultChan <- asyncCtxResult{idx: idx, fish: (*c).Value(ctxKeyFish{}), pId: pidIn, ok: true}
		}, i)
	}

	// GoAsync 在协程池提交失败（池满/已关闭）时会静默丢弃任务，
	// 因此不能无条件 wg.Wait()，必须有超时兜底，否则测试会永久挂住
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("异步任务未在 5s 内全部执行完（可能被协程池丢弃）")
	}
	close(resultChan)

	got := 0
	for r := range resultChan {
		got++
		if !r.ok {
			t.Errorf("i=%d 未取到 context", r.idx)
			continue
		}
		if r.fish != "章鱼" {
			t.Errorf("i=%d fish = %v, want 章鱼", r.idx, r.fish)
		}
	}
	if got != total {
		t.Fatalf("执行了 %d 个任务, want %d", got, total)
	}

	// 主 goroutine 里追加一个 pig，再起的异步任务应能同时看到 pig 和 fish
	c2, pId2, _ := goroutines.GetContext()
	if c2 == nil {
		t.Fatal("主 goroutine 取不到 context")
	}
	if v := (*c2).Value(ctxKeyFish{}); v != "章鱼" {
		t.Fatalf("主 goroutine fish = %v, want 章鱼", v)
	}

	c2New := context.WithValue(*c2, ctxKeyPig{}, "猪")
	goroutines.SetContext(&c2New)
	t.Logf("pig2: %v %v %s", c2New.Value(ctxKeyPig{}), c2New.Value(ctxKeyFish{}), pId2)

	pigChan := make(chan string, 1)
	goroutines.GoAsync(func(params ...any) {
		c3, pId3, _ := goroutines.GetContext()
		if c3 == nil {
			pigChan <- "ctx=nil"
			return
		}
		pigChan <- fmt.Sprintf("pig=%v fish=%v pId=%s",
			(*c3).Value(ctxKeyPig{}), (*c3).Value(ctxKeyFish{}), pId3)
	})

	select {
	case r := <-pigChan:
		t.Logf("pig3: %s", r)
	case <-time.After(3 * time.Second):
		t.Fatal("第二个异步任务未执行（可能被协程池丢弃）")
	}
}

func TestExampleGoSync(t *testing.T) {
	start := time.Now()
	ctx := context.Background()
	_, _ = goroutines.GoAsyncTimeout(ctx, 3*time.Second, func(ctx context.Context, params ...any) (bool, error) {
		time.Sleep(2 * time.Second)
		fmt.Println("test")
		return true, nil

	})

	fmt.Println(time.Since(start).Seconds())

	time.Sleep(3 * time.Second)

}
