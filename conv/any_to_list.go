package conv

import (
	"reflect"
)

func anyToSlice(in any) ([]any, bool) {
	if s, ok := in.([]any); ok {
		return s, true
	}
	s := reflect.ValueOf(in)
	kind := s.Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		l := s.Len()
		ret := make([]any, l)
		for i := range l {
			ret[i] = s.Index(i).Interface()
		}
		return ret, true
	default:
		return nil, false
	}
}
