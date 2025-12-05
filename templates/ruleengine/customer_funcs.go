package ruleengine

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/samber/lo"
	"github.com/shopspring/decimal"
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
		var d decimal.Decimal
		switch v := arg.(type) {
		case float64:
			d = decimal.NewFromFloat(v)
		case float32:
			d = decimal.NewFromFloat32(v)
		case int:
			d = decimal.NewFromInt(int64(v))
		case int64:
			d = decimal.NewFromInt(v)
		case int32:
			d = decimal.NewFromInt32(v)
		case uint:
			d = decimal.NewFromUint64(uint64(v))
		case uint32:
			d = decimal.NewFromUint64(uint64(v))
		case uint64:
			d = decimal.NewFromUint64(v)
		}
		if !d.IsZero() {
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
		if intTemp, ok := conv.Int64(args[1]); ok {
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
