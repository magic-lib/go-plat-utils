package cond

import (
	"reflect"
	"slices"
)

// Contains 数组中是否含有某元素
func Contains[T comparable](s []T, e T) (bool, int) {
	i := slices.Index(s, e)
	if i >= 0 {
		return true, i
	}
	return false, -1
}

// IsBytes 判断 src 是否为 []byte 类型
func IsBytes(src any) bool {
	if src == nil {
		return false
	}
	t := reflect.TypeOf(src)
	return t.Kind() == reflect.Slice && t.Elem().Kind() == reflect.Uint8
}

// IsArray 判断 是否是数组类型
func IsArray(in any) bool {
	v := reflect.ValueOf(in)
	return v.Kind() == reflect.Array || v.Kind() == reflect.Slice
}
