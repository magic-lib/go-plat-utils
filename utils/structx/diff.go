package structx

import "reflect"

// DiffFields 比较两个结构体中指定字段的属性值是否发生改变，并返回改变的字段名
func DiffFields[T any](t1 T, t2 T, fieldsToCompare ...string) []string {
	var changedFields []string

	// T 为指针时先解引用；非结构体没有字段可比较
	oldValue := reflect.Indirect(reflect.ValueOf(t1))
	newValue := reflect.Indirect(reflect.ValueOf(t2))
	if oldValue.Kind() != reflect.Struct || newValue.Kind() != reflect.Struct {
		return changedFields
	}

	for _, fieldName := range fieldsToCompare {
		oldFieldValue := oldValue.FieldByName(fieldName)
		newFieldValue := newValue.FieldByName(fieldName)
		// 字段不存在或未导出时跳过，否则 Interface() 会 panic
		if !oldFieldValue.IsValid() || !newFieldValue.IsValid() ||
			!oldFieldValue.CanInterface() || !newFieldValue.CanInterface() {
			continue
		}

		if !reflect.DeepEqual(oldFieldValue.Interface(), newFieldValue.Interface()) {
			changedFields = append(changedFields, fieldName)
		}
	}

	return changedFields
}
