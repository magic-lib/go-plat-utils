package sign

import (
	"context"
	"testing"
	"time"
)

// nonce 保存时间不能短于时间戳有效窗口（2 倍 TimestampTimeout），否则过期后可重放
func TestNonceTimeoutCoversTimestampWindow(t *testing.T) {
	getToken := func(ctx context.Context, appId string) (string, error) { return "t", nil }
	cases := []struct {
		in        *BaseSignCheck
		wantNonce time.Duration
	}{
		{&BaseSignCheck{GetAppTokenFunc: getToken}, 10 * time.Minute},
		{&BaseSignCheck{GetAppTokenFunc: getToken, NonceTimeout: time.Minute}, 10 * time.Minute},
		{&BaseSignCheck{GetAppTokenFunc: getToken, TimestampTimeout: time.Minute, NonceTimeout: 30 * time.Minute}, 30 * time.Minute},
	}
	for i, c := range cases {
		svc, err := New(c.in)
		if err != nil {
			t.Fatal(err)
		}
		if got := svc.(*BaseSignCheck).NonceTimeout; got != c.wantNonce {
			t.Errorf("case %d: NonceTimeout = %v, want %v", i, got, c.wantNonce)
		}
	}
}
