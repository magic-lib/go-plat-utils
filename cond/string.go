package cond

import "regexp"
import "encoding/json"

// IsUUID 是否是UUID字符串
func IsUUID(uuid string) bool {
	// 总长度应为36（32个字符 + 4个连字符）
	if len(uuid) != 36 {
		return false
	}

	// 正则表达式：匹配小写字母和数字，格式为 8-4-4-4-12
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	match, err := regexp.MatchString(pattern, uuid)
	if err != nil {
		return false
	}
	return match
}

// IsJson 是否是json字符串
func IsJson(text string) bool {
	var temp any
	if err := json.Unmarshal([]byte(text), &temp); err != nil {
		return false
	}

	// 检查解析后的类型是否为对象或数组
	switch temp.(type) {
	case map[string]any: // JSON对象
		return true
	case []any: // JSON数组
		return false
	default: // 其他类型（数字、字符串、布尔值、null等）
		return false
	}
}
