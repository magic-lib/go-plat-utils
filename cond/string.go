package cond

import (
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/json-iterator/go"
	"regexp"
	"strings"
)
import "encoding/json"

// IsUUID 是否是UUID字符串
func IsUUID(uuid string) bool {
	// 总长度应为36（32个字符 + 4个连字符）
	if len(uuid) != 36 {
		return false
	}

	// 正则：仅接受 全小写 或 全大写 的十六进制，不接受大小写混用；格式为 8-4-4-4-12
	pattern := `^(?:[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}|[0-9A-F]{8}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{4}-[0-9A-F]{12})$`

	match, err := regexp.MatchString(pattern, uuid)
	if err != nil {
		return false
	}
	return match
}

// IsJson 是否是json字符串
func IsJson(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}

	var temp any
	if err := json.Unmarshal([]byte(text), &temp); err != nil {
		return false
	}

	// 检查解析后的类型是否为对象或数组
	switch temp.(type) {
	case map[string]any: // JSON对象
		return true
	case []any: // JSON数组
		return true
	default: // 其他类型（数字、字符串、布尔值、null等）
		return false
	}
}

// IsJsonMap 是否是jsonMap字符串
func IsJsonMap(text string) bool {
	text = strings.TrimSpace(text)
	if text == "" {
		return false
	}

	var temp any
	if err := json.Unmarshal([]byte(text), &temp); err != nil {
		return false
	}
	// 检查解析后的类型是否为对象或数组
	switch temp.(type) {
	case map[string]any: // JSON对象
		return true
	default: // 其他类型（数字、字符串、布尔值、null等）
		return false
	}
}

// IsStringEmpty 是否是空字符串，不写成EmptyString，是为了好找方法
func IsStringEmpty(text string) bool {
	if len(text) == 0 {
		return true
	}
	if strings.TrimSpace(text) == "" {
		return true
	}
	return false
}

var sortJson = jsoniter.ConfigCompatibleWithStandardLibrary

// IsSameJson 判断两段JSON语义是否相等
func IsSameJson(jsonStrA, jsonStrB string) bool {
	var a any
	var b any

	if err := sortJson.Unmarshal([]byte(jsonStrA), &a); err != nil {
		return false
	}
	if err := sortJson.Unmarshal([]byte(jsonStrB), &b); err != nil {
		return false
	}

	diff := cmp.Diff(a, b)
	if diff != "" {
		fmt.Println(diff)
		return false
	}
	return true
	//return reflect.DeepEqual(a, b)
}
