package httputil

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/id-generator/id"
	"github.com/magic-lib/go-plat-utils/utils"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

// GetDomainHost 取得域名或上一级域名，leftNum为剩下的域名段，最高级为qq.com
func GetDomainHost(request *http.Request, leftNum int) string {
	if leftNum <= 2 {
		leftNum = 2
	}
	newHost := request.Header.Get("User-Host")
	if newHost == "" {
		newHost = request.Host
	}
	hostList := strings.Split(newHost, ".")
	newHostList := make([]string, 0)

	for _, one := range hostList {
		if one != "" {
			newHostList = append(newHostList, one)
		}
	}

	if len(newHostList) > leftNum {
		startNum := len(newHostList) - leftNum
		newHostList = newHostList[startNum:]
	}

	newHost = strings.Join(newHostList, ".")
	return newHost
}

// Ping 是否相通
func Ping(host string, port string, timeout time.Duration) error {
	if timeout == 0 {
		timeout = 2 * time.Second
	}
	address := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		log.Println("ERR:" + host + ">" + err.Error())
		return err
	}
	if conn != nil {
		_ = conn.Close()
		return nil
	}
	return fmt.Errorf("connect failed")
}

// GetLogId 生成唯一日志id
func GetLogId() string {
	logId := id.NewUUID()
	randomStr := fmt.Sprintf("%s%s", logId, utils.RandomString(12))
	newLogId := id.GetUUID(randomStr)
	logIdFront := newLogId[0:24]

	logIdEnd := conv.FormatFromUnixTime(conv.ShortTimeForm12)

	return strings.ToLower(fmt.Sprintf("%s%s", logIdFront, logIdEnd))
}

// GetAvailablePort 获取可用端口号
func GetAvailablePort() (int, error) {
	// 创建临时 TCP 监听器，端口指定为 0（系统自动分配可用端口）
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, fmt.Errorf("创建监听器失败: %w", err)
	}
	defer func() {
		_ = listener.Close() // 关闭监听器，释放端口
	}()

	// 获取监听器实际绑定的地址
	addr := listener.Addr().(*net.TCPAddr)
	return addr.Port, nil
}
