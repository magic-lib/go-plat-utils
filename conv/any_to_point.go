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

	// 其他类型的指针：v.(*E) 已经断言失败，不能再强转（会 panic），解引用后按值继续转换
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil
		}
		v = rv.Elem().Interface()
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
		if !rv.IsValid() { // 解引用后为 nil 时无值可转换，rv.Type() 会 panic
			return ptr
		}
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
