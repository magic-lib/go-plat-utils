package templates

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

// subFormulaTemplate 降额公式模板：{{operator}}、{{rate}} 先做变量替换，再由规则引擎计算
const subFormulaTemplate = "last_credit{{operator}}Max(FloorToDigit(10,product_amount*{{rate}}),10)"

// 结果 = 上次授信 - Max(按 10 向下取整(借款金额 * rate), 10)
func TestSubFormulaTemplate(t *testing.T) {
	cases := []struct {
		productAmount float64 // 借款金额
		lastCredit    float64 // 上次授信
		rate          float64
		want          string
	}{
		{500, 500, 0.05, "480"},  // 500 + MIN(-20, -10) = 480
		{100, 100, 0.05, "90"},   // 100 + MIN(0, -10) = 90
		{300, 400, 0.1, "370"},   // 400 + MIN(-30, -10) = 370
		{1000, 1000, 0.2, "800"}, // 1000 + MIN(-200, -10) = 800
		{360, 360, 0.5, "180"},   // 360 + MIN(-180, -10) = 180
		{150, 200, 0.05, "190"},  // 200 + MIN(0, -10) = 190
	}

	e := NewRuleExprEngine()
	for _, c := range cases {
		args := map[string]any{
			"operator":       "-",
			"rate":           c.rate,
			"product_amount": c.productAmount,
			"last_credit":    c.lastCredit,
		}
		ret, err := e.RunString(subFormulaTemplate, args)
		if err != nil {
			t.Fatalf("借款金额=%v 上次授信=%v rate=%v: %v", c.productAmount, c.lastCredit, c.rate, err)
		}
		if got := conv.String(ret); got != c.want {
			t.Errorf("借款金额=%v 上次授信=%v rate=%v: got %s, want %s", c.productAmount, c.lastCredit, c.rate, got, c.want)
		}
	}
}
