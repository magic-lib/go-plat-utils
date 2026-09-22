package goroutines_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/goroutines"
)

func TestGoAsyncTimeoutNormal(t *testing.T) {
	got, err := goroutines.GoAsyncTimeout(context.Background(), time.Second,
		func(ctx context.Context, paramsIn ...any) (string, error) {
			return fmt.Sprintf("%v-%v", paramsIn[0], paramsIn[1]), nil
		}, "a", 1)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got != "a-1" {
		t.Fatalf("got = %q, want %q", got, "a-1")
	}
}

// ctx 传 nil 时不应 panic，等价于 context.Background()
func TestGoAsyncTimeoutNilCtx(t *testing.T) {
	got, err := goroutines.GoAsyncTimeout[int](nil, time.Second,
		func(ctx context.Context, paramsIn ...any) (int, error) {
			return 7, nil
		})
	if err != nil || got != 7 {
		t.Fatalf("got = %v, err = %v, want 7, nil", got, err)
	}
}

// 超时必须能被 errors.Is 判定为 context.DeadlineExceeded
func TestGoAsyncTimeoutTimeout(t *testing.T) {
	start := time.Now()
	_, err := goroutines.GoAsyncTimeout[string](context.Background(), 100*time.Millisecond,
		func(ctx context.Context, paramsIn ...any) (string, error) {
			<-ctx.Done() // 模拟响应 ctx 的长任务
			return "", ctx.Err()
		})
	if err == nil {
		t.Fatal("want timeout error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want wrap DeadlineExceeded", err)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("elapsed = %v, want ~100ms", d)
	}
}

// 任务 panic 时必须立刻拿到错误，而不是被 GoSync 吞掉后白等到超时
func TestGoAsyncTimeoutPanic(t *testing.T) {
	start := time.Now()
	_, err := goroutines.GoAsyncTimeout[string](context.Background(), 5*time.Second,
		func(ctx context.Context, paramsIn ...any) (string, error) {
			panic("boom")
		})
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("want panic error, got nil")
	}
	if !strings.Contains(err.Error(), "boom") {
		t.Fatalf("err = %v, want contain boom", err)
	}
	// 修复前这里会一直卡到 5s 超时
	if elapsed > time.Second {
		t.Fatalf("elapsed = %v, want < 1s (不应等满超时)", elapsed)
	}
}

// 父 ctx 取消后应立即返回，而不是死等到 timeout 到期
func TestGoAsyncTimeoutParentCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, err := goroutines.GoAsyncTimeout[string](ctx, 5*time.Second,
		func(ctxIn context.Context, paramsIn ...any) (string, error) {
			<-ctxIn.Done()
			return "", ctxIn.Err()
		})
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("want cancel error, got nil")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want wrap context.Canceled", err)
	}
	if elapsed > time.Second {
		t.Fatalf("elapsed = %v, want ~50ms", elapsed)
	}
}

// timeout <= 0 表示纯异步：立即返回零值，但任务仍然会被执行
func TestGoAsyncTimeoutNoTimeout(t *testing.T) {
	done := make(chan struct{})
	start := time.Now()
	got, err := goroutines.GoAsyncTimeout[string](context.Background(), 0,
		func(ctx context.Context, paramsIn ...any) (string, error) {
			close(done)
			return "ignored", nil
		})
	if got != "" || err != nil {
		t.Fatalf("got = %q, err = %v, want empty, nil", got, err)
	}
	if d := time.Since(start); d > 200*time.Millisecond {
		t.Fatalf("elapsed = %v, want immediate return", d)
	}
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("纯异步任务没有被执行")
	}
}
