package structx

import (
	"reflect"
)

// MergeStructs 将多个源结构体的非默认值属性合并到目标结构体中
// target: 目标结构体指针，结果将写入此结构体
// sources: 一个或多个源结构体（可以是结构体或结构体指针）
// 返回值：如果发生错误则返回错误信息
func MergeStructs(target interface{}, sources ...interface{}) error {
	if target == nil {
		return nil
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return nil
	}

	targetElem := targetValue.Elem()
	if targetElem.Kind() != reflect.Struct {
		return nil
	}

	// 遍历所有源结构体
	for _, source := range sources {
		if source == nil {
			continue
		}

		sourceValue := reflect.ValueOf(source)

		// 如果是指针，获取其指向的值
		if sourceValue.Kind() == reflect.Ptr {
			if sourceValue.IsNil() {
				continue
			}
			sourceValue = sourceValue.Elem()
		}

		if sourceValue.Kind() != reflect.Struct {
			continue
		}

		// 合并当前源结构体的属性
		mergeStructFields(targetElem, sourceValue)
	}

	return nil
}

// mergeStructFields 将源结构体的非默认值字段合并到目标结构体
func mergeStructFields(target, source reflect.Value) {
	sourceType := source.Type()

	for i := 0; i < source.NumField(); i++ {
		sourceField := source.Field(i)
		targetField := target.FieldByName(sourceType.Field(i).Name)

		// 检查目标字段是否存在且可设置
		if !targetField.IsValid() || !targetField.CanSet() {
			continue
		}

		// 只复制非默认值的字段
		if !isDefaultValue(sourceField) {
			targetField.Set(sourceField)
		}
	}
}

// isDefaultValue 检查字段是否为默认值
func isDefaultValue(field reflect.Value) bool {
	switch field.Kind() {
	case reflect.Bool:
		return !field.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return field.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return field.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return field.Complex() == 0
	case reflect.String:
		return field.String() == ""
	case reflect.Ptr, reflect.Interface, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func:
		return field.IsNil()
	case reflect.Array:
		// 对于数组，检查是否所有元素都是默认值
		for i := 0; i < field.Len(); i++ {
			if !isDefaultValue(field.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		// 对于结构体，递归检查所有字段
		for i := 0; i < field.NumField(); i++ {
			if !isDefaultValue(field.Field(i)) {
				return false
			}
		}
		return true
	default:
		return false
	}
}
