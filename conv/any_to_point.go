package conv

import (
	"reflect"
)

func Pointer[E any](v any) *E {
	if v == nil {
		return nil
	}
	if p, ok := v.(*E); ok {
		return p
	}

	vType := reflect.TypeOf(v)
	if vType.Kind() == reflect.Ptr {
		return v.(*E) // 已经是指针，直接转型
	}

	ptr := new(E)
	switch val := any(ptr).(type) {
	case *int:
		if intV, ok := toInt(v); ok {
			*val = intV
		}
	case *int32:
		if int32V, ok := toInt32(v); ok {
			*val = int32V
		}
	case *int64:
		if int64V, ok := Int64(v); ok {
			*val = int64V
		}
	case *string:
		*val = String(v)
	case *bool:
		if boolV, ok := Bool(v); ok {
			*val = boolV
		}
	case *float64:
		if floatV, ok := toFloat64(v); ok {
			*val = floatV
		}
	default:
		rv := reflect.ValueOf(v)
		target := reflect.ValueOf(ptr).Elem()
		if rv.Type() != target.Type() {
			toElem, err := Convert[E](v)
			if err == nil {
				rv = reflect.ValueOf(toElem)
			} else if rv.CanConvert(target.Type()) {
				rv = rv.Convert(target.Type())
			}
		}
		if rv.Type() == target.Type() {
			target.Set(rv)
		}
	}
	return ptr
}
