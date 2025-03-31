package websocket

import (
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"net/http"
)

// StartService 设置 WebSocket 服务端, 放置到http服务中
func StartService(r *http.Request, w http.ResponseWriter, serviceHandler func(msgByte []byte) ([]byte, error)) error {
	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		return err
	}
	defer conn.Close()

	// 接收客户端消息
	msg, op, err := wsutil.ReadServerData(conn)
	if err != nil {
		return err
	}

	resp, err := serviceHandler(msg)
	if err != nil {
		return err
	}

	return wsutil.WriteServerMessage(conn, op, resp)
}
