package cond

import (
	"reflect"
)

// IsNil 判断是否为空
func IsNil(i interface{}) bool {
	if i == nil {
		return true
	}
	vi := reflect.ValueOf(i)
	kind := vi.Kind()
	if kind == reflect.Ptr ||
		kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.UnsafePointer ||
		kind == reflect.Map ||
		kind == reflect.Interface ||
		kind == reflect.Slice {
		return vi.IsNil()
	}
	return false
}

// IsZero 判断变量是否为零值
func IsZero(val any) bool {
	//常用的类型，提高执行效率
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
	case float32, float64, complex64, complex128:
		return val == 0
	case bool:
		return val == false
	case string:
		return val == ""
	}

	return reflect.ValueOf(val).IsZero()
}
