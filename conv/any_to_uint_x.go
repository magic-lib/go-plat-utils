package conv

import (
	"math"
	"reflect"
	"strconv"
)

// Uint 将给定的值转换为 uint
func Uint(i any) (uint, bool) {
	v, ok := Uint64(i)
	if !ok {
		return 0, false
	}
	return uint(v), true
}

// Uint8 将给定的值转换为 uint8
func Uint8(i any) (uint8, bool) {
	value, ok := Uint64(i)
	if !ok {
		return 0, false
	}
	if value > math.MaxUint8 {
		return 0, false
	}
	return uint8(value), true
}

// Uint16 将给定的值转换为 uint16
func Uint16(i any) (uint16, bool) {
	value, ok := Uint64(i)
	if !ok {
		return 0, false
	}
	if value > math.MaxUint16 {
		return 0, false
	}
	return uint16(value), true
}

// Uint32 将给定的值转换为 uint32
func Uint32(i any) (uint32, bool) {
	value, ok := Uint64(i)
	if !ok {
		return 0, false
	}
	if value > math.MaxUint32 {
		return 0, false
	}
	return uint32(value), true
}

// Uint64 将给定的值转换为 uint64
func Uint64(i any) (uint64, bool) {
	if i == nil {
		return 0, false
	}

	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0, false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intValue := v.Int()
		if intValue < 0 {
			return 0, false
		}
		return uint64(intValue), true
	case reflect.Float32, reflect.Float64:
		floatValue := v.Float()
		if floatValue < 0 {
			return 0, false
		}
		return uint64(floatValue), true
	case reflect.Complex64, reflect.Complex128:
		realValue := real(v.Complex())
		if realValue < 0 {
			return 0, false
		}
		return uint64(realValue), true
	case reflect.String:
		strValue := v.String()
		uintValue, err := strconv.ParseUint(strValue, 10, 64)
		if err != nil {
			return 0, false
		}
		return uintValue, true
	case reflect.Bool:
		if v.Bool() {
			return 1, true
		} else {
			return 0, true
		}
	default:
		return 0, false
	}
}
