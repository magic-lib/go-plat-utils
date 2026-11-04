package interceptor_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// panic 时返回给客户端的错误不包含堆栈
func TestRecoveryHidesStack(t *testing.T) {
	recovery := interceptor.NewGrpcServer().RecoveryUnaryServerInterceptor()
	_, err := recovery(context.Background(), nil, &grpc.UnaryServerInfo{FullMethod: "/svc/M"},
		func(ctx context.Context, req any) (any, error) { panic("boom") })
	st, _ := status.FromError(err)
	if st.Code() != codes.Internal || !strings.Contains(st.Message(), "boom") {
		t.Fatalf("err = %v", err)
	}
	if strings.Contains(st.Message(), "goroutine") || strings.Contains(st.Message(), ".go:") {
		t.Fatalf("返回给客户端的错误不应包含堆栈: %s", st.Message())
	}
}

func TestRetryRespectsContext(t *testing.T) {
	retry := interceptor.NewGrpcClient().RetryUnaryClientInterceptor(3, 200*time.Millisecond)
	calls := 0
	unavailable := func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		calls++
		return status.Error(codes.Unavailable, "down")
	}

	// 共 4 次调用，只在中间等待 3 次，最后一次失败后不再等待
	start := time.Now()
	_ = retry(context.Background(), "/svc/M", nil, nil, nil, unavailable)
	if calls != 4 {
		t.Fatalf("调用次数 = %d, want 4", calls)
	}
	if cost := time.Since(start); cost >= 750*time.Millisecond {
		t.Fatalf("最后一次失败后不应再等待, 耗时 %v", cost)
	}

	// ctx 取消后立即停止重试
	calls = 0
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	start = time.Now()
	_ = retry(ctx, "/svc/M", nil, nil, nil, unavailable)
	if cost := time.Since(start); cost >= 150*time.Millisecond || calls != 1 {
		t.Fatalf("ctx 超时后应立即返回, 耗时 %v, 调用 %d 次", cost, calls)
	}
}
