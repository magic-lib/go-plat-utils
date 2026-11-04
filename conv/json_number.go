package conv

import (
	"encoding/json"

	jsoniterForNil "github.com/magic-lib/go-plat-utils/internal/jsoniter/go"
)

var (
	// jsonUseNumber 与默认配置一致，但数字解析为 json.Number，避免大整数先变成 float64 而丢精度
	jsonUseNumber = jsoniterForNil.Config{EscapeHTML: true, UseNumber: true}.Froze()
	// jsonSortedNoEscape 按 key 排序且不转义 HTML，在 jcs 无法处理大整数时作为替代，保持输出的 key 顺序与转义方式一致
	jsonSortedNoEscape = jsoniterForNil.Config{SortMapKeys: true}.Froze()
)

// maxSafeInteger float64 能精确表示的最大整数 2^53
const maxSafeInteger = 1 << 53

// normalizeJsonNumber 递归处理 UseNumber 解析出的 json.Number：
// 超出 ±2^53 的整数保留为 json.Number 以免丢精度，其余数字仍转为 float64，与原有解析结果的类型保持一致。
// 返回处理后的值，以及是否保留了 json.Number。
func normalizeJsonNumber(v any) (any, bool) {
	switch val := v.(type) {
	case json.Number:
		if i, err := val.Int64(); err == nil {
			if i > maxSafeInteger || i < -maxSafeInteger {
				return val, true
			}
			return float64(i), false
		}
		if isIntegerString(val.String()) { // 超出 int64 范围的整数
			return val, true
		}
		f, err := val.Float64()
		if err != nil {
			return val, true
		}
		return f, false
	case map[string]any:
		keep := false
		for k, one := range val {
			newOne, oneKeep := normalizeJsonNumber(one)
			val[k] = newOne
			keep = keep || oneKeep
		}
		return val, keep
	case []any:
		keep := false
		for i, one := range val {
			newOne, oneKeep := normalizeJsonNumber(one)
			val[i] = newOne
			keep = keep || oneKeep
		}
		return val, keep
	default:
		return v, false
	}
}

// isIntegerString 是否为十进制整数字符串（可带负号）
func isIntegerString(s string) bool {
	if len(s) > 0 && s[0] == '-' {
		s = s[1:]
	}
	if s == "" {
		return false
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
