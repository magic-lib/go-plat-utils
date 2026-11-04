package websocket

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

// DefaultMaxMessageSize 默认单条消息大小上限（1MB）
const DefaultMaxMessageSize int64 = 1 << 20

// ErrMessageTooLarge 消息超过大小上限
var ErrMessageTooLarge = errors.New("websocket message too large")

type serviceOptions struct {
	allowedOrigins []string
	maxMessageSize int64
}

// ServiceOption StartService 的可选配置
type ServiceOption func(o *serviceOptions)

// WithAllowedOrigins 额外允许的 Origin（如 https://a.example.com），传 "*" 表示允许所有来源。
// 默认只允许同源请求（无 Origin 头，或 Origin 的 host 与请求 Host 一致）。
func WithAllowedOrigins(origins ...string) ServiceOption {
	return func(o *serviceOptions) {
		o.allowedOrigins = append(o.allowedOrigins, origins...)
	}
}

// WithMaxMessageSize 单条消息大小上限（字节），<=0 时使用默认值
func WithMaxMessageSize(n int64) ServiceOption {
	return func(o *serviceOptions) {
		if n > 0 {
			o.maxMessageSize = n
		}
	}
}

// StartService 设置 WebSocket 服务端, 放置到http服务中。
// 默认只接受同源请求、单条消息不超过 DefaultMaxMessageSize，可通过 opts 调整。
func StartService(r *http.Request, w http.ResponseWriter, serviceHandler func(msgByte []byte) ([]byte, error), opts ...ServiceOption) error {
	o := &serviceOptions{maxMessageSize: DefaultMaxMessageSize}
	for _, opt := range opts {
		opt(o)
	}

	// 校验 Origin，防止跨站 WebSocket 劫持
	if !isOriginAllowed(r, o.allowedOrigins) {
		http.Error(w, "origin not allowed", http.StatusForbidden)
		return fmt.Errorf("websocket origin not allowed: %s", r.Header.Get("Origin"))
	}

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 接收客户端消息（限制大小，避免超大消息占满内存）
	msg, op, err := readServerDataLimited(conn, o.maxMessageSize)
	if err != nil {
		return err
	}

	resp, err := serviceHandler(msg)
	if err != nil {
		return err
	}

	return wsutil.WriteServerMessage(conn, op, resp)
}

// isOriginAllowed 无 Origin 头或与请求 Host 同源时放行，否则需在白名单中
func isOriginAllowed(r *http.Request, allowed []string) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if u, err := url.Parse(origin); err == nil && strings.EqualFold(u.Host, r.Host) {
		return true
	}
	for _, one := range allowed {
		if one == "*" || strings.EqualFold(one, origin) {
			return true
		}
	}
	return false
}

// readServerDataLimited 与 wsutil.ReadServerData 一致，额外限制单帧及整条消息的大小
func readServerDataLimited(rw io.ReadWriter, maxSize int64) ([]byte, ws.OpCode, error) {
	controlHandler := wsutil.ControlFrameHandler(rw, ws.StateServerSide)
	rd := wsutil.Reader{
		Source:         rw,
		State:          ws.StateServerSide,
		CheckUTF8:      true,
		MaxFrameSize:   maxSize,
		OnIntermediate: controlHandler,
	}
	for {
		hdr, err := rd.NextFrame()
		if err != nil {
			return nil, 0, err
		}
		if hdr.OpCode.IsControl() {
			if err = controlHandler(hdr, &rd); err != nil {
				return nil, 0, err
			}
			continue
		}
		// 分片消息由多个帧组成，再用 LimitReader 限制整条消息的总大小
		bts, err := io.ReadAll(io.LimitReader(&rd, maxSize+1))
		if err != nil {
			return nil, 0, err
		}
		if int64(len(bts)) > maxSize {
			return nil, 0, ErrMessageTooLarge
		}
		return bts, hdr.OpCode, nil
	}
}
