package crontab

//import (
//	"context"
//	"crypto/sha1"
//	"encoding/hex"
//	"fmt"
//	"math/rand"
//	"sync"
//	"time"
//)
//
//// LeaseStore 一致性存储抽象：支持 CAS 抢租约 + 续租
//type LeaseStore interface {
//	TryAcquire(jobID, runID string, ttl time.Duration) (acquired bool)
//	Renew(jobID, runID string, ttl time.Duration) (ok bool)
//	Release(jobID, runID string)
//	Get(jobID string) (runID string, exp time.Time, ok bool)
//}
//
//// 内存版，仅演示；真实场景应换成 etcd/Spanner/带一致性语义的存储
//type memStore struct {
//	mu   sync.Mutex
//	data map[string]lease
//}
//type lease struct {
//	runID string
//	exp   time.Time
//}
//
//func newMemStore() *memStore { return &memStore{data: map[string]lease{}} }
//
//func (m *memStore) TryAcquire(jobID, runID string, ttl time.Duration) bool {
//	m.mu.Lock()
//	defer m.mu.Unlock()
//
//	now := time.Now()
//	cur, ok := m.data[jobID]
//	if ok && cur.exp.After(now) {
//		return false // 还没过期，别人持有
//	}
//	m.data[jobID] = lease{runID: runID, exp: now.Add(ttl)}
//	return true
//}
//
//func (m *memStore) Renew(jobID, runID string, ttl time.Duration) bool {
//	m.mu.Lock()
//	defer m.mu.Unlock()
//
//	now := time.Now()
//	cur, ok := m.data[jobID]
//	if !ok || cur.runID != runID || cur.exp.Before(now) {
//		return false
//	}
//	m.data[jobID] = lease{runID: runID, exp: now.Add(ttl)}
//	return true
//}
//
//func (m *memStore) Release(jobID, runID string) {
//	m.mu.Lock()
//	defer m.mu.Unlock()
//	cur, ok := m.data[jobID]
//	if ok && cur.runID == runID {
//		delete(m.data, jobID)
//	}
//}
//
//func (m *memStore) Get(jobID string) (string, time.Time, bool) {
//	m.mu.Lock()
//	defer m.mu.Unlock()
//	cur, ok := m.data[jobID]
//	return cur.runID, cur.exp, ok
//}
//
//// job 定义：为了简单，用 interval 代替 cron 表达式
//type Job struct {
//	ID       string
//	Every    time.Duration
//	Jitter   time.Duration
//	LeaseTTL time.Duration
//	Run      func(ctx context.Context) error
//}
//
//// runID 做一个可追踪的去重 token（真实场景可带上 scheduleTime）
//func makeRunID(jobID string, scheduleTime time.Time) string {
//	h := sha1.Sum([]byte(fmt.Sprintf("%s|%d", jobID, scheduleTime.UnixNano())))
//	return hex.EncodeToString(h[:8])
//}
//
//type Scheduler struct {
//	store LeaseStore
//	node  string
//}
//
//func NewScheduler(store LeaseStore, node string) *Scheduler {
//	return &Scheduler{store: store, node: node}
//}
//
//func (s *Scheduler) Start(ctx context.Context, job Job) {
//	go func() {
//		// 为了演示：对齐到“下一次 tick”
//		next := time.Now().Truncate(job.Every).Add(job.Every)
//
//		for {
//			select {
//			case <-ctx.Done():
//				return
//			case <-time.After(time.Until(next)):
//				// 抖动：防止所有 job 同一毫秒开跑
//				if job.Jitter > 0 {
//					time.Sleep(time.Duration(rand.Int63n(int64(job.Jitter))))
//				}
//
//				scheduleTime := next
//				next = next.Add(job.Every)
//
//				runID := makeRunID(job.ID, scheduleTime)
//
//				if !s.store.TryAcquire(job.ID, runID, job.LeaseTTL) {
//					// 没抢到，说明别的节点在跑，或刚跑完还在 TTL 内
//					continue
//				}
//
//				// 拿到租约了：开一个续租协程，模拟“执行者活着就续租”
//				runCtx, cancel := context.WithCancel(ctx)
//				renewDone := make(chan struct{})
//				go func() {
//					ticker := time.NewTicker(job.LeaseTTL / 3)
//					defer ticker.Stop()
//					defer close(renewDone)
//
//					for {
//						select {
//						case <-runCtx.Done():
//							return
//						case <-ticker.C:
//							if !s.store.Renew(job.ID, runID, job.LeaseTTL) {
//								// 续租失败：可能被抢占/过期/存储抖动
//								cancel()
//								return
//							}
//						}
//					}
//				}()
//
//				// 执行 + 简单重试（指数退避），重点是：业务侧仍要幂等
//				err := s.runWithRetry(runCtx, job, 3)
//
//				cancel()
//				<-renewDone
//				s.store.Release(job.ID, runID)
//
//				if err != nil {
//					fmt.Printf("[%s] job=%s run=%s FAIL: %v\n", s.node, job.ID, runID, err)
//				} else {
//					fmt.Printf("[%s] job=%s run=%s OK\n", s.node, job.ID, runID)
//				}
//			}
//		}
//	}()
//}
//
//func (s *Scheduler) runWithRetry(ctx context.Context, job Job, max int) error {
//	backoff := 200 * time.Millisecond
//	for i := 0; i <= max; i++ {
//		if err := job.Run(ctx); err == nil {
//			return nil
//		} else if i == max {
//			return err
//		}
//		select {
//		case <-ctx.Done():
//			return ctx.Err()
//		case <-time.After(backoff):
//			backoff *= 2
//		}
//	}
//	return nil
//}

//
//func main() {
//	rand.Seed(time.Now().UnixNano())
//
//	store := newMemStore()
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel()
//
//	job := Job{
//		ID:       "daily-billing",
//		Every:    2 * time.Second,
//		Jitter:   300 * time.Millisecond,
//		LeaseTTL: 2 * time.Second,
//		Run: func(ctx context.Context) error {
//			// 模拟：偶发失败 + 模拟耗时
//			time.Sleep(400 * time.Millisecond)
//			if rand.Intn(5) == 0 {
//				return fmt.Errorf("downstream 503")
//			}
//			return nil
//		},
//	}
//
//	// 模拟两台“机房节点”同时跑同一个 job
//	NewScheduler(store, "us-central1-a").Start(ctx, job)
//	NewScheduler(store, "europe-west1-b").Start(ctx, job)
//
//	time.Sleep(12 * time.Second)
//}
