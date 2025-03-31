// Package conv 转换方法
package conv

import (
	"net"
	"strconv"
	"strings"
)

// UInt32ToIP 将uint32类型转化为ipv4地址
func UInt32ToIP(val uint32) string {
	ipData := net.IPv4(byte(val>>24), byte(val>>16&0xFF), byte(val>>8)&0xFF, byte(val&0xFF))
	return ipData.String()
}

// IPToUInt32 ip转数字
func IPToUInt32(ipAddr string) uint32 {
	bits := strings.Split(ipAddr, ".")
	if len(bits) == 4 {
		b0, _ := strconv.Atoi(bits[0])
		b1, _ := strconv.Atoi(bits[1])
		b2, _ := strconv.Atoi(bits[2])
		b3, _ := strconv.Atoi(bits[3])
		var sum uint32
		sum += uint32(b0) << 24
		sum += uint32(b1) << 16
		sum += uint32(b2) << 8
		sum += uint32(b3)
		return sum
	}
	return 0
}
