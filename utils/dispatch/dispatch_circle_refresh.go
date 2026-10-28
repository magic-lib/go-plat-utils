package dispatch

import (
	"context"
	"log"
	"sort"
	"time"
)

// fetch 执行一次拉取，并在配置了 FetchTimeout 时套上超时，避免慢查询无限期阻塞刷新循环。
func (d *Dispatcher) fetch(ctx context.Context) ([]Agent, error) {
	if d.Fetch == nil {
		return nil, nil
	}
	if d.FetchTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.FetchTimeout)
		defer cancel()
	}
	return d.Fetch(ctx)
}

// start 启动定时刷新：等待 period 后拉取一次最新列表并刷新环形队列，如此循环。
// 语义是“上次刷新结束后再等 period”（而非固定频率），因此慢查询时不会背靠背补跑、放大下游压力；
// 刷新在单个常驻 goroutine 里串行执行，不会与自身重叠，也不会出现刷新任务堆积。
//
// 注意：这里用裸 goroutine 而不是 goroutines.GoAsync——常驻的 for-select 循环会永久占用一个池 worker。
func (d *Dispatcher) start(ctx context.Context) {
	go func() {
		timer := time.NewTimer(d.Period)
		defer timer.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-timer.C:
				err := d.Refresh(ctx)
				if err != nil {
					log.Println("[dispatch] 刷新客服列表失败:", err)
				}
				timer.Reset(d.Period)
			}
		}
	}()
}

// Refresh 立即拉取一次最新列表并刷新环形队列（人员变动时手动触发）。
// 与定时刷新共用同一把 refreshMu，不会并发 fetch；Fetch 为 nil 时直接返回 nil。
func (d *Dispatcher) Refresh(ctx context.Context) error {
	if d.Fetch == nil {
		return nil
	}
	d.refreshMu.Lock()
	defer d.refreshMu.Unlock()
	ctx = ctxOr(ctx)
	agents, err := d.fetch(ctx)
	if err != nil {
		return err
	}
	d.apply(agents)
	return nil
}

func (d *Dispatcher) init(ctx context.Context) error {
	if d.Fetch == nil {
		// 没有拉取函数时，用配置里直接给出的列表初始化
		if len(d.Agents) > 0 {
			d.apply(d.Agents)
		}
		return nil
	}

	return d.Refresh(ctx)
}

// apply 用最新列表刷新环形队列：过滤 nil、按 GetAgentId() 升序排序，不改变 cursor。
// 由于 cursor 记录的是（ID, 实例序号），列表增删后下一次 Assign() 仍能正确取到下一个。
func (d *Dispatcher) apply(fetched []Agent) {
	// 复制后再排序，避免修改外部切片；顺带过滤掉 nil，防止分配时 panic
	online := make([]Agent, 0, len(fetched))
	for _, a := range fetched {
		if a == nil {
			continue
		}
		online = append(online, a)
	}
	sort.Slice(online, func(i, j int) bool {
		return online[i].GetAgentId() < online[j].GetAgentId()
	})

	d.mu.Lock()
	d.agents = online
	d.mu.Unlock()
}
