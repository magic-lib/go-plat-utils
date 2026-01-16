package conv

import "reflect"

func anyToSlice(in any) ([]any, bool) {
	value := reflect.ValueOf(in)
	if value.Kind() != reflect.Slice {
		return nil, false
	}
	result := make([]any, value.Len())
	for i := 0; i < value.Len(); i++ {
		result[i] = value.Index(i).Interface()
	}
	return result, true
}
