package logs_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/magic-lib/go-plat-utils/logs"
)

// logger 传到新协程中使用（新协程未缓存 ctx）时不能 panic
func TestCtxLoggerInNewGoroutine(t *testing.T) {
	var written atomic.Int32
	logger, _ := logs.NewCtxLogger(context.Background(), logs.DEBUG, func(ctx context.Context, logInfo *logs.LogData) {
		written.Add(1)
	})
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		logger.Error("in new goroutine")
	}()
	if r := <-done; r != nil {
		t.Fatalf("新协程中打日志 panic: %v", r)
	}
	if written.Load() == 0 {
		t.Fatal("日志未写出")
	}
}
