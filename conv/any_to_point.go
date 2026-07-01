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
		*val = v.(int)
	case *string:
		*val = v.(string)
	case *bool:
		*val = v.(bool)
	case *float64:
		*val = v.(float64)
	default:
		rv := reflect.ValueOf(v)
		reflect.ValueOf(ptr).Elem().Set(rv)
	}
	return ptr
}
