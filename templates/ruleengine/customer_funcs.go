package ruleengine

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"reflect"
	"strings"
	"time"
)

// customerFunc 自定义方法列表
type customerFunc struct {
}

func (r *customerFunc) getAllDecimalList(args ...any) []decimal.Decimal {
	decimalList := make([]decimal.Decimal, 0)
	for _, arg := range args {
		d, err := conv.Convert[decimal.Decimal](arg)
		if err == nil {
			decimalList = append(decimalList, d)
		}
	}
	return decimalList
}

// relationByNumber 两数相互运算
func (r *customerFunc) relationByNumber(f func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal, args ...any) float64 {
	decimalList := r.getAllDecimalList(args...)
	if len(decimalList) == 0 {
		return 0
	}
	var total decimal.Decimal
	for i, d := range decimalList {
		if i == 0 {
			total = d
			continue
		}
		total = f(total, d)
	}
	return total.InexactFloat64()
}

// Add 两数相加
func (r *customerFunc) Add(args ...any) (any, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Add(d2)
	}, args...), nil
}

// Sub 两数相减
func (r *customerFunc) Sub(args ...any) (any, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Sub(d2)
	}, args...), nil
}

// Mul 两数相乘
func (r *customerFunc) Mul(args ...any) (any, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Mul(d2)
	}, args...), nil
}

// Div 两数相除
func (r *customerFunc) Div(args ...any) (any, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Div(d2)
	}, args...), nil
}

