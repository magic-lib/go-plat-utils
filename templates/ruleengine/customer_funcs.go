package ruleengine

import (
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/shopspring/decimal"
	"reflect"
)

// customerFunc 自定义方法列表
type customerFunc struct {
}

func (r *customerFunc) getAllDecimalList(args ...interface{}) []decimal.Decimal {
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

// relationByNumber 两数相加
func (r *customerFunc) relationByNumber(f func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal, args ...interface{}) float64 {
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
	ret, _ := total.Float64()
	return ret
}

// AddByNumber 两数相加
func (r *customerFunc) AddByNumber(args ...interface{}) (interface{}, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Add(d2)
	}, args...), nil
}

// SubByNumber 两数相减
func (r *customerFunc) SubByNumber(args ...interface{}) (interface{}, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Sub(d2)
	}, args...), nil
}

// MulByNumber 两数相乘
func (r *customerFunc) MulByNumber(args ...interface{}) (interface{}, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Mul(d2)
	}, args...), nil
}

// DivByNumber 两数相除
func (r *customerFunc) DivByNumber(args ...interface{}) (interface{}, error) {
	return r.relationByNumber(func(d1 decimal.Decimal, d2 decimal.Decimal) decimal.Decimal {
		return d1.Div(d2)
	}, args...), nil
}

// Has 数组是否包含某元素
func (r *customerFunc) Has(args ...interface{}) (interface{}, error) {
	if args == nil || len(args) != 2 {
		return false, nil
	}
	listInterface := args[0]
	item := conv.String(args[1])
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
func (r *customerFunc) In(args ...interface{}) (interface{}, error) {
	if args == nil || len(args) != 2 {
		return false, nil
	}
	return r.Has(args[1], args[0])
}
