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

	// 发送消息到服务器
	if err = wsutil.WriteClientMessage(conn, ws.OpText, msgByte); err != nil {
		return nil, err
	}

	// 接收服务器的回复
	retMsg, _, err := wsutil.ReadClientData(conn)
	if err != nil {
		return nil, err
	}

	return retMsg, nil
}
