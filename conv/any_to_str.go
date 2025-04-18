package conv

import (
	"fmt"
	"strconv"
	"unsafe"
)

// BytesToString 字节数组转字符串
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// Float32ToString float32转字符串
func Float32ToString(f float32) string {
	return fmt.Sprintf("%v", f)
}

// Float64ToString float64转字符串
func Float64ToString(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// StrToBytes 字符串转字节数组
func StrToBytes(v string) []byte {
	return *(*[]byte)(unsafe.Pointer(&v))
}
