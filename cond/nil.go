package cond

import (
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const eps = 1e-9

// IsNil 判断是否为空
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	value := reflect.ValueOf(v)
	kind := value.Kind()
	if kind == reflect.Ptr ||
		kind == reflect.Chan ||
		kind == reflect.Func ||
		kind == reflect.UnsafePointer ||
		kind == reflect.Map ||
		kind == reflect.Interface ||
		kind == reflect.Slice {
		if value.IsNil() {
			return true
		}
	} else if kind == reflect.Struct {
		for i := 0; i < value.NumField(); i++ {
			if !value.Field(i).IsZero() {
				return false
			}
		}
		return true
	} else if kind == reflect.Invalid {
		return true
	} else if kind == reflect.String ||
		kind == reflect.Bool {
		return false
	} else if kind == reflect.Int ||
		kind == reflect.Int8 ||
		kind == reflect.Int16 ||
		kind == reflect.Int32 ||
		kind == reflect.Int64 {
		return false
	} else if kind == reflect.Float32 ||
		kind == reflect.Float64 {
		return false
	} else if kind == reflect.Uint ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64 {
		return false
	} else if kind == reflect.Complex64 ||
		kind == reflect.Complex128 {
		return false
	}

	if !value.IsValid() {
		return true
	}

	if value == reflect.ValueOf(nil) {
		return true
	}

	return false
}

// IsZero 判断变量是否为零值
func IsZero(val any) bool {
	if val == nil {
		return true
	}
	//常用的类型，提高执行效率
	switch val.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
	case complex64, complex128:
		return val == 0
	case float32, float64:
		valFloat32, ok := val.(float32)
		if ok {
			return math.Abs(float64(valFloat32)) < eps
		}
		valFloat64, ok := val.(float64)
		if ok {
			return math.Abs(valFloat64) < eps
		}
	case bool:
		return val == false
	case string:
		return val == ""
	case time.Time:
		if valTime, ok := val.(time.Time); ok {
			return IsTimeEmpty(valTime)
		}
	default:
	}

	rValue := reflect.ValueOf(val)
	switch rValue.Kind() {
	case reflect.Bool:
		return !rValue.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rValue.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rValue.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rValue.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return rValue.IsNil()
	case reflect.Slice, reflect.Map, reflect.String:
		return rValue.Len() == 0
	default:
	}

	return reflect.ValueOf(val).IsZero()
}

// IsPointer 判断是否是指针类型
func IsPointer(val any) bool {
	if val == nil {
		return false // nil没有类型，不算指针
	}
	// 获取值的类型并判断其Kind是否为指针
	return reflect.TypeOf(val).Kind() == reflect.Ptr
}

// errorTypeCache 缓存「类型 → IsError 判定结果」。
// NumMethod/Method(i) 走反射且 Method(i) 会构造 reflect.Value（有分配），
// 同一类型的判定结果恒定不变，按 reflect.Type 缓存可避免重复计算。
var errorTypeCache sync.Map // map[reflect.Type]bool

// IsError 判断 val 是否是一个「纯粹的」错误对象，需要同时满足：
//  1. val 实现了 error 接口（方法集内含签名为 Error() string 的方法）；
//  2. 该类型的**导出**方法只允许是 Error，以及可选的 Unwrap。
//
// 之所以加第 2 条：Go 的接口是隐式实现，业务结构体只要碰巧定义了
// Error() string 就会被 val.(error) 命中；而这类结构体通常还带有
// GetXxx/ToXxx/MarshalJSON 等其它导出方法，据此可把它们与真正的错误区分开。
//
// Unwrap 之所以放行：它是 errors 包约定的错误链方法（Unwrap() error），
// 带它说明该类型仍是一条纯粹的错误链的成员，而非某个碰巧有 Error 方法的业务对象。
// errors.New/fmt.Errorf（含 %w 包装）/errors.Join 的结果都会被判为 true。
//
// 未导出方法不计入统计；嵌入结构体提升上来的导出方法计入。
//
// 注意：判定基于 val 的动态类型本身。若 Error 使用指针接收者，则传值 T{}
// 连条件 1 都不满足（返回 false），需传 &T{}；这与 Go 自身的方法集规则一致。
func IsError(val any) bool {
	if val == nil {
		return false
	}
	if _, ok := val.(error); !ok {
		return false
	}

	typ := reflect.TypeOf(val)
	if cached, ok := errorTypeCache.Load(typ); ok {
		return cached.(bool)
	}
	ret := hasOnlyErrorMethods(typ)
	errorTypeCache.Store(typ, ret)
	return ret
}