// Has 数组是否包含某元素
func (r *customerFunc) Has(args ...any) (any, error) {
	if len(args) != 2 {
		if len(args) == 1 {
			return false, nil
		}
		if len(args) > 2 {
			//这是一个bug，会将数组变成动态参数
			arg1 := args[0 : len(args)-1]
			return r.Has(arg1, args[len(args)-1])
		}
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	listInterface := args[0]
	item, _ := conv.Convert[string](args[1])
	listType := reflect.TypeOf(listInterface)
	listValue := reflect.ValueOf(listInterface)
	if listType.Kind() == reflect.Slice {
		for i := 0; i < listValue.Len(); i++ {
			if conv.String(listValue.Index(i).Interface()) == item {
				return true, nil
			}
		}
	} else if listType.Kind() == reflect.String {
		//这种字符串的格式：`["a", "b"]`
		list := make([]any, 0)
		_ = conv.Unmarshal(listInterface, &list)
		for _, v := range list {
			if conv.String(v) == item {
				return true, nil
			}
		}
	}
	return false, nil
}

// In 是否存在某数组中
func (r *customerFunc) In(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	return r.Has(args[1], args[0])
}

// Between 是否在某个范围内 Between(num, "[3,6]")
func (r *customerFunc) Between(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	rangeTemplate := conv.String(args[1])
	rangeTemplate = strings.Join(strings.Fields(rangeTemplate), "")

	rangeData, err := parseRange(rangeTemplate)
	if err != nil {
		return false, err
	}
	if rangeData == nil {
		return false, nil
	}
	num, err := conv.Convert[float64](args[0])
	if err != nil {
		return false, err
	}
	return rangeData.match(num), nil
}

// Is 是否是某一个类型
func (r *customerFunc) Is(args ...any) (any, error) {
	if len(args) <= 1 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	typeName := conv.String(args[0])
	typeName = strings.ToLower(typeName)
	if typeName == "nil" {
		return cond.IsNil(args[1]), nil
	}
	if typeName == "zero" {
		return cond.IsZero(args[1]), nil
	}
	if typeName == "number" {
		return cond.IsNumeric(args[1]), nil
	}
	if typeName == "time" {
		return cond.IsTime(conv.String(args[1])), nil
	}
	return false, fmt.Errorf("不支持的格式：%s", typeName)
}
func (r *customerFunc) As(args ...any) (any, error) {
	if len(args) <= 1 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	typeName := conv.String(args[0])
	typeName = strings.ToLower(typeName)
	if typeName == "nil" {
		return nil, nil
	}
	if typeName == "string" {
		return conv.String(args[1]), nil
	}
	if typeName == "int" {
		if intTemp, err1 := conv.Convert[int](args[1]); err1 == nil {
			return intTemp, nil
		}
		return 0, fmt.Errorf("参数不是int类型：%v", args[1])
	}
	if typeName == "int64" {
		if intTemp, err := conv.Convert[int64](args[1]); err == nil {
			return intTemp, nil
		}
		return 0, fmt.Errorf("参数不是int64类型：%v", args[1])
	}
	if typeName == "bool" {
		if boolTemp, err1 := conv.Convert[bool](args[1]); err1 == nil {
			return boolTemp, nil
		}
		return false, fmt.Errorf("参数不是bool类型：%v", args[1])
	}
	if typeName == "time" {
		if timeTemp, err1 := conv.Convert[time.Time](args[1]); err1 == nil {
			return timeTemp, nil
		}
		return time.Time{}, fmt.Errorf("参数不是time类型：%v", args[1])
	}
	return false, fmt.Errorf("不支持的格式：%s", typeName)
}
func (r *customerFunc) Replace(args ...any) (any, error) {
	if len(args) <= 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	oldStr := conv.String(args[1])
	newStr := conv.String(args[2])
	num := -1
	if len(args) >= 4 {
		if numTemp, err1 := conv.Convert[int](args[3]); err1 == nil {
			num = numTemp
		}
	}
	return strings.Replace(conv.String(args[0]), oldStr, newStr, num), nil
}
func (r *customerFunc) Split(args ...any) (any, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	splitArr := make([]string, 0)
	lo.ForEach(args[1:], func(item any, _ int) {
		splitArr = append(splitArr, conv.String(item))
	})
	return utils.Split(conv.String(args[0]), splitArr), nil
}
func (r *customerFunc) Contains(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	return strings.Contains(conv.String(args[0]), conv.String(args[1])), nil
}
func (r *customerFunc) JsonGet(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	jsonStr := conv.String(args[0])
	pathStr := conv.String(args[1])
	gResult := gjson.Get(jsonStr, pathStr)
	if !gResult.Exists() {
		return "", nil
	}
	return gResult.Value(), nil
}

// Join 连接字符串，第一个字符为连接符
// Join("/", "a", "b", "c")
func (r *customerFunc) Join(args ...any) (any, error) {
	sep := ""
	var dataList []any
	if len(args) == 0 {
		return "", nil
	} else if len(args) == 1 {
		if list, ok := args[0].([]any); ok {
			dataList = list
		} else {
			return conv.String(args[0]), nil
		}
	} else if len(args) == 2 {
		sep = conv.String(args[0])
		if list, ok := args[1].([]string); ok {
			dataList = lo.Map(list, func(item string, index int) any {
				return any(item)
			})
		} else {
			dataList = args[1:]
		}
	} else {
		sep = conv.String(args[0])
		dataList = args[1:]
	}
	retStr := make([]string, 0)
	lo.ForEach(dataList, func(item any, _ int) {
		retStr = append(retStr, conv.String(item))
	})
	return strings.Join(retStr, sep), nil
}

// If 三元运算符
func (r *customerFunc) If(args ...any) (any, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("ternary function requires exactly 3 arguments: condition, trueValue, falseValue")
	}
	// 第一个参数必须是布尔类型
	condition, ok := args[0].(bool)
	if !ok {
		return nil, fmt.Errorf("first argument to ternary must be a boolean")
	}
	if condition {
		return args[1], nil
	}
	return args[2], nil
}

// Switch 多分支选择，避免多重嵌套 If。
// 用法: Switch(value, case1, result1, case2, result2, ..., defaultValue)
//   - args[0]         : 待匹配的值
//   - 后续成对        : (候选值, 命中返回值)
//   - 最后一个参数     : 都不命中时的默认值
//
// 例: Switch(status, "A", "苹果", "B", "香蕉", "未知")
func (r *customerFunc) Switch(args ...any) (any, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("switch requires at least 3 arguments: value, [candidate, result]..., defaultValue")
	}
	if (len(args)-1)%2 == 0 {
		return nil, fmt.Errorf("switch arguments must be value, then pairs of (candidate, result), then a defaultValue")
	}
	value := args[0]
	// 逐对比较（用 conv.Equal 做跨类型宽松相等，如 "1" 与 1）
	for i := 1; i < len(args)-1; i += 2 {
		if cond.IsEqual(value, args[i]) {
			return args[i+1], nil
		}
	}
	// 全部未命中，返回默认值
	return args[len(args)-1], nil
}

