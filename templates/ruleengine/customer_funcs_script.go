package ruleengine

import (
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/dop251/goja"
	"github.com/magic-lib/go-plat-utils/conv"
)

// jsRunTimeout 单个 Js 脚本的最长执行时间，超时中断，避免死循环把规则计算卡住
const jsRunTimeout = time.Second

// JavaScript 用 goja 按 JavaScript 语义执行脚本：JavaScript('脚本', 参数1, 参数2, ...)。
//   - 第 1 个参数是脚本，其余参数依次注入为 arg0、arg1……，也可以用数组 args 访问
//   - 参数是 map（key 为字符串）时，还会把它的每个 key 直接注入为同名变量，脚本里可直接写 key，
//     例如 JavaScript('nid.substr(-6, 1) == 4', user)，其中 user = {"nid": "151024/92/1"}；
//     同名时 map 的 key 优先，多个 map 有相同 key 时后面的覆盖前面的
//   - 适合需要 JS 宽松语义的场景，例如 1 == "1" 为 true、"a" + 5 得到 "a5"
//   - 结果为数字时统一转成 float64，与规则里的数字字面量类型一致，可直接用 == 比较
//
// 例：JavaScript('arg0.substr(-6, 1) == 4', nid) && a > 2。
// 注意：规则引擎的字符串遇到单引号或双引号都会结束，脚本里要写字符串时，引号必须用反斜杠转义，
// 例如 JavaScript('arg0 == \"VIP\"', level)、JavaScript('\'a\' + 5')。
// 每次调用都新建 goja.Runtime，可并发使用；执行超过 jsRunTimeout 会被中断并返回错误。
func (r *customerFunc) JavaScript(args ...any) (any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("参数数量不对：%v", args)
	}
	script, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("JavaScript 第一个参数必须是脚本字符串：%v", args[0])
	}

	vm := goja.New()
	jsArgs := args[1:]
	for i, one := range jsArgs {
		if err := vm.Set("arg"+strconv.Itoa(i), one); err != nil {
			return nil, fmt.Errorf("JavaScript 注入参数 arg%d 失败: %w", i, err)
		}
	}
	if err := vm.Set("args", jsArgs); err != nil {
		return nil, fmt.Errorf("JavaScript 注入参数 args 失败: %w", err)
	}
	// map 参数：按 key 注入为同名变量，脚本里不用再写 arg0.xxx
	for _, one := range jsArgs {
		if err := setJsMapVars(vm, one); err != nil {
			return nil, err
		}
	}

	timer := time.AfterFunc(jsRunTimeout, func() {
		vm.Interrupt(fmt.Sprintf("执行超过 %s", jsRunTimeout))
	})
	defer timer.Stop()

	value, err := vm.RunString(script)
	if err != nil {
		return nil, fmt.Errorf("JavaScript 执行出错 %q: %w", script, err)
	}
	return jsExportValue(value), nil
}

// setJsMapVars 参数是 key 为字符串的 map 时，把每个 key 注入为同名 JS 变量；其他类型不处理
func setJsMapVars(vm *goja.Runtime, arg any) error {
	mapValue := reflect.ValueOf(arg)
	if mapValue.Kind() != reflect.Map || mapValue.Type().Key().Kind() != reflect.String {
		return nil
	}
	iter := mapValue.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		if err := vm.Set(key, iter.Value().Interface()); err != nil {
			return fmt.Errorf("JavaScript 注入变量 %s 失败: %w", key, err)
		}
	}
	return nil
}

// jsExportValue 把 goja 结果导出为 Go 值；整数会被 goja 导出成 int64，统一转成 float64
func jsExportValue(value goja.Value) any {
	if value == nil || goja.IsUndefined(value) || goja.IsNull(value) {
		return nil
	}
	ret := value.Export()
	switch ret.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32:
		if f, err := conv.Convert[float64](ret); err == nil {
			return f
		}
	}
	return ret
}