// hasOnlyErrorMethods 判断 typ 的导出方法是否只由 Error 和 Unwrap 组成，且 Error 必须存在
func hasOnlyErrorMethods(typ reflect.Type) bool {
	hasError := false
	for i := 0; i < typ.NumMethod(); i++ {
		method := typ.Method(i)
		if !method.IsExported() { // 未导出方法忽略
			continue
		}
		switch method.Name {
		case "Error":
			hasError = true
		case "Unwrap":
			// 错误链约定方法，放行
		default:
			return false // 出现第 3 方导出方法，立即结束
		}
	}
	return hasError
}

// IsEqual 宽松相等判断，支持跨类型比较，适用于规则引擎中的值匹配场景：
//   - 类型相同：使用 reflect.DeepEqual
//   - 均为数值或可解析为数值的字符串：统一转为 float64 后比较
//     （如 int64(1) 与 float64(1)、字符串 "1" 与数字 1 均视为相等）
//   - 一边为字符串、另一边为布尔：按 "true"/"false" 比较
//   - 其余类型不同的情况：退回字符串形态比较（如 decimal 的 String() 表示）
//
// 注意：浮点数比较使用误差容差，避免 1.0 与 1.0000000001 误判不等。
func IsEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}

	// 1) 同类型直接 DeepEqual（最快，覆盖 struct/map/slice/同类型数值等）
	if reflect.TypeOf(a) == reflect.TypeOf(b) {
		return reflect.DeepEqual(a, b)
	}

	// 2) 数值 / 数字字符串 → 转 float64 比较
	af, aOk := toFloatValue(a)
	bf, bOk := toFloatValue(b)
	if aOk && bOk {
		return math.Abs(af-bf) < 1e-9
	}

	// 3) 一边字符串、一边布尔
	if bs, ok := b.(string); ok {
		if ab, ok2 := a.(bool); ok2 {
			return strings.EqualFold(bs, strconv.FormatBool(ab))
		}
	}
	if as, ok := a.(string); ok {
		if bb, ok2 := b.(bool); ok2 {
			return strings.EqualFold(as, strconv.FormatBool(bb))
		}
	}

	// 4) 退回字符串形态比较（覆盖 decimal、"1.0" 与 1 等已在上步处理外的其余情况）
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// toFloatValue 将数值或可解析的数字字符串转为 float64。
// 返回 (value, true)；无法转换时返回 (0, false)。
func toFloatValue(v any) (float64, bool) {
	switch n := v.(type) {
	case int:
		return float64(n), true
	case int8:
		return float64(n), true
	case int16:
		return float64(n), true
	case int32:
		return float64(n), true
	case int64:
		return float64(n), true
	case uint:
		return float64(n), true
	case uint8:
		return float64(n), true
	case uint16:
		return float64(n), true
	case uint32:
		return float64(n), true
	case uint64:
		return float64(n), true
	case float32:
		return float64(n), true
	case float64:
		return n, true
	case string:
		s := strings.TrimSpace(n)
		if s == "" {
			return 0, false
		}
		// 复用 IsNumeric 判断，再用 strconv 解析
		if IsNumeric(s) {
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return f, true
			}
		}
		return 0, false
	default:
		return 0, false
	}
}
