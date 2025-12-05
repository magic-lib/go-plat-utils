package conv

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// MapToStruct 将 map[string]interface{} 转换为 struct 指针
func mapToStruct(m map[string]any, dstPoint any) error {
	// 检查目标是否为 struct 指针
	dstValue := reflect.ValueOf(dstPoint)
	if dstValue.Kind() != reflect.Ptr || dstValue.Elem().Kind() != reflect.Struct {
		return errors.New("目标必须是 struct 指针")
	}

	// 获取 struct 的值（解引用指针）
	dstStructVal := dstValue.Elem()
	for key, value := range m {
		field, found := findField(dstStructVal.Type(), key)
		if !found {
			continue // 忽略未匹配的字段
		}

		// 检查字段是否可设置
		fieldVal := dstStructVal.FieldByName(field.Name)
		if !fieldVal.CanSet() {
			continue // 跳过不可设置的字段（如私有字段）
		}

		// 递归转换值并赋值给字段
		if err := setFieldValue(fieldVal, value); err != nil {
			return fmt.Errorf("字段 %s 赋值失败: %v", field.Name, err)
		}
	}
	return nil
}

// findField 查找 struct 中与 map key 匹配的字段（优先 json 标签，再字段名忽略大小写）
func findField(structType reflect.Type, key string) (reflect.StructField, bool) {
	//c := new(getNewService)
	//valueTemp := c.GetSrcFromStructField(srcStruct, dstColumnField)

	t := new(toolsService)
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		dstColumnJsonName := t.getTagJsonName(field)
		if dstColumnJsonName != "" {
			if dstColumnJsonName == key {
				return field, true
			}
		}
		if strings.EqualFold(field.Name, key) {
			return field, true
		}
	}
	return reflect.StructField{}, false
}

// setFieldValue 递归设置字段值（处理基本类型、嵌套 struct、切片等）
func setFieldValue(fieldVal reflect.Value, value interface{}) error {
	if value == nil {
		return nil // 忽略 nil 值
	}

	// 获取字段类型和值类型
	fieldType := fieldVal.Type()
	valueVal := reflect.ValueOf(value)

	// 处理指针类型（如字段是 *int，值是 int）
	if fieldVal.Kind() == reflect.Ptr {
		// 若字段是 nil 指针，先初始化
		if fieldVal.IsNil() {
			fieldVal.Set(reflect.New(fieldType.Elem()))
		}
		// 递归处理指针指向的元素
		return setFieldValue(fieldVal.Elem(), value)
	}

	// 处理切片类型（如 []int、[]Struct）
	if fieldVal.Kind() == reflect.Slice {
		return setSliceValue(fieldVal, value)
	}

	// 处理嵌套 struct（值是 map[string]interface{}）
	if fieldVal.Kind() == reflect.Struct && valueVal.Kind() == reflect.Map {
		// 将 map 转换为嵌套 struct
		nestedMap, ok := value.(map[string]any)
		if !ok {
			return errors.New("嵌套值不是 map[string]interface{}")
		}
		return mapToStruct(nestedMap, fieldVal.Addr().Interface())
	}

	// 处理基本类型（如 int、string、bool 等）
	return setBasicValue(fieldVal, value)
}

// setSliceValue 处理切片类型的赋值（支持 []interface{} 转换为 []T）
func setSliceValue(sliceVal reflect.Value, value interface{}) error {
	// 检查值是否为切片
	valueSlice, ok := value.([]interface{})
	if !ok {
		return errors.New("值不是 []interface{} 类型")
	}

	// 获取切片元素类型
	elemType := sliceVal.Type().Elem()
	// 创建新切片
	newSlice := reflect.MakeSlice(sliceVal.Type(), len(valueSlice), len(valueSlice))

	// 遍历切片元素，递归转换
	for i, elem := range valueSlice {
		elemVal := newSlice.Index(i)
		// 若元素是 struct 或指针，初始化并递归赋值
		if elemVal.Kind() == reflect.Struct || (elemVal.Kind() == reflect.Ptr && elemVal.Type().Elem().Kind() == reflect.Struct) {
			// 初始化元素（如指针）
			if elemVal.Kind() == reflect.Ptr && elemVal.IsNil() {
				elemVal.Set(reflect.New(elemType.Elem()))
			}
			// 递归转换 map 为 struct
			if err := setFieldValue(elemVal, elem); err != nil {
				return fmt.Errorf("切片元素 %d 转换失败: %v", i, err)
			}
		} else {
			// 基本类型切片（如 []int）
			if err := setBasicValue(elemVal, elem); err != nil {
				return fmt.Errorf("切片元素 %d 转换失败: %v", i, err)
			}
		}
	}

	// 赋值给目标切片
	sliceVal.Set(newSlice)
	return nil
}

// setBasicValue 处理基本类型的赋值（如 int、string、bool 等）
func setBasicValue(fieldVal reflect.Value, value interface{}) error {
	valueVal := reflect.ValueOf(value)

	// 直接赋值（类型完全匹配）
	if valueVal.Type().AssignableTo(fieldVal.Type()) {
		fieldVal.Set(valueVal)
		return nil
	}

	// 类型转换（如 float64 转 int，string 转 bool 等）
	switch fieldVal.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return setIntValue(fieldVal, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return setUintValue(fieldVal, value)
	case reflect.Float32, reflect.Float64:
		return setFloatValue(fieldVal, value)
	case reflect.Bool:
		return setBoolValue(fieldVal, value)
	case reflect.String:
		fieldVal.SetString(String(value))
		return nil
	default:
		return fmt.Errorf("不支持的类型转换: 目标类型 %s, 源值 %v", fieldVal.Kind(), value)
	}
}

func setIntValue(fieldVal reflect.Value, value interface{}) error {
	intVal, err := Convert[int64](value)
	if err == nil {
		fieldVal.SetInt(intVal)
		return nil
	}
	return fmt.Errorf("无法将 %T 转换为 int", value)
}

func setUintValue(fieldVal reflect.Value, value interface{}) error {
	intVal, err := Convert[uint64](value)
	if err == nil {
		fieldVal.SetUint(intVal)
		return nil
	}
	return fmt.Errorf("无法将 %T 转换为 uint", value)
}

func setFloatValue(fieldVal reflect.Value, value interface{}) error {
	floatVal, err := Convert[float64](value)
	if err == nil {
		fieldVal.SetFloat(floatVal)
		return nil
	}
	return fmt.Errorf("无法将 %T 转换为 float", value)
}

func setBoolValue(fieldVal reflect.Value, value interface{}) error {
	parsed, err := Convert[bool](value)
	if err == nil {
		fieldVal.SetBool(parsed)
		return nil
	}
	return fmt.Errorf("无法将 %T 转换为 bool", value)
}
