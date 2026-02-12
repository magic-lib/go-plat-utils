package cond

import (
	"reflect"
	"strings"
	"time"
)

// IsNil 判断是否为空
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	value := reflect.ValueOf(v)
	kind := value.Kind()
	if kind == reflect.Ptr ||
		kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.UnsafePointer ||
		kind == reflect.Map ||
		kind == reflect.Interface ||
		kind == reflect.Slice {
		if value.IsNil() {
			return true
		}
	} else if kind == reflect.Struct {
		for i := 0; i < value.NumField(); i++ {
			if !value.Field(i).IsZero() {
				return false
			}
		}
		return true
	} else if kind == reflect.Invalid {
		return true
	} else if kind == reflect.String ||
		kind == reflect.Bool {
		return false
	} else if kind == reflect.Int ||
		kind == reflect.Int8 ||
		kind == reflect.Int16 ||
		kind == reflect.Int32 ||
		kind == reflect.Int64 {
		return false
	} else if kind == reflect.Float32 ||
		kind == reflect.Float64 {
		return false
	} else if kind == reflect.Uint ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64 {
		return false
	} else if kind == reflect.Complex64 ||
		kind == reflect.Complex128 {
		return false
	}

	if !value.IsValid() {
		return true
	}

	if value == reflect.ValueOf(nil) {
		return true
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
	case time.Time:
		if valTime, ok := val.(time.Time); ok {
			return IsTimeEmpty(valTime)
		}
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
