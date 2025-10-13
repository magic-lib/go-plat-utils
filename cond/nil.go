package cond

import (
	"reflect"
	"strings"
)

// IsNil 判断是否为空
func IsNil(i any) bool {
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
	if val == nil {
		return true
	}
	//常用的类型，提高执行效率
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
	case float32, float64, complex64, complex128:
		return val == 0
	case bool:
		return val == false
	case string:
		str, ok := val.(string)
		if ok {
			return strings.TrimSpace(str) == ""
		}
		return val == ""
	default:
	}

	rValue := reflect.ValueOf(val)
	switch rValue.Kind() {
	case reflect.String:
		return rValue.Len() == 0
	case reflect.Bool:
		return !rValue.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rValue.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rValue.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rValue.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return rValue.IsNil()
	default:
	}

	return reflect.ValueOf(val).IsZero()
}

// IsPointer 判断是否是指针类型
func IsPointer(val any) bool {
	if val == nil {
		return false // nil没有类型，不算指针
	}
	// 获取值的类型并判断其Kind是否为指针
	return reflect.TypeOf(val).Kind() == reflect.Ptr
}
