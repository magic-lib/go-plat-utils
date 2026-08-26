package templates

import (
	"testing"
)

func TestMatchNumberRange(t *testing.T) {
	cases := []struct {
		name string
		val  any
		want any
		ok   bool
	}{
		// 精确值
		{"exact int", 135, "135", true},
		{"exact string", "135", "135", true},
		{"exact float", 135.0, "135", true},

		// [10,20] 闭区间
		{"closed lower", 10, "[10,20]", true},
		{"closed upper", 20, "[10,20]", true},
		{"closed middle", "15", "[10,20]", true},
		{"closed below", 9.99, "default_value", false},
		{"closed above", 20.01, "(20,50]", true},

		// (30, +inf) 左开右无穷
		{"left open above", 31, "(30, +inf)", true},
		{"left open big", 999999, "(30, +inf)", true},
		{"left open boundary", 30, "(20,50]", true},
		{"left open float", 30.5, "(30, +inf)", true},

		// (-inf, 5] 左无穷右闭
		{"right closed negative", -100, "(-inf, 5]", true},
		{"right closed boundary", 5, "(-inf, 5]", true},
		{"right closed above", 5.5, "default_value", false},

		// (20,50] 左开右闭（注意 (30, +inf) 在其前，>30 的值先命中 (30, +inf)）
		{"half open boundary", 20, "[10,20]", true},
		{"half open above", 21, "(20,50]", true},
		{"half open upper", 50, "(30, +inf)", true},
		{"half open beyond", 50.5, "(30, +inf)", true},

		// 未命中任何规则返回 default
		{"default fallback", 100, "(30, +inf)", true},
		{"default fallback float", 7, "default_value", false},

		// 非法输入
		{"bad json", nil, "default_value", false},
		{"non numeric", "abc", "default_value", false},
		{"empty", "", "", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg := ""
			if c.name == "bad json" {
				cfg = `{invalid`
			}
			ruleList := []string{
				"",
				"135",
				"[10,20]",
				"(30, +inf)",
				"(-inf, 5]",
				"(20,50]"}
			got, ok := MatchRangeExpr(ruleList, c.val)
			if ok != c.ok {
				t.Fatalf("MatchRangeExpr(%q, %v) ok = %v, want %v", cfg, c.val, ok, c.ok)
			}
			//t.Logf("MatchRangeExpr(%v) = %v (%v)", c.val, got, ok)

			wantKey := ""
			if got >= 0 {
				wantKey = ruleList[got]
			} else {
				wantKey = "default_value"
			}
			if wantKey != c.want {
				t.Errorf("MatchRangeExpr(%q, %v) = %v (%T), want %v (%T)", cfg, c.val, got, got, c.want, c.want)
			}
		})
	}
}
