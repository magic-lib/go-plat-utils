package conv

import (
	"database/sql"
	"reflect"
	"strconv"
	"strings"
)

// Bool 将给定的值转换为bool
// Deprecated: 该方法已废弃，请使用 conv.Convert[bool](v)
func Bool(i any) (bool, bool) {
	if i == nil {
		return false, true
	}
	if b, ok := i.(bool); ok {
		return b, true
	}
	if b, ok := toBool(i); ok {
		return b, true
	}
	if b, err := toConvert[int64](i); err == nil {
		if b == 0 {
			return false, true
		}
		return true, true
	}
	vBool, ok := getBySqlNullBool(i)
	if ok {
		return vBool, true
	}

	v := reflect.ValueOf(i)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return false, true
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0, true
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0, true
	case reflect.Complex64, reflect.Complex128:
		return v.Complex() != 0, true
	case reflect.String:
		return toBool(v.String())
	case reflect.Slice, reflect.Array: //数组只要里面含有元素就表示为true
		length := v.Len()
		if length == 0 {
			return false, true
		}
		return true, true
	default:
		return Bool(String(i))
	}
}

func getBySqlNullBool(src any) (bool, bool) {
	if strNull, ok := src.(sql.NullBool); ok {
		if strNull.Valid {
			return strNull.Bool, true
		}
		return false, true
	}
	return false, false
}

func toBool(val any) (bool, bool) {
	switch v := val.(type) {
	case bool:
		return v, true
	case string:
		v = strings.TrimSpace(v)
		if v == "" {
			return false, true
		}
		valLower := strings.ToLower(v)
		if valLower == "true" || valLower == "yes" {
			return true, true
		} else if valLower == "false" || valLower == "no" {
			return false, true
		}
		boolValue, err := strconv.ParseBool(valLower)
		if err != nil {
			return false, false
		}
		return boolValue, true
	case int, int8, int16, int32, int64:
		return v != 0, true
	case float32, float64:
		return v != 0, true
	default:
		return false, false
	}
}
