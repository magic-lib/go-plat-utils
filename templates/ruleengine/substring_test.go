package ruleengine

import (
	"fmt"
	"testing"
)

func TestSubString(t *testing.T) {
	r := NewEngineLogic()
	params := map[string]any{"name": "张三丰", "code": 123456, "empty": "", "none": nil}
	cases := []struct {
		rule string
		want string
	}{
		{"SubString('abcdef', 1, 4)", "bcde"},
		{"SubString('abc', 1, 4)", "bc"}, // 超出部分截断
		{"SubString('abc', 0)", "a"},     // 省略长度，取 1 个字符
		{"SubString('abc', 2)", "c"},
		{"SubString('abc', -1)", "c"}, // 负数从后往前数
		{"SubString('abc', -2, 2)", "bc"},
		{"SubString('abc', -3, 1)", "a"},
		{"SubString('abc', -5, 3)", "a"},   // 负数起点超过开头，从开头算起按实际范围截断
		{"SubString('abc', -3, 5)", "abc"}, // 起点有效时，长度超出末尾才截断
		{"SubString('abc', 3)", ""},        // 取不到返回空字符串
		{"SubString('abc', -4)", ""},
		{"SubString('abc', 1, 0)", ""},
		{"SubString('abcdef', 3, -2)", "cd"}, // 负数个数：从起点（含）往左数，按原顺序返回
		{"SubString('abcdef', 3, -1)", "d"},
		{"SubString('abcdef', 2, -5)", "abc"},  // 往左超出开头的部分截断
		{"SubString('abcdef', -1, -3)", "def"}, // 负数起点与负数个数组合
		{"SubString('abc', 5, -3)", ""},        // 往左数仍然完全在末尾之后
		{"SubString('abc', 5, -4)", "c"},       // 往左数回到字符串范围内，按实际范围截断
		{"SubString(name, 2, -2)", "三丰"},
		{"SubString(name, 1, 2)", "三丰"}, // 按字符计数，中文不会截断成乱码
		{"SubString(name, -1)", "丰"},
		{"SubString(code, 0, 3)", "123"}, // 数字按字符串处理
		{"SubString(empty, 0)", ""},
		{"SubString(none, 0)", ""},
	}
	for _, c := range cases {
		ret, err := r.EvaluateString(c.rule, params)
		if err != nil {
			t.Fatalf("%s: %v", c.rule, err)
		}
		if ret != c.want {
			t.Errorf("%s = %#v, want %#v", c.rule, ret, c.want)
		}
	}
}

func TestSubStringInvalidArgs(t *testing.T) {
	//r := templates.NewRuleExprEngine()
	//for _, rule := range []string{"SubString('abc')", "SubString('abc', 'x')", "SubString('abc', 1, 'y')", "SubString('abc', 1, 2, 3)"} {
	//	if _, err := r.RunString(rule, nil); err == nil {
	//		t.Errorf("%s: expected error", rule)
	//	}
	//}
	//=='1' || SubString(nid, -6, 1)=='2' || SubString(nid, -6, 1)=='3'
	r := NewEngineLogic()
	for _, rule := range []string{"SubString(nid, -6)=='1' || SubString(nid, -6, 1)=='2' || SubString(nid, -6, 1)=='3'"} {
		aa, err := r.RunStringAny(rule, map[string]any{
			"nid": "151024/92/1",
		})
		if err != nil {
			t.Errorf("%s: expected error", rule)
		}
		fmt.Println("aa:", aa)
	}
}
func TestSubStringInvalidArgs1(t *testing.T) {
	r := NewEngineLogic()
	for _, rule := range []string{"SubString(nid, -6)=='4'", "SubString(nid, -6)==4"} {
		aa, err := r.RunStringAny(rule, map[string]any{
			"nid": "151024/92/1",
		})
		if err != nil {
			t.Errorf("%s: expected error", rule)
		}
		fmt.Println("aa:", aa)
	}
}
