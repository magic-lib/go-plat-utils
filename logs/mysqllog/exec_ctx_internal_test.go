package mysqllog

import (
	"context"
	"testing"
)

// 关闭（ctx 已取消）后最后一次写库仍需可用的 ctx，并带超时
func TestExecCtxAfterClose(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ml := &mysqlLogger{ctx: ctx, cancel: cancel}

	runCtx, runCancel := ml.execCtx()
	if runCtx != ctx {
		t.Fatal("运行中应使用 ml.ctx")
	}
	runCancel()

	cancel()
	closeCtx, closeCancel := ml.execCtx()
	defer closeCancel()
	if closeCtx.Err() != nil {
		t.Fatalf("关闭后写库的 ctx 不应已取消: %v", closeCtx.Err())
	}
	if _, ok := closeCtx.Deadline(); !ok {
		t.Fatal("关闭后写库的 ctx 应带超时")
	}
}
