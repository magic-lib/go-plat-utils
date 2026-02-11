package conv

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/shopspring/decimal"
	"reflect"
	"time"
)

// Convert 转换泛型
func Convert[T any](v any) (t T, err error) {
	if cond.IsArray(t) && cond.IsArray(v) {
		t, err = toConvertList[T](v)
		if err == nil {
			return t, nil
		}
	}
	return toConvert[T](v)
}

func toConvertList[T any](v any) (T, error) {
	var t T
	list, ok := anyToSlice(v)
	if !ok {
		return t, fmt.Errorf("convert not match []any")
	}
	targetType := reflect.TypeOf(t)
	elementType := targetType.Elem()
	sliceValue := reflect.MakeSlice(targetType, 0, len(list))
	for _, item := range list {
		itemT, err := ConvertForType(elementType, item)
		if err != nil {
			var zero T
			return zero, err
		}
		sliceValue = reflect.Append(sliceValue, reflect.ValueOf(itemT))
	}
	return sliceValue.Interface().(T), nil
}

func toConvert[T any](v any) (T, error) {
	// 类型断言：尝试将v转换为T
	if result, ok := v.(T); ok {
		return result, nil
	}
	var target T
	targetType := reflect.TypeOf(target)

	targetValue, err := ConvertForType(targetType, v)
	if err != nil {
		return target, err
	}
	if targetT, ok := targetValue.(T); ok {
		return targetT, nil
	}
	return target, fmt.Errorf("convert not match T")
}

// ConvertForType 泛型转换
func ConvertForType(targetType reflect.Type, v any) (any, error) {
	valueType := reflect.TypeOf(v)
	// 检查类型是否匹配
	if valueType == targetType {
		return v, nil
	}
	// 如果目标类型是any的话，则直接返回即可
	if targetType.Kind() == reflect.Interface {
		return v, nil
	}

	convErr := fmt.Errorf("unsupported targetType: %s, value: %s", targetType.String(), String(v))
	contConv := false

	switch targetType {
	case reflect.TypeOf(true):
		{
			if convRet, ok := Bool(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(0):
		{
			if convRet, ok := toInt(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(int8(0)):
		{
			if convRet, ok := toInt8(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(int16(0)):
		{
			if convRet, ok := toInt16(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(int32(0)):
		{
			if convRet, ok := toInt32(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(int64(0)):
		{
			if convRet, ok := Int64(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(uint(0)):
		{
			if convRet, ok := toUint(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(uint8(0)):
		{
			if convRet, ok := toUint8(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(uint16(0)):
		{
			if convRet, ok := toUint16(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(uint32(0)):
		{
			if convRet, ok := toUint32(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(uint64(0)):
		{
			if convRet, ok := toUint64(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(float32(0)):
		{
			if convRet, ok := toFloat32(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(float64(0)):
		{
			if convRet, ok := toFloat64(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(""):
		{
			return String(v), nil
		}
	case reflect.TypeOf(time.Time{}):
		{
			if convRet, ok := Time(v); ok {
				return convRet, nil
			}
		}
	case reflect.TypeOf(decimal.Decimal{}):
		{
			convRet, ok := toDecimal(v)
			if ok {
				return convRet, nil
			}
			return convRet, convErr
		}
	default:
		//log.Println("ConvertForType: ", convErr.Error())
		contConv = true
	}

	if !contConv {
		return v, convErr
	}

	target := reflect.Zero(targetType)

	elemType := targetType
	// 判断T是否为指针类型
	if targetType.Kind() == reflect.Ptr {
		elemType = targetType.Elem()
	}

	targetPtrValue := reflect.New(elemType).Interface()
	err := Unmarshal(v, targetPtrValue)
	if err != nil {
		err = toAssignTo(v, targetPtrValue)
		if err != nil {
			return target, convErr
		}
		return reflect.ValueOf(targetPtrValue).Elem().Interface(), nil
	}

	if targetType.Kind() == reflect.Ptr {
		if reflect.TypeOf(targetPtrValue) == targetType {
			return targetPtrValue, nil
		}
		return target, convErr
	}

	return reflect.ValueOf(targetPtrValue).Elem().Interface(), nil
}
