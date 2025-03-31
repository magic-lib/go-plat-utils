package utils

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

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

	// 将参数列表转换为reflect.Value切片
	var retValues []reflect.Value
	if len(args) > 0 {
		argSlice := make([]reflect.Value, len(args))
		funParamLen := fnType.NumIn() //如果是变长的情况

		isVariadicParam := false
		if funParamLen > 0 && fnType.IsVariadic() {
			isVariadicParam = true
		}

		for i, arg := range args {
			if arg == nil {
				if funParamLen > 0 {
					var fnArgType reflect.Type

					if isVariadicParam {
						if i < funParamLen-1 {
							fnArgType = fnType.In(i)
						} else {
							//判断最后一个参数是不是slice类型
							lastParam := fnType.In(funParamLen - 1)
							fnArgType = lastParam.Elem()
						}
					} else {
						if i < funParamLen {
							fnArgType = fnType.In(i)
						}
					}

					//需要检查是否是interface类型
					if fnArgType.Kind() == reflect.Interface {
						argSlice[i] = reflect.Zero(reflect.TypeOf((*any)(nil)).Elem())
						continue
					}
					if fnArgType.Kind() == reflect.Ptr {
						argSlice[i] = reflect.Zero(fnArgType)
						continue
					}
				}
			}
			argSlice[i] = reflect.ValueOf(arg)
		}
		// 调用函数，捕获panic
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("[FuncExecute] Recovered from panic:", r)
			}
		}()

		{ //arg与原方程数量不对，这里需要特殊处理一下
			if len(argSlice) != funParamLen {
				//判断最后一个是不是变长参数
				if funParamLen > 0 {

					paramLengthError := false
					if isVariadicParam {
						if len(argSlice) < funParamLen-1 {
							paramLengthError = true
						}
					} else {
						if len(argSlice) > funParamLen {
							argSlice = argSlice[0:funParamLen]
						} else {
							paramLengthError = true
						}
					}

					if paramLengthError {
						return []any{}, fmt.Errorf("[FuncExecute] param length error: %d, func param length: %d",
							len(argSlice), funParamLen)
					}
				} else {
					// 如果原方法没有参数，则这里设为nil
					argSlice = nil
				}
			}
		}

		// 调用函数并返回结果
		retValues = fnValue.Call(argSlice)
	} else {
		retValues = fnValue.Call(nil)
	}

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
