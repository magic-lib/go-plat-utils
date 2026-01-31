package utils

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"reflect"
	"runtime"
	"strings"
)

// ContextTypedHandler 带有上下文的支持泛型的方法
type ContextTypedHandler[TReq, TResp any] func(ctx context.Context, args TReq) (TResp, error)
type ContextAnyHandler func(ctx context.Context, args any) (any, error)

// ContextTypedToAnyHandler 相互转换
func ContextTypedToAnyHandler[TReq, TResp any](method any) (ContextAnyHandler, error) {
	if method == nil {
		panic("method is nil")
	}
	methodFun, ok := method.(ContextTypedHandler[TReq, TResp])
	if !ok {
		methodName, isMethod := GetFuncName(method)
		if !isMethod {
			return nil, fmt.Errorf("%s error: method is not function", methodName)
		}
		return nil, fmt.Errorf("method is not func(ctx context.Context, param P) (V, error): %s", methodName)
	}
	return func(ctx context.Context, param any) (any, error) {
		//断言
		paramPtr, ok := param.(TReq)
		if !ok {
			return nil, fmt.Errorf("param is not %T", paramPtr)
		}
		//调用方法
		return methodFun(ctx, paramPtr)
	}, nil
}

// GetFuncName 获取方法名
func GetFuncName(i any) (name string, isMethod bool) {
	if fullName, ok := i.(string); ok {
		return fullName, false
	}
	fullName := runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
	isMethod = strings.ContainsAny(fullName, "*")
	elements := strings.Split(fullName, ".")
	shortName := elements[len(elements)-1]
	return strings.TrimSuffix(shortName, "-fm"), isMethod
}

// FuncExecute 根据参数，可以调用任何方法
func FuncExecute(function any, args ...any) (result []any, err error) {
	// 将空接口转换为reflect.Value
	fnType := reflect.TypeOf(function)
	fnValue := reflect.ValueOf(function)

	// 检查是否为可调用的函数
	if fnValue.Kind() != reflect.Func {
		return nil, fmt.Errorf("[FuncExecute] not function")
	}

	trueArgList, err := FuncArgList(function, args...)
	if err != nil {
		return nil, err
	}

	// 调用函数，捕获panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("[FuncExecute] Recovered from panic:", r)
		}
	}()
	argSlice := make([]reflect.Value, len(trueArgList))
	for i, v := range trueArgList {
		argSlice[i] = reflect.ValueOf(v)
	}
	// 调用函数并返回结果
	retValues := fnValue.Call(argSlice)

	numOut := fnType.NumOut()

	if numOut == 0 && len(retValues) == 0 {
		return []any{}, nil
	}

	if len(retValues) != numOut {
		return []any{}, fmt.Errorf("[FuncExecute] ret error, "+
			"retValue number is %d, but Type number is %d", len(retValues), numOut)
	}

	//retTypes := make([]reflect.Type, numOut)
	//for i := 0; i < numOut; i++ {
	//	retTypes[i] = fnType.Out(i)
	//}

	result = make([]any, numOut)
	for i := 0; i < numOut; i++ {
		result[i] = retValues[i].Interface()
	}

	return result, nil
}

// FuncInTypeList 获取函数的参数类型列表和是否有可变参数
func FuncInTypeList(function any) ([]reflect.Type, bool, error) {
	value := reflect.ValueOf(function)
	if value.Kind() != reflect.Func {
		return nil, false, fmt.Errorf("[FuncArgTypeList] not function")
	}
	fnType := reflect.TypeOf(function)
	numIn := fnType.NumIn()
	paramTypes := make([]reflect.Type, numIn)
	for i := 0; i < numIn; i++ {
		paramTypes[i] = fnType.In(i)
	}
	hasVariadicParam := false
	if numIn > 0 && fnType.IsVariadic() {
		hasVariadicParam = true
	}
	return paramTypes, hasVariadicParam, nil
}

func FuncArgList(function any, args ...any) ([]any, error) {
	argTypeList, hasVariadicParam, err := FuncInTypeList(function)
	if err != nil {
		return nil, err
	}
	if len(argTypeList) == 0 {
		return []any{}, nil
	}
	var trueArgList = args
	if !hasVariadicParam {
		if len(args) > len(argTypeList) {
			trueArgList = args[0:len(argTypeList)]
		}
	}

	getZeroConvertValue := func(fnArgType reflect.Type) reflect.Value {
		if fnArgType.Kind() == reflect.Interface {
			return reflect.Zero(reflect.TypeOf((*any)(nil)).Elem())
		}
		if fnArgType.Kind() == reflect.Ptr {
			return reflect.Zero(fnArgType)
		}
		return reflect.Zero(fnArgType)
	}

	// 补充参数
	argTypeLen := len(argTypeList)
	for len(trueArgList) < argTypeLen {
		if hasVariadicParam {
			if len(trueArgList) >= argTypeLen-1 {
				break
			}
		}
		zeroValue := getZeroConvertValue(argTypeList[len(trueArgList)]).Interface()
		trueArgList = append(trueArgList, zeroValue)
	}
	var variadicType reflect.Type
	if hasVariadicParam {
		variadicType = argTypeList[len(argTypeList)-1].Elem()
	}

	getConvertValue := func(arg any, thisArgType reflect.Type) reflect.Value {
		if arg == nil {
			return getZeroConvertValue(thisArgType)
		}
		argValue := reflect.ValueOf(arg)
		if argValue.Type().ConvertibleTo(thisArgType) {
			return argValue.Convert(thisArgType)
		} else {
			newArg, err := conv.ConvertForType(thisArgType, arg)
			if err == nil {
				return reflect.ValueOf(newArg)
			}
		}
		return getZeroConvertValue(thisArgType)
	}

	argSlice := make([]reflect.Value, 0, len(trueArgList))
	for i, arg := range trueArgList {
		if hasVariadicParam && i >= len(argTypeList)-1 {
			thisValue := getConvertValue(arg, variadicType)
			argSlice = append(argSlice, thisValue)
		} else {
			thisArgType := argTypeList[i]
			thisValue := getConvertValue(arg, thisArgType)
			argSlice = append(argSlice, thisValue)
		}
	}

	result := make([]any, len(argSlice))
	for i, v := range argSlice {
		result[i] = v.Interface()
	}
	return result, nil
}
