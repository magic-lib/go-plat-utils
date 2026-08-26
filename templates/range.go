package templates

import (
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
	"math"
	"strconv"
	"strings"
)

// MatchRangeExpr 解析 ranges 区间配置 JSON，判断 val 命中哪条规则表达式。
//
// 表达式支持两种形式：
//   - 精确值："135"
//   - 区间："[10,20]"、"(30, +inf)"、"(-inf, 5]"、"(20,50]"；
//     "["/"]" 表示包含边界，"("/")" 表示不包含边界，"+inf"/"-inf"（大小写不敏感）表示无穷。
func MatchRangeExpr(rangeList []string, val any) (int, bool) {
	if val == nil {
		return -1, false
	}
	// 首先直接命中
	currentIndex := -1
	lo.ForEachWhile(rangeList, func(expr string, i int) bool {
		if strings.TrimSpace(expr) == conv.String(val) {
			currentIndex = i
			return false
		}
		return true
	})
	if currentIndex >= 0 {
		return currentIndex, true
	}
	num, err := conv.Convert[float64](val)
	if err != nil {
		return -1, false
	}
	for i, expr := range rangeList {
		if matchNumberExpr(expr, num) {
			return i, true
		}
	}
	return -1, false
}

// matchNumberExpr 判断单个规则表达式是否匹配 num
func matchNumberExpr(expr string, num float64) bool {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return false
	}
	first, last := expr[0], expr[len(expr)-1]
	if first == '[' || first == '(' {
		if last != ']' && last != ')' {
			return false
		}
		inner := strings.TrimSpace(expr[1 : len(expr)-1])
		parts := strings.Split(inner, ",")
		if len(parts) != 2 {
			return false
		}
		lo, loInf, loOk := parseBound(parts[0])
		hi, hiInf, hiOk := parseBound(parts[1])
		if !loOk || !hiOk {
			return false
		}
		if !loInf {
			if first == '[' {
				if num < lo {
					return false
				}
			} else if num <= lo {
				return false
			}
		}
		if !hiInf {
			if last == ']' {
				if num > hi {
					return false
				}
			} else if num >= hi {
				return false
			}
		}
		return true
	}
	// 精确值
	if v, err := strconv.ParseFloat(expr, 64); err == nil {
		return num == v
	}
	return false
}

// parseBound 解析区间边界，返回 (值, 是否为无穷, 是否合法)
func parseBound(s string) (float64, bool, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "+inf", "inf", "infinity", "+infinity":
		return math.Inf(1), true, true
	case "-inf", "-infinity":
		return math.Inf(-1), true, true
	}
	v, err := conv.Convert[float64](s)
	if err != nil {
		return 0, false, false
	}
	return v, false, true
}
