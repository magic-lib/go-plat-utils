package crontab_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/crontab"
	"github.com/robfig/cron/v3"
)

// 任务 panic 时不能让进程崩溃，也不能影响其他任务
func TestStartJobsPanicRecover(t *testing.T) {
	var okRuns, panicRuns atomic.Int32
	err := crontab.StartJobs(true, map[string]cron.FuncJob{
		"@every 1s": func() {
			panicRuns.Add(1)
			panic("job panic")
		},
	}, map[string]cron.FuncJob{
		"@every 2s": func() { okRuns.Add(1) },
	})
	if err != nil {
		t.Fatalf("StartJobs err: %v", err)
	}
	defer crontab.Stop()

	if okRuns.Load() != 1 || panicRuns.Load() != 1 {
		t.Fatalf("立即执行次数 ok=%d panic=%d, want 1 1", okRuns.Load(), panicRuns.Load())
	}
	// 等调度器至少触发一次会 panic 的任务
	time.Sleep(1500 * time.Millisecond)
	if panicRuns.Load() < 2 {
		t.Fatalf("调度器未按时触发任务, panicRuns=%d", panicRuns.Load())
	}
}
