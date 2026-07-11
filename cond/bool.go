package cond

import "strings"

// IsBool 判断是否是bool类型
func IsBool(n any) bool {
	if _, ok := n.(bool); ok {
		return true
	}
	if nStr, ok := n.(string); ok {
		boolStr := strings.ToLower(nStr)
		return boolStr == "true" || boolStr == "false"
	}
	return false
}
