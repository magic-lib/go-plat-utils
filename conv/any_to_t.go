package conv

import "reflect"

// Convert 转换泛型
func Convert[T any](v any) (T, bool) {
	// 类型断言：尝试将v转换为T
	if result, ok := v.(T); ok {
		return result, true
	}

	var target T
	targetType := reflect.TypeOf(target)
	valueType := reflect.TypeOf(v)

	// 检查类型是否匹配
	if valueType == targetType {
		if targetT, ok := reflect.ValueOf(v).Interface().(T); ok {
			return targetT, true
		}
	}

	elemType := targetType
	// 判断T是否为指针类型
	if targetType.Kind() == reflect.Ptr {
		elemType = targetType.Elem()
	}

	targetPtrValue := reflect.New(elemType).Interface()
	err := Unmarshal(v, targetPtrValue)
	if err != nil {
		err = AssignTo(v, targetPtrValue)
		if err != nil {
			return target, false
		}
		return reflect.ValueOf(targetPtrValue).Elem().Interface().(T), true
	}

	if targetType.Kind() == reflect.Ptr {
		if targetValue, ok := targetPtrValue.(T); ok {
			return targetValue, true
		}
		return target, false
	}

	return reflect.ValueOf(targetPtrValue).Elem().Interface().(T), true
}
