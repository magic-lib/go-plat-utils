package utils

import (
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	"reflect"
	"slices"
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
// Deprecated: 该方法已废弃
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

// AsArray 将可变参数列表为数组
func AsArray[T comparable](el ...T) []T {
	return el
}

// ArrayContains 判断数组中是否含有指定元素
// Deprecated: 该方法已废弃
func ArrayContains[T comparable](arr []T, el T) bool {
	return slices.Contains(arr, el)
}

// ArrayCompare 数组比较
// 依次返回，新增值，删除值，以及不变值
func ArrayCompare[T comparable](newArr []T, oldArr []T) ([]T, []T, []T) {
	newSet := make(map[T]bool)
	oldSet := make(map[T]bool)

	// 将新数组和旧数组的元素分别添加到对应的哈希集合中
	for _, elem := range newArr {
		newSet[elem] = true
	}
	for _, elem := range oldArr {
		oldSet[elem] = true
	}

	var (
		added      []T
		deleted    []T
		unmodified []T
	)

	// 遍历新数组，根据元素是否存在于旧数组进行分类
	for _, elem := range newArr {
		if oldSet[elem] {
			unmodified = append(unmodified, elem)
		} else {
			added = append(added, elem)
		}
	}

	// 遍历旧数组，找出被删除的元素
	for _, elem := range oldArr {
		if !newSet[elem] {
			deleted = append(deleted, elem)
		}
	}

	return added, deleted, unmodified
}

// ArrayToMap 数组转为map
// Deprecated: 该方法已废弃 lo.KeyBy
// keyFunc key的主键
func ArrayToMap[T any, K comparable](arr []T, keyFunc func(val T) K) map[K]T {
	return lo.KeyBy(arr, keyFunc)
}

// ArrayMap 数组映射，即将一数组元素通过映射函数转换为另一数组
func ArrayMap[T any, K any](arr []T, mapFunc func(val T) K) []K {
	res := make([]K, len(arr))
	for i, val := range arr {
		res[i] = mapFunc(val)
	}
	return res
}

// ArrayMapFilter 数组映射并过滤，若mapFunc返回false，则不映射该元素到新数组。
func ArrayMapFilter[T any, K any](arr []T, mapFilterFunc func(val T) (K, bool)) []K {
	res := make([]K, 0)
	for _, val := range arr {
		mapRes, needMap := mapFilterFunc(val)
		if needMap {
			res = append(res, mapRes)
		}
	}
	return res
}

// ArrayChunk 将数组或切片按固定大小分割成小数组
func ArrayChunk[T any](arr []T, chunkSize int) [][]T {
	var chunks [][]T
	for i := 0; i < len(arr); i += chunkSize {
		end := i + chunkSize
		if end > len(arr) {
			end = len(arr)
		}
		chunks = append(chunks, arr[i:end])
	}
	return chunks
}

// ArraySplit 将数组切割为指定个数的子数组，并尽可能均匀
func ArraySplit[T any](arr []T, numGroups int) [][]T {
	if numGroups > len(arr) {
		numGroups = len(arr)
	}

	arrayLen := len(arr)
	if arrayLen < 1 {
		return [][]T{}
	}
	// 计算每个子数组的大小
	size := arrayLen / numGroups
	remainder := arrayLen % numGroups

	// 创建一个存放子数组的切片
	subArrays := make([][]T, numGroups)

	// 分割数组为子数组
	start := 0
	for i := range subArrays {
		subSize := size
		if i < remainder {
			subSize++
		}
		subArrays[i] = arr[start : start+subSize]
		start += subSize
	}

	return subArrays
}

// ArrayReduce reduce操作
func ArrayReduce[T any, V any](arr []T, initialValue V, reducer func(V, T) V) V {
	value := initialValue
	for _, a := range arr {
		value = reducer(value, a)
	}
	return value
}

// ArrayRemoveFunc 数组元素移除操作
func ArrayRemoveFunc[T any](arr []T, isDeleteFunc func(T) bool) []T {
	var newArr []T
	for _, a := range arr {
		if !isDeleteFunc(a) {
			newArr = append(newArr, a)
		}
	}
	return newArr
}

// ArrayRemoveBlank 移除元素中的空元素
func ArrayRemoveBlank[T any](arr []T) []T {
	return ArrayRemoveFunc(arr, func(val T) bool {
		return cond.IsZero(val)
	})
}

// ArrayDeduplicate 数组元素去重
func ArrayDeduplicate[T comparable](arr []T) []T {
	encountered := map[T]bool{}
	result := make([]T, 0)

	for v := range arr {
		if !encountered[arr[v]] {
			encountered[arr[v]] = true
			result = append(result, arr[v])
		}
	}

	return result
}

// ArrayAnyMatches 给定字符串是否包含指定数组中的任意字符串， 如：["time", "date"] , substr : timestamp，返回true
func ArrayAnyMatches(arr []string, subStr string) bool {
	for _, itm := range arr {
		if strings.Contains(subStr, itm) {
			return true
		}
	}
	return false
}

// ArrayFilter 过滤函数，根据提供的条件函数将切片中的元素进行过滤
func ArrayFilter[T any](array []T, fn func(T) bool) []T {
	var filtered []T
	for _, val := range array {
		if fn(val) {
			filtered = append(filtered, val)
		}
	}
	return filtered
}

// AnyMatch 查找数组中是否存在满足条件的元素
func AnyMatch[T any](array []T, fn func(T) bool) bool {
	for _, val := range array {
		if fn(val) {
			return true
		}
	}
	return false
}
