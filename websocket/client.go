package websocket

import (
	"context"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	//github.com/gorilla/websocket
)

// PostMessage 启动客户端
func PostMessage(ctx context.Context, wsUrl string, msgByte []byte) ([]byte, error) {
	// 连接到 WebSocket 服务器 "ws://localhost:8080"
	conn, _, _, err := ws.Dial(ctx, wsUrl)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// ws.Dial 只在建连阶段使用 ctx，读写阶段需要自己处理：有截止时间就设置到连接上，
	// ctx 被取消时关闭连接，避免服务端不回复时永久阻塞
	if deadline, ok := ctx.Deadline(); ok {
		if err = conn.SetDeadline(deadline); err != nil {
			return nil, err
		}
	}
	stop := context.AfterFunc(ctx, func() {
		_ = conn.Close()
	})
	defer stop()

	// 发送消息到服务器
	if err = wsutil.WriteClientMessage(conn, ws.OpText, msgByte); err != nil {
		return nil, err
	}

	// 接收服务器的回复：客户端读取服务端数据应使用 ReadServerData（原来用 ReadClientData，按服务端视角要求帧带掩码，必然失败）
	retMsg, _, err := wsutil.ReadServerData(conn)
	if err != nil {
		return nil, err
	}

	return retMsg, nil
}
