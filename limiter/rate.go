package limiter

import (
	"go.uber.org/ratelimit"
	"golang.org/x/time/rate"
	"net/http"
)

// RateLimitMiddleware 在HTTP中间件中使用
func RateLimitMiddleware(next http.Handler, limitPerSecond, size int) http.Handler {
	// 初始化：每秒10个令牌，桶容量30
	limiter := rate.NewLimiter(rate.Limit(limitPerSecond), size)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func UberRateWaitMiddleware(next http.Handler, limitPerSecond int) http.Handler {
	// 初始化：每秒10个令牌，桶容量30
	rl := ratelimit.New(limitPerSecond)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl.Take()
		next.ServeHTTP(w, r)
	})
}
func RedisRateWaitMiddleware(next http.Handler, limitPerSecond int) http.Handler {
	// 初始化：每秒10个令牌，桶容量30
	rl := ratelimit.New(limitPerSecond)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rl.Take()
		next.ServeHTTP(w, r)
	})
}
