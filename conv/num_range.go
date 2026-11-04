package conv

import (
	"errors"
	"math"
	"strconv"
)

// 数字转换的越界规则：超过目标类型的最大值时取最大值，低于最小值时取最小值（无符号类型的最小值为 0），
// 视为转换成功；NaN 无法比较大小，转换为整数时返回失败。

// clampInt64 将 v 限制在 [minV, maxV] 范围内
func clampInt64(v, minV, maxV int64) int64 {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

// uint64ToInt64 超过 int64 最大值时取最大值，避免直接转换回绕成负数
func uint64ToInt64(u uint64) int64 {
	if u > math.MaxInt64 {
		return math.MaxInt64
	}
	return int64(u)
}

// floatToInt64 浮点数转 int64，越界取边界值（越界时直接转换的结果依赖平台），NaN 返回失败
func floatToInt64(f float64) (int64, bool) {
	if math.IsNaN(f) {
		return 0, false
	}
	if f >= math.MaxInt64 { // float64(math.MaxInt64) 等于 2^63，已超出 int64 范围
		return math.MaxInt64, true
	}
	if f <= math.MinInt64 {
		return math.MinInt64, true
	}
	return int64(f), true
}

// floatToUint64 浮点数转 uint64，负数取 0，越界取最大值，NaN 返回失败
func floatToUint64(f float64) (uint64, bool) {
	if math.IsNaN(f) {
		return 0, false
	}
	if f <= 0 {
		return 0, true
	}
	if f >= math.MaxUint64 { // float64(math.MaxUint64) 等于 2^64，已超出 uint64 范围
		return math.MaxUint64, true
	}
	return uint64(f), true
}

// parseInt64 解析十进制整数字符串，数值越界时取边界值（strconv 在 ErrRange 时已返回边界值）
func parseInt64(s string) (int64, error) {
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return 0, err
	}
	return v, nil
}

// parseUint64 解析十进制无符号整数字符串，越界取最大值，负数取 0
func parseUint64(s string) (uint64, error) {
	v, err := strconv.ParseUint(s, 10, 64)
	if err == nil || errors.Is(err, strconv.ErrRange) {
		return v, nil
	}
	// ParseUint 不接受符号，带符号的整数按其实际值处理：负数取 0
	if iv, intErr := parseInt64(s); intErr == nil {
		if iv < 0 {
			return 0, nil
		}
		return uint64(iv), nil
	}
	return 0, err
}
