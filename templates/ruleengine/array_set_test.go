package ruleengine

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

func TestArraySetFunctions(t *testing.T) {
	r := NewEngineLogic()
	params := map[string]any{
		"a":    []int{1, 2, 3, 3},
		"b":    []string{"2", "3", "4"},
		"c":    []any{3.0, 2.0, 9.0},
		"tags": `["vip","new"]`,
		"none": nil,
	}
	cases := []struct {
		rule string
		want string
	}{
		// 交集：去重，按第一个数组的顺序
		{"Intersect(Array(1,2,3), Array(2,3,4))", "[2,3]"},
		{"Intersect(Array(1,2,3), Array(3,2), Array(2,3,5))", "[2,3]"},
		{"Intersect(Array('a','b'), Array('c'))", "[]"},
		{"Intersect(Array(1,2,2))", "[1,2]"},
		{"Intersect(a, b, c)", "[2,3]"},                    // 参数里的数组，数字和字符串按字符串比较
		{"Intersect(tags, Array('new','old'))", `["new"]`}, // JSON 数组字符串
		{"Intersect(a, none)", "[]"},                       // nil 视为空数组
		{"Intersect([1,2,3], [3,1])", "[1,3]"},             // [] 字面量写法
		// 并集：去重，按出现顺序
		{"Union(Array(1,2), Array(2,3))", "[1,2,3]"},
		{"Union(Array(1,2), Array(2,3), Array(5,1))", "[1,2,3,5]"},
		{"Union(a, none)", "[1,2,3]"},
		{"Union(tags, Array('vip','old'))", `["vip","new","old"]`},
		{"Union(['a'], ['b','a'])", `["a","b"]`},
		// 与其他数组函数组合
		{"Len(Intersect(a, b))", "2"},
		{"In(2, Union(Array(1), Array(2)))", "true"},
	}
	for _, c := range cases {
		ret, err := r.EvaluateString(c.rule, params)
		if err != nil {
			t.Fatalf("%s: %v", c.rule, err)
		}
		if got := conv.String(ret); got != c.want {
			t.Errorf("%s = %s, want %s", c.rule, got, c.want)
		}
	}
}

func TestArraySetFunctionsInvalidArg(t *testing.T) {
	r := NewEngineLogic()
	for _, rule := range []string{"Intersect(Array(1,2), 5)", "Union('abc', Array(1))"} {
		if _, err := r.EvaluateString(rule, nil); err == nil {
			t.Errorf("%s: expected error for non-array argument", rule)
		}
	}
}
