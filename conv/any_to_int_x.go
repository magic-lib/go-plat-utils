package conv

import (
	"encoding/binary"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// Int 将给定的值转换为 int
func Int(i any) (int, bool) {
	v, ok := Int64(i)
	if !ok {
		return 0, false
	}
	return int(v), true
}

// Int8 将给定的值转换为 int8
func Int8(i any) (int8, bool) {
	value, ok := Int64(i)
	if !ok {
		return 0, false
	}
	if value < math.MinInt8 || value > math.MaxInt8 {
		return 0, false
	}
	return int8(value), true
}

// Int16 将给定的值转换为 int16
func Int16(i any) (int16, bool) {
	value, ok := Int64(i)
	if !ok {
		return 0, false
	}
	if value < math.MinInt16 || value > math.MaxInt16 {
		return 0, false
	}
	return int16(value), true
}

// Int32 将给定的值转换为 int32
func Int32(i any) (int32, bool) {
	value, ok := Int64(i)
	if !ok {
		return 0, false
	}
	if value < math.MinInt32 || value > math.MaxInt32 {
		return 0, false
	}
	return int32(value), true
}

// Int64 将给定的值转换为 int64
func Int64(i any) (int64, bool) {
	if i == nil {
		return 0, false
	}

	intTemp, ok := forceToInt64(i)
	if ok {
		return intTemp, true
	}

	switch i.(type) {
	case string:
		intStr := strings.TrimSpace(i.(string))
		num, err := strconv.Atoi(intStr)
		if err != nil {
			//转数字错误，判断是否小数点后面全部是0，则可以去掉 3.0000
			strList := strings.Split(intStr, ".")
			if len(strList) == 2 {
				num1, err := strconv.Atoi(strList[1])
				if err == nil && num1 == 0 {
					num2, err := strconv.Atoi(strList[0])
					if err == nil {
						return int64(num2), true
					}
				}
			}

			return 0, false
		}
		return int64(num), true
	case []byte:
		bits := i.([]byte)
		if len(bits) == 8 {
			return int64(binary.LittleEndian.Uint64(bits)), true
		} else if len(bits) <= 4 {
			num, err := strconv.Atoi(string(bits))
			if err != nil {
				return 0, false
			}
			return int64(num), true
		}
	}

	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return 0, false
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return int64(v.Float()), true
	case reflect.String:
		intValue, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return 0, false
		}
		return intValue, true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(v.Uint()), true
	case reflect.Complex64:
		return int64(real(v.Complex())), true
	case reflect.Complex128:
		return int64(real(v.Complex())), true
	case reflect.Bool:
		if v.Bool() {
			return 1, true
		} else {
			return 0, true
		}
	default:
		//转成字符串来处理
		return Int64(String(i))
	}
}

func forceToInt64(i interface{}) (int64, bool) {
	switch i.(type) {
	case uint:
		return int64(i.(uint)), true
	case uint8:
		return int64(i.(uint8)), true
	case uint16:
		return int64(i.(uint16)), true
	case uint32:
		return int64(i.(uint32)), true
	case uint64:
		return int64(i.(uint64)), true
	case int:
		return int64(i.(int)), true
	case int8:
		return int64(i.(int8)), true
	case int16:
		return int64(i.(int16)), true
	case int32:
		return int64(i.(int32)), true
	case int64:
		return i.(int64), true
	case float32:
		return int64(i.(float32)), true
	case float64:
		return int64(i.(float64)), true
	}
	return 0, false
}
