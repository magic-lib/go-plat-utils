package conv

import (
	"encoding/json"
	"fmt"
	"math"
	"reflect"
	"strconv"
)

// Float32 将给定的值转换为float32
func Float32(i any) (float32, bool) {
	f64, ok := Float64(i)
	if !ok {
		return 0, false
	}
	if f64 < -math.MaxFloat32 || f64 > math.MaxFloat32 {
		return 0, false
	}
	return float32(f64), true
}

// Float64 将给定的值转换为float64
func Float64(i any) (float64, bool) {
	if i == nil {
		return 0, false
	}
	if f, ok := i.(float32); ok {
		str := fmt.Sprintf("%f", f)
		v, _ := strconv.ParseFloat(str, 64)
		return v, true
	}

	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0, false
		}
		v = v.Elem()
		if !v.IsValid() {
			return 0, false
		}
		i = v.Interface()
		if i == nil {
			return 0, false
		}
	}

	switch v1 := i.(type) {
	case []byte:
		t, err := strconv.ParseFloat(string(v1), 64)
		if err == nil {
			return t, true
		}
		return 0, false
	case json.Number:
		i, err := v1.Float64()
		if err != nil {
			return 0, false
		}
		return i, true
	case string:
		if len(v1) > 15 {
			return 0, false
		}
		t, err := strconv.ParseFloat(v1, 64)
		if err != nil {
			return 0, false
		}
		return t, true
	}

	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.String:
		floatValue, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return 0, false
		}
		return floatValue, true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), true
	case reflect.Complex64, reflect.Complex128:
		return real(v.Complex()), true
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
