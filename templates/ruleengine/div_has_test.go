package ruleengine_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
)

func TestDivAndHasEdgeCases(t *testing.T) {
	params := map[string]any{"zero": 0, "nilList": nil, "list": []any{"a", "b"}}
	testCases := []struct {
		rule  string
		want  any
		isErr bool
	}{
		{`Div(6, 3)`, float64(2), false},
		{`Div(6, zero)`, nil, true},
		{`Div(6, 2, 0)`, nil, true},
		{`Div('x')`, float64(0), false},
		{`Has(nilList, 'a')`, false, false},
		{`In('a', nilList)`, false, false},
		{`Has(list, 'a')`, true, false},
	}
	for _, tc := range testCases {
		got, err := ruleengine.NewEngineLogic().EvaluateString(tc.rule, params)
		if tc.isErr {
			if err == nil {
				t.Errorf("%s: 期望报错, 实际返回 %v", tc.rule, got)
			}
			continue
		}
		if err != nil || got != tc.want {
			t.Errorf("%s = %v, %v; want %v", tc.rule, got, err, tc.want)
		}
	}
}
