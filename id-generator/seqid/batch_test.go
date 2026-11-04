package seqid_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-cache/cache"
	"github.com/magic-lib/go-plat-locker/lock"
	"github.com/magic-lib/go-plat-utils/id-generator/seqid"
)

// tokenLocker 模拟带 token 的锁：只有加锁的同一个实例才能解锁
type tokenLocker struct {
	locked bool
	stats  *lockStats
}

type lockStats struct {
	created, locks, badUnlocks int
}

func (l *tokenLocker) Lock(context.Context) error {
	l.locked = true
	l.stats.locks++
	return nil
}
func (l *tokenLocker) UnLock(context.Context) error {
	if !l.locked {
		l.stats.badUnlocks++
		return errors.New("unlock by another instance")
	}
	l.locked = false
	return nil
}
func (l *tokenLocker) TryLock(ctx context.Context) (bool, error) { return true, l.Lock(ctx) }
func (l *tokenLocker) TryLockFunc(ctx context.Context, f func()) (bool, error) {
	return true, l.LockFunc(ctx, f)
}
func (l *tokenLocker) LockFunc(ctx context.Context, f func()) error {
	_ = l.Lock(ctx)
	f()
	return l.UnLock(ctx)
}
func (l *tokenLocker) LockerType() string { return "token" }

func TestIDGeneratorBatchAndLocker(t *testing.T) {
	stats := &lockStats{}
	gen, err := seqid.NewIDGenerator[int64](&seqid.IDGenerator[int64]{
		Namespace: "test",
		IdStore:   cache.NewMemGoCache[int64](time.Hour, time.Hour),
		LockFunc: func(ns, key string) lock.Locker {
			stats.created++
			return &tokenLocker{stats: stats}
		},
	})
	if err != nil {
		t.Fatalf("NewIDGenerator: %v", err)
	}
	if gen.BatchSize != 100 {
		t.Fatalf("默认 BatchSize = %d, want 100", gen.BatchSize)
	}
	for want := int64(1); want <= 250; want++ {
		got, err := gen.Generate("order")
		if err != nil || got != want {
			t.Fatalf("第 %d 次 Generate = %d, %v; want %d", want, got, err, want)
		}
	}
	// 250 个 id 只需取 3 批，每批加锁一次，且解锁都由加锁的同一实例完成
	if stats.locks != 3 || stats.badUnlocks != 0 || stats.created != 3 {
		t.Fatalf("加锁 %d 次、创建锁 %d 个、错误解锁 %d 次; want 3, 3, 0", stats.locks, stats.created, stats.badUnlocks)
	}
}
