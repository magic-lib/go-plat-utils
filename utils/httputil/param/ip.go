package param

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
)

// IPv4 得到当前机器IP地址
func IPv4() (net.IP, error) {
	addrStr, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrStr {
		ipAddr, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		if ipAddr.IsLoopback() {
			continue
		}
		if ipAddr.To4() != nil {
			return ipAddr, nil
		}
	}
	return nil, fmt.Errorf("未找到IP地址")
}

// ClientIP 得到客户端IP地址
func ClientIP(r *http.Request) string {
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

// MachineCode 获取机器的唯一实例Id，ip+pid
func MachineCode() string {
	return getLocalIP() + "/" + strconv.Itoa(os.Getpid())
}

// getLocalIP 获取本机IP:随便返回一个就行，多网卡模式(eth0/eth1)、容器化部署模式干扰因素都不需要考虑。
func getLocalIP() string {
	// 1. 获取所有网络接口
	interfaces, err := net.Interfaces()
	if err != nil {
		return ""
	}

	// 2. 遍历网卡，寻找符合条件的 IP
	for _, ipFace := range interfaces {
		// 过滤回环接口和禁用接口
		if ipFace.Flags&net.FlagLoopback != 0 || // 排除回环地址
			ipFace.Flags&net.FlagUp == 0 { // 排除未启用的接口
			continue
		}

		// 3. 获取接口的 IP 地址列表
		addrTemp, err := ipFace.Addrs()
		if err != nil {
			continue
		}

		// 4. 寻找第一个 IPv4 地址
		for _, addr := range addrTemp {
			ipNet, ok := addr.(*net.IPNet)
			if !ok {
				continue
			}

			ip := ipNet.IP.To4() // 转换为 IPv4 格式
			if ip == nil || ip.IsLoopback() {
				continue
			}

			// 返回字符串格式的 IP（如 192.168.1.100）
			return ip.String()
		}
	}

	// 5. 未找到有效 IP 时的兜底逻辑
	return ""
}

//// GetPathParam Returns a string parameter from request path or req.Attributes
//func GetPathParam(name string, req *http.Request) (param string) {
//	restfulReq := restful.NewRequest(req)
//	// Get parameter from request path
//	param = restfulReq.PathParameter(name)
//	if param != "" {
//		return param
//	}
//
//	// Get parameter from request attributes (set by intermediates)
//	attr := restfulReq.Attribute(name)
//	if attr != nil {
//		param, _ = attr.(string)
//	}
//	return
//}
//
//// GetIntParam 取得int参数
//func GetIntParam(req *http.Request, name string, def int) int {
//	restfulReq := restful.NewRequest(req)
//	num := def
//	if strNum := restfulReq.QueryParameter(name); len(strNum) > 0 {
//		num, _ = strconv.Atoi(strNum)
//	}
//	return num
//}
