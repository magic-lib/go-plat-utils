package ruleengine

import "testing"

// 引号里的数组字面量是普通字符串，不能被改写成 Array(...)
func TestQuotedArrayLiteralNotRewritten(t *testing.T) {
	r := NewEngineLogic()
	cases := []string{
		`name == '[1,2]'`,
		`name == "[1,2]"`,
	}
	for _, rule := range cases {
		ret, err := r.EvaluateString(rule, map[string]any{"name": "[1,2]"})
		if err != nil {
			t.Fatalf("%s: %v", rule, err)
		}
		if ret != true {
			t.Fatalf("%s = %v, want true", rule, ret)
		}
	}
	empty, err := r.EvaluateString(`name == '[]'`, map[string]any{"name": "[]"})
	if err != nil || empty != true {
		t.Fatalf("empty: %v %v", empty, err)
	}
	// 引号外的数组仍然转换
	in, err := r.EvaluateString(`In(2, [1,2,3])`, nil)
	if err != nil || in != true {
		t.Fatalf("In: %v %v", in, err)
	}
}
