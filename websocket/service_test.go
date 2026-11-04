package websocket_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gobwas/ws"
	"github.com/magic-lib/go-plat-utils/websocket"
)

func newServer(t *testing.T, opts ...websocket.ServiceOption) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = websocket.StartService(r, w, func(msg []byte) ([]byte, error) {
			return append([]byte("echo:"), msg...), nil
		}, opts...)
	}))
}

func wsURL(s *httptest.Server) string { return "ws" + strings.TrimPrefix(s.URL, "http") }

func TestStartServiceEcho(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	got, err := websocket.PostMessage(ctx, wsURL(srv), []byte("hi"))
	if err != nil || string(got) != "echo:hi" {
		t.Fatalf("PostMessage = %q, %v", got, err)
	}
}

func TestStartServiceMessageTooLarge(t *testing.T) {
	srv := newServer(t, websocket.WithMaxMessageSize(8))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if got, err := websocket.PostMessage(ctx, wsURL(srv), []byte("0123456789abcdef")); err == nil {
		t.Fatalf("超过大小上限应失败, got %q", got)
	}
}

func TestStartServiceCrossOrigin(t *testing.T) {
	srv := newServer(t)
	defer srv.Close()
	dialer := ws.Dialer{Header: ws.HandshakeHeaderHTTP(http.Header{"Origin": {"https://evil.example"}})}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if conn, _, _, err := dialer.Dial(ctx, wsURL(srv)); err == nil {
		_ = conn.Close()
		t.Fatal("跨域请求应被拒绝")
	}

	allowSrv := newServer(t, websocket.WithAllowedOrigins("https://evil.example"))
	defer allowSrv.Close()
	conn, _, _, err := dialer.Dial(ctx, wsURL(allowSrv))
	if err != nil {
		t.Fatalf("白名单内的 Origin 应允许: %v", err)
	}
	_ = conn.Close()
}