// SwitchExpr 按布尔条件分支：SwitchExpr(cond1, res1, cond2, res2, ..., defaultValue)
// 例: SwitchExpr(score >= 90, "优", score >= 80, "良", "及格")
func (r *customerFunc) SwitchExpr(args ...any) (any, error) {
	if len(args) < 3 || (len(args)-1)%2 != 0 {
		return nil, fmt.Errorf("switchExpr requires value-less pairs: (condition, result)..., defaultValue")
	}
	for i := 0; i < len(args)-1; i += 2 {
		condBool, ok := args[i].(bool)
		if !ok {
			return nil, fmt.Errorf("switchExpr condition at position %d must be bool", i)
		}
		if condBool {
			return args[i+1], nil
		}
	}
	return args[len(args)-1], nil
}

func (r *customerFunc) Max(args ...any) (any, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("参数为空")
	}
	var currentNum float64
	var found bool
	var notFirst bool
	lo.ForEach(args, func(item any, _ int) {
		one, err := conv.Convert[float64](item)
		if err != nil {
			return
		}
		found = true
		if !notFirst {
			currentNum = one
			notFirst = true
			return
		}
		if one > currentNum {
			currentNum = one
		}
	})
	if !found {
		return 0, fmt.Errorf("没有找到数字")
	}

	return currentNum, nil
}

func (r *customerFunc) Min(args ...any) (any, error) {
	if len(args) == 0 {
		return 0, fmt.Errorf("参数为空")
	}
	var currentNum float64
	var found bool
	var notFirst bool
	lo.ForEach(args, func(item any, _ int) {
		one, err := conv.Convert[float64](item)
		if err != nil {
			return
		}
		found = true
		if !notFirst {
			currentNum = one
			notFirst = true
			return
		}
		if one < currentNum {
			currentNum = one
		}
	})
	if !found {
		return 0, fmt.Errorf("没有找到数字")
	}
	return currentNum, nil
}

func (r *customerFunc) getDecimalBaseAndNum(args ...any) (int64, decimal.Decimal, error) {
	var initNum float64 = 0
	var initBase int64 = 0
	initDecimalNum := decimal.NewFromFloat(initNum)

	var baseNum any = 10
	var numDecimal any = 0

	if len(args) == 0 {
		return initBase, initDecimalNum, fmt.Errorf("参数数量不对，需要2个参数：位数和数字")
	} else if len(args) == 1 {
		numDecimal = args[0]
	} else if len(args) == 2 {
		baseNum = args[0]
		numDecimal = args[1]
	}

	// 获取基数参数（10, 100, 1000等）
	base, err := conv.Convert[int64](baseNum)
	if err != nil {
		return initBase, initDecimalNum, fmt.Errorf("基数参数转换失败：%v", baseNum)
	}

	// 验证基数是否为10的幂次方
	if !isValidBase(base) {
		return initBase, initDecimalNum, fmt.Errorf("基数必须是10的幂次方（1, 10, 100, 1000...），当前值：%v", base)
	}

	// 获取数字参数
	num, err := conv.Convert[float64](numDecimal)
	if err != nil {
		return initBase, initDecimalNum, fmt.Errorf("数字参数转换失败：%v", numDecimal)
	}

	return base, decimal.NewFromFloat(num), nil
}

// CeilToDigit 指定位数向上取整，默认是10进位
func (r *customerFunc) CeilToDigit(args ...any) (any, error) {
	intBase, decimalNum, err := r.getDecimalBaseAndNum(args...)
	if err != nil {
		return 0, err
	}

	decimalBase := decimal.NewFromInt(intBase)
	result := decimalNum.Div(decimalBase).Ceil().Mul(decimalBase)

	return result.InexactFloat64(), nil
}

// FloorToDigit 指定位数向下取整，默认是10进位
func (r *customerFunc) FloorToDigit(args ...any) (any, error) {
	intBase, decimalNum, err := r.getDecimalBaseAndNum(args...)
	if err != nil {
		return 0, err
	}

	decimalBase := decimal.NewFromInt(intBase)

	// 除以基数，向上取整，再乘以基数
	result := decimalNum.Div(decimalBase).Floor().Mul(decimalBase)

	return result.InexactFloat64(), nil
}

// isValidBase 验证基数是否为10的幂次方（1, 10, 100, 1000...）
func isValidBase(base int64) bool {
	if base <= 0 {
		return false
	}
	if base == 1 {
		return true
	}
	for base >= 10 {
		base = base / 10
	}
	return base == 1
}
