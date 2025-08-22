package utils

import (
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	"reflect"
	"strconv"
	"strings"
)

// Split 字符串，通过多个分隔符进行分隔
func Split(s string, sep []string) []string {
	if s == "" {
		return []string{}
	}
	if len(sep) == 0 {
		return []string{s}
	}
	sepStr := sep[0]
	for i, one := range sep {
		if i == 0 {
			continue
		}
		s = strings.Replace(s, one, sepStr, -1)
	}
	return strings.Split(s, sepStr)
}

//func getKind(kind reflect.Kind) reflect.Kind {
//	switch {
//	case kind >= reflect.Int && kind < reflect.Int64:
//		return reflect.Int
//	case kind >= reflect.Uint && kind < reflect.Uint64:
//		return reflect.Uint
//	case kind >= reflect.Float32 && kind <= reflect.Float64:
//		return reflect.Float64
//	default:
//		return kind
//	}
//}

// Join 任意数组类型，通过分隔符连接成字符串
func Join(i any, s string) string {
	if i == nil {
		return ""
	}
	i = reflect.Indirect(reflect.ValueOf(i)).Interface()
	r := make([]string, 0)
	switch i.(type) {
	case [][]byte:
		temp := i.([][]byte)
		for _, one := range temp {
			r = append(r, string(one))
		}
		return strings.Join(r, s)
	case []string:
		temp := i.([]string)
		return strings.Join(temp, s)
	case []int:
		temp := i.([]int)
		for _, one := range temp {
			r = append(r, strconv.Itoa(one))
		}
		return strings.Join(r, s)
	case []int64:
		temp := i.([]int64)
		for _, one := range temp {
			r = append(r, strconv.FormatInt(one, 10))
		}
		return strings.Join(r, s)
	case []any:
		temp := i.([]any)
		for _, one := range temp {
			r = append(r, conv.String(one))
		}
		return strings.Join(r, s)
	case string:
		return conv.String(i)
	}
	return conv.String(i)

}

// AppendUniq 给数组添加不重复的对象
func AppendUniq[T comparable](slice []T, elems ...T) []T {
	slice = append(slice, elems...)
	return lo.Uniq(slice)
}

// RemoveItem 移除一个元素
func RemoveItem[T comparable](slice []T, oneElem T) []T {
	return lo.Filter(slice, func(item T, index int) bool {
		return item != oneElem
	})
}

// SliceDiff find elements that in slice1 but not in slice2
func SliceDiff[T comparable](slice1 []T, slice2 []T) []T {
	diff := make([]T, 0)
	for _, s1 := range slice1 {
		found := false
		for _, s2 := range slice2 {
			if reflect.TypeOf(s1).Name() == reflect.TypeOf(s2).Name() &&
				reflect.ValueOf(s1).Interface() == reflect.ValueOf(s2).Interface() {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, s1)
		}
	}

	return diff
}

// NextByRing 从一个循环里取下一个，数组会构成一个圈
func NextByRing[K comparable, V any](
	vsList []V,
	last V,
	key func(this V) K,             // 用于元素唯一标识的提取函数，判断是否想等使用
	next func(this V, last V) bool, // 如果未找到key，则用此来判断元素下一个元素的条件函数
) V {
	// 处理空切片情况
	vLen := len(vsList)
	if vLen == 0 {
		return last
	}
	// 只有一个元素时直接返回该元素
	if vLen == 1 || cond.IsNil(last) {
		return vsList[0]
	}

	idx := -1
	var nextOne V
	var foundNext bool
	lo.ForEachWhile(vsList, func(item V, index int) bool {
		if cond.IsNil(item) {
			return true
		}
		if key(item) == key(last) {
			idx = index
			return false
		}
		if !foundNext {
			// 只取第一个符合条件的元素
			if next(item, last) {
				foundNext = true
				nextOne = item
			}
		}
		return true
	})
	if idx >= 0 {
		return vsList[(idx+1)%vLen]
	}
	// 返回找到的候选元素或列表第一个元素
	if foundNext {
		return nextOne
	}

	return vsList[0]
}
