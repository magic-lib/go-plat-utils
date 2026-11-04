package conv

import (
	"math"
	"reflect"
)

// toUint 将给定的值转换为 uint
func toUint(i any) (uint, bool) {
	v, ok := toUint64(i)
	if !ok {
		return 0, false
	}
	if v > math.MaxUint {
		return math.MaxUint, true
	}
	return uint(v), true
}

// toUint8 将给定的值转换为 uint8
func toUint8(i any) (uint8, bool) {
	value, ok := toUint64(i)
	if !ok {
		return 0, false
	}
	if value > math.MaxUint8 {
		return math.MaxUint8, true
	}
	return uint8(value), true
}

// toUint16 将给定的值转换为 uint16
func toUint16(i any) (uint16, bool) {
	value, ok := toUint64(i)
	if !ok {
		return 0, false
	}
	if value > math.MaxUint16 {
		return math.MaxUint16, true
	}
	return uint16(value), true
}

// toUint32 将给定的值转换为 uint32
func toUint32(i any) (uint32, bool) {
	value, ok := toUint64(i)
	if !ok {
		return 0, false
	}
	if value > math.MaxUint32 {
		return math.MaxUint32, true
	}
	return uint32(value), true
}

// toUint64 将给定的值转换为 uint64
func toUint64(i any) (uint64, bool) {
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
		if intValue < 0 { // 低于最小值取 0
			return 0, true
		}
		return uint64(intValue), true
	case reflect.Float32, reflect.Float64:
		return floatToUint64(v.Float())
	case reflect.Complex64, reflect.Complex128:
		return floatToUint64(real(v.Complex()))
	case reflect.String:
		strValue := v.String()
		uintValue, err := parseUint64(strValue)
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
