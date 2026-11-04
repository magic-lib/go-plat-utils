package ruleengine_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
	"github.com/shopspring/decimal"
)

// decimal 参数按数值参与计算，不能按系数
func TestDecimalParameterValue(t *testing.T) {
	fromDB, _ := decimal.NewFromString("1.00") // 系数 100，指数 -2
	cases := []struct {
		val  decimal.Decimal
		want float64
	}{
		{fromDB, 1},
		{decimal.NewFromFloat(100), 100}, // 系数 1，指数 2
		{decimal.NewFromFloat(2.5), 2.5},
	}
	for _, c := range cases {
		got, err := ruleengine.NewEngineLogic().EvaluateString("amount * 1", map[string]any{"amount": c.val})
		if err != nil || got != c.want {
			t.Errorf("amount=%s: got %v, %v; want %v", c.val, got, err, c.want)
		}
	}
}
