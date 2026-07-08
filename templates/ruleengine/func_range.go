package ruleengine

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// rangeInfo 解析后的区间结构
type rangeInfo struct {
	Min       *float64 // nil = -∞
	Max       *float64 // nil = +∞
	MinClosed bool     // 左边界是否包含 = [
	MaxClosed bool     // 右边界是否包含 = ]
}

// 正则匹配区间格式（兼容整数与小数）：
// 分组1：左括号 ( / [
// 分组2：左值（整数/小数/-∞）
// 分组3：右值（整数/小数/+∞）
// 分组4：右括号 ) / ]
// 数字部分允许可选符号、整数部分、可选的小数部分
var rangeRegex = regexp.MustCompile(`^([\(\[])(-?\d+(?:\.\d+)?|-∞),(\+?\d+(?:\.\d+)?|\+∞)([\)\]])$`)

func parseRange(s string) (*rangeInfo, error) {
	s = strings.TrimSpace(s)
	match := rangeRegex.FindStringSubmatch(strings.Join(strings.Fields(s), ""))
	if len(match) != 5 {
		return nil, fmt.Errorf("%s", "invalid range format, example: [4,7] / (-∞,3.5] / [91.2,+∞)")
	}

	leftBracket, leftValStr, rightValStr, rightBracket := match[1], match[2], match[3], match[4]
	info := &rangeInfo{
		MinClosed: leftBracket == "[",
		MaxClosed: rightBracket == "]",
	}

	// 解析左边界 Min
	if leftValStr == "-∞" {
		info.Min = nil
	} else {
		val, err := strconv.ParseFloat(leftValStr, 64)
		if err != nil {
			return nil, fmt.Errorf("parse min value fail: %w", err)
		}
		info.Min = &val
	}

	// 解析右边界 Max
	if rightValStr == "+∞" {
		info.Max = nil
	} else {
		val, err := strconv.ParseFloat(rightValStr, 64)
		if err != nil {
			return nil, fmt.Errorf("parse max value fail: %w", err)
		}
		info.Max = &val
	}

	return info, nil
}

// match 判断目标值是否匹配当前区间，兼容整数与小数
func (r *rangeInfo) match(diffDay float64) bool {
	// 校验左边界
	if r.Min != nil {
		minVal := *r.Min
		if r.MinClosed {
			if diffDay < minVal {
				return false
			}
		} else {
			if diffDay <= minVal {
				return false
			}
		}
	}

	// 校验右边界
	if r.Max != nil {
		maxVal := *r.Max
		if r.MaxClosed {
			if diffDay > maxVal {
				return false
			}
		} else {
			if diffDay >= maxVal {
				return false
			}
		}
	}
	return true
}
