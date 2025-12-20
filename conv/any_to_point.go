package conv

import (
	"reflect"
)

func Pointer(v any) any {
	if v == nil {
		return nil
	}
	vType := reflect.TypeOf(v)
	if vType.Kind() == reflect.Ptr {
		return v
	}

	switch val := v.(type) {
	case int:
		return &val
	case string:
		return &val
	case bool:
		return &val
	case float64:
		return &val
	case struct{}:
		return &val
	default:
		rv := reflect.ValueOf(v)
		ptr := reflect.New(rv.Type())
		ptr.Elem().Set(rv)
		return ptr.Interface()
	}
}
