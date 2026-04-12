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

// ContextMethodToAnyHandler 相互转换
func ContextMethodToAnyHandler[TReq, TResp any](method any) (ContextAnyHandler, error) {
	if method == nil {
		panic("method is nil")
	}

	// 这两种类型底层完全一样，但仍然是两种不同的类型
	methodFun, ok := method.(func(ctx context.Context, args TReq) (TResp, error))
	if !ok {
		methodFun, ok = method.(ContextTypedHandler[TReq, TResp])
		if !ok {
			// 最后进行转换
			var err error
			methodFun, err = ContextMethodToTypeHandler[TReq, TResp](method)
			if err != nil {
				methodName, isMethod := GetFuncName(method)
				if !isMethod {
					return nil, fmt.Errorf("%s error: method is not function", methodName)
				}
				return nil, fmt.Errorf("method is not func(ctx context.Context, param P) (V, error): %s", methodName)
			}
		}
	}
	return func(ctx context.Context, param any) (any, error) {
		paramPtr, ok := param.(TReq)
		if !ok {
			var zero TReq
			actionParam, err := conv.ConvertForType(reflect.TypeOf(zero), param)
			if err != nil {
				funcName, _ := GetFuncName(method)
				paramList, _, _ := FuncInTypeList(method)
				return nil, fmt.Errorf("methodName: %s, %s, param is %s, not %T, err: %v", funcName, conv.String(paramList), conv.String(param), reflect.TypeOf(zero).String(), err)
			}
			paramPtr, ok = actionParam.(TReq)
			if !ok {
				return nil, fmt.Errorf("param is not %T", paramPtr)
			}
		}
		retData, err := methodFun(ctx, paramPtr)
		retDataPtr, ok := any(retData).(TResp)
		if !ok {
			var zero TResp
			return nil, fmt.Errorf("retData is %T, not %T", retDataPtr, reflect.TypeOf(zero).Name())
		}
		//调用方法
		return retData, err
	}, nil
}
func ContextMethodToTypeHandler[TReq, TResp any](method any) (ContextTypedHandler[TReq, TResp], error) {
	if method == nil {
		return nil, fmt.Errorf("method is nil")
	}

	methodValue := reflect.ValueOf(method)
	methodType := methodValue.Type()

	if methodType.Kind() != reflect.Func {
		return nil, fmt.Errorf("method is not a function")
	}

	methodName, _ := GetFuncName(method)

	// 验证参数数量：必须是 2 个（context.Context 和 any）
	if methodType.NumIn() != 2 ||
		methodType.NumOut() != 2 {
		return nil, fmt.Errorf("%s must have 2 input or output parameters, got %d, %d: %s", methodName,
			methodType.NumIn(), methodType.NumOut(), methodType.String())
	}

	// 验证第一个参数是 context.Context
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	if !methodType.In(0).Implements(ctxType) {
		return nil, fmt.Errorf("%s first parameter must be context.Context, got %v: %s", methodName, methodType.In(0), methodType.String())
	}
	// 验证第二个参数类型是否与 TReq 匹配
	reqType := reflect.TypeOf((*TReq)(nil)).Elem()
	if !methodType.In(1).AssignableTo(reqType) && methodType.In(1).Kind() != reflect.Interface {
		return nil, fmt.Errorf("%s second parameter must be assignable to %v, got %v: %s", methodName, reqType, methodType.In(1), methodType.String())
	}

	// 验证第一个返回值类型是否与 TResp 匹配
	respType := reflect.TypeOf((*TResp)(nil)).Elem()
	if !methodType.Out(0).AssignableTo(respType) && respType.Kind() != reflect.Interface {
		return nil, fmt.Errorf("%s first return value must be assignable to %v, got %v: %s", methodName, respType, methodType.Out(0), methodType.String())
	}

	// 验证第二个返回值是 error
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !methodType.Out(1).Implements(errorType) {
		return nil, fmt.Errorf("%s second return value must be error, got %v: %s", methodName, methodType.Out(1), methodType.String())
	}

	// 所有验证通过，返回包装后的处理函数
	return func(ctx context.Context, args TReq) (TResp, error) {
		var zeroResp TResp
		if ctx == nil {
			ctx = context.Background()
		}

		results, exeErr := FuncExecute(method, ctx, args)
		if exeErr != nil {
			return zeroResp, exeErr
		}

		// 提取返回值
		if len(results) != 2 {
			return zeroResp, fmt.Errorf("unexpected number of return values: expected 2, got %d", len(results))
		}
		if ret1, ok1 := results[0].(TResp); ok1 {
			zeroResp = ret1
		}
		exeErr = nil
		if ret2, ok2 := results[1].(error); ok2 {
			exeErr = ret2
		}
		return zeroResp, exeErr
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
