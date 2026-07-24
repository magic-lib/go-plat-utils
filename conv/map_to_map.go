package conv

// cloneSrcAnyMap 若 src 是 map[string]any 或 *map[string]any，返回其深拷贝副本（ok=true）；否则 ok=false。
func cloneSrcAnyMap(src any) (map[string]any, bool) {
	switch m := src.(type) {
	case map[string]any:
		return cloneAnyMap(m), true
	case *map[string]any:
		if m == nil {
			return nil, true
		}
		return cloneAnyMap(*m), true
	default:
		return nil, false
	}
}

// cloneAnyMap 深拷贝 map[string]any，保留每个值的原始 Go 类型（不经过 JSON）。
func cloneAnyMap(src map[string]any) map[string]any {
	if src == nil {
		return nil
	}
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = cloneAnyValue(v)
	}
	return dst
}

// cloneAnyValue 递归深拷贝 any 值：重点处理 map[string]any 与 []any 容器，其余类型直接保留（类型不丢、也无必要深拷贝）。
func cloneAnyValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return cloneAnyMap(val)
	case []any:
		out := make([]any, len(val))
		for i, item := range val {
			out[i] = cloneAnyValue(item)
		}
		return out
	case []map[string]any:
		out := make([]map[string]any, len(val))
		for i, item := range val {
			out[i] = cloneAnyMap(item)
		}
		return out
	default:
		return v
	}
}
