package conv

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/shopspring/decimal"
	"math"
	"reflect"
	"strings"
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

// ConvertForTypeString 按类型名称字符串将 raw 转换为对应类型的值（ConvertForType 的字符串版）。
// targetType 可使用 conv.GoTypeXxx 常量。
// targetType 为空时原样返回 raw；支持的类型名与 ConvertForType 对齐，另加 null：
// string/bool/int/int8/int16/int32/int64/uint/uint8/uint16/uint32/uint64/
// float32/float64/decimal(decimal.Decimal)/time(time.Time)/nil(null)/slice(array)/map。
// 成功返回 (转换后的值, true)；类型未知或转换失败返回 (raw, false)。
func ConvertForTypeString(targetType string, raw any) (any, bool) {
	if targetType == "" {
		return raw, true
	}
	targetType = strings.ToLower(strings.TrimSpace(targetType))
	switch targetType {
	case GoTypeNil:
		return nil, true
	case GoTypeString:
		return String(raw), true
	case GoTypeBool:
		if convRet, err := Convert[bool](raw); err == nil {
			return convRet, true
		}
	case GoTypeInt:
		if convRet, ok := toInt(raw); ok {
			return convRet, true
		}
	case GoTypeInt8:
		if convRet, ok := toInt8(raw); ok {
			return convRet, true
		}
	case GoTypeInt16:
		if convRet, ok := toInt16(raw); ok {
			return convRet, true
		}
	case GoTypeInt32:
		if convRet, ok := toInt32(raw); ok {
			return convRet, true
		}
	case GoTypeInt64:
		if convRet, err := Convert[int64](raw); err == nil {
			return convRet, true
		}
	case GoTypeUint:
		if convRet, ok := toUint(raw); ok {
			return convRet, true
		}
	case GoTypeUint8:
		if convRet, ok := toUint8(raw); ok {
			return convRet, true
		}
	case GoTypeUint16:
		if convRet, ok := toUint16(raw); ok {
			return convRet, true
		}
	case GoTypeUint32:
		if convRet, ok := toUint32(raw); ok {
			return convRet, true
		}
	case GoTypeUint64:
		if convRet, ok := toUint64(raw); ok {
			return convRet, true
		}
	case GoTypeFloat32:
		if convRet, ok := toFloat32(raw); ok {
			return convRet, true
		}
	case GoTypeFloat64:
		if convRet, ok := toFloat64(raw); ok {
			return convRet, true
		}
	case GoTypeDecimal:
		if convRet, ok := toDecimal(raw); ok {
			return convRet, true
		}
	case GoTypeTime:
		if convRet, err := Convert[time.Time](raw); err == nil {
			return convRet, true
		}
	case GoTypeSlice:
		var expAny = make([]any, 0)
		err := Unmarshal(raw, &expAny)
		if err == nil {
			return expAny, true
		}
	case GoTypeMap:
		var expMap = make(map[string]any)
		err := Unmarshal(raw, &expMap)
		if err == nil {
			return expMap, true
		}
	}
	return raw, false
}

// ConvertForTypeJs 按前端 JavaScript 的类型名将 raw 转换为对应类型的值
// （ConvertForTypeString 的 JS 版，类型名忽略大小写与首尾空格）。
// jsType 可使用 conv.JsTypeXxx 常量。
// 支持的 JS 类型名及映射：
//   - string  -> string
//   - boolean -> bool
//   - number  -> int64（整数形态）或 float64（浮点形态），见 toJsNumber
//   - bigint  -> int64
//   - null/undefined -> nil
//   - array   -> []any
//   - object  -> map[string]any
//   - date    -> time.Time
//
// symbol/function 等无法经 JSON 传递的类型不在支持范围内。
// 成功返回 (转换后的值, true)；类型未知或转换失败返回 (raw, false)。
func ConvertForTypeJs(jsType string, raw any) (any, bool) {
	if jsType == "" {
		return raw, true
	}
	jsType = strings.ToLower(strings.TrimSpace(jsType))
	switch jsType {
	case JsTypeString:
		return String(raw), true
	case JsTypeBoolean:
		return ConvertForTypeString(GoTypeBool, raw)
	case JsTypeNumber:
		return toJsNumber(raw)
	case JsTypeBigInt:
		return ConvertForTypeString(GoTypeInt64, raw)
	case JsTypeNull, JsTypeUndefined:
		return nil, true
	case JsTypeArray:
		return ConvertForTypeString(GoTypeSlice, raw)
	case JsTypeObject:
		return ConvertForTypeString(GoTypeMap, raw)
	case JsTypeDate:
		return ConvertForTypeString(GoTypeTime, raw)
	}
	return raw, false
}

// toJsNumber 将 JS 的 number 转换为 Go 值：
//   - 浮点类型：整数值转 int64，否则保留 float64
//   - 字符串：含小数点/指数符号按 float64，否则按 int64
//   - 其他数字类型：先试 int64，失败再试 float64
func toJsNumber(raw any) (any, bool) {
	switch v := raw.(type) {
	case float64:
		if v == math.Trunc(v) {
			return int64(v), true
		}
		return v, true
	case float32:
		f := float64(v)
		if f == math.Trunc(f) {
			return int64(f), true
		}
		return f, true
	}
	if s, ok := raw.(string); ok {
		s = strings.TrimSpace(s)
		if strings.ContainsAny(s, ".eE") {
			return ConvertForTypeString(GoTypeFloat64, s)
		}
		return ConvertForTypeString(GoTypeInt64, s)
	}
	if v, ok := ConvertForTypeString(GoTypeInt64, raw); ok {
		return v, true
	}
	return ConvertForTypeString(GoTypeFloat64, raw)
}
