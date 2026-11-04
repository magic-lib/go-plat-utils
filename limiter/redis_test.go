package limiter_test

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
	"github.com/magic-lib/go-plat-utils/limiter"
)

func TestRedisLimiterAllow(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer func() { _ = client.Close() }()

	rl := limiter.NewRedisLimiter(client)
	ctx := context.Background()
	const key, limit, window = "limit:test", 5, 300 * time.Millisecond

	// 同一毫秒内的多次请求都要计数，前 limit 次放行，之后拒绝
	for i := 0; i < limit+2; i++ {
		ok, err := rl.Allow(ctx, key, limit, window)
		if err != nil {
			t.Fatalf("Allow err: %v", err)
		}
		if want := i < limit; ok != want {
			t.Fatalf("第 %d 次请求 Allow = %v, want %v", i+1, ok, want)
		}
	}
	if n, _ := client.ZCard(ctx, key).Result(); n != limit {
		t.Fatalf("窗口内计数 = %d, want %d", n, limit)
	}

	// 过了窗口时间后应重新放行（修复前窗口被放大 1000 倍，会一直拒绝）
	time.Sleep(window + 50*time.Millisecond)
	ok, err := rl.Allow(ctx, key, limit, window)
	if err != nil || !ok {
		t.Fatalf("窗口过后 Allow = %v, %v; want true", ok, err)
	}
}
