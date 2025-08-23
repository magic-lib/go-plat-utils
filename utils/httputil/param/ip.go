package param

import (
	"fmt"
	"net"
	"net/http"
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
