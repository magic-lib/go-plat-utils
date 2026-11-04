package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
)

// 服务端不回复时，PostMessage 应在 ctx 超时后返回，而不是永久阻塞
func TestPostMessageRespectsContext(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, _, _, err := ws.UpgradeHTTP(r, w)
		if err != nil {
			return
		}
		defer conn.Close()
		<-block
	}))
	defer srv.Close()
	defer close(block) // 先放行 handler，再关闭 server

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := PostMessage(ctx, "ws"+strings.TrimPrefix(srv.URL, "http"), []byte("hi"))
		done <- err
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error after context timeout")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("PostMessage blocked after context timeout")
	}
}
