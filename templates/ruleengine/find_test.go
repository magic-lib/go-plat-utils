package ruleengine_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
)

// TestFindFunction 验证 Find 既支持原来的相等匹配，也支持传入表达式对每个元素求值
func TestFindFunction(t *testing.T) {
	userList := []any{
		map[string]any{"name": "tom", "age": 12, "addr": map[string]any{"city": "bj"}},
		map[string]any{"name": "jerry", "age": 20, "addr": map[string]any{"city": "sh"}},
		map[string]any{"name": "jack", "age": 30, "addr": map[string]any{"city": "gz"}},
	}
	jerry := map[string]any{"name": "jerry", "age": 20, "addr": map[string]any{"city": "sh"}}
	jack := map[string]any{"name": "jack", "age": 30, "addr": map[string]any{"city": "gz"}}

	testCases := []struct {
		name string
		rule string
		args map[string]any
		want any
	}{
		{"相等匹配-数字", `Find(Array(1,2,3), 2)`, nil, 2.0},
		{"相等匹配-字符串", `Find(Array('a','b'), 'b')`, nil, "b"},
		{"相等匹配-找不到", `Find(Array(1,2,3), 9)`, nil, nil},
		{"相等匹配-找不到", `Is('nil', nilJack)`, map[string]any{"nilJack": nil}, true},

		{"表达式-item取字段", `Find(users, 'item.age > 18')`, map[string]any{"users": userList}, jerry},
		{"表达式-字段直接暴露", `Find(users, 'age > 25')`, map[string]any{"users": userList}, jack},
		{"表达式-中括号写法", `Find(users, '[item.age] > 25')`, map[string]any{"users": userList}, jack},
		{"表达式-嵌套字段", `Find(users, 'item.addr.city == \'gz\'')`, map[string]any{"users": userList}, jack},
		{"表达式-逻辑组合", `Find(users, 'age >= 20 && name != \'jerry\'')`, map[string]any{"users": userList}, jack},
		{"表达式-标量数组", `Find(Array(1,2,30), 'item > 5')`, nil, 30.0},
		{"表达式-找不到", `Find(users, 'item.age > 100')`, map[string]any{"users": userList}, nil},
		{"表达式-引用不到变量则退回相等匹配", `Find(Array('tom','jack'), 'jack')`, nil, "jack"},
	}

	for _, tc := range testCases {
		got, err := ruleengine.NewEngineLogic().EvaluateString(tc.rule, tc.args)
		if err != nil {
			t.Errorf("%s: EvaluateString(%q) err = %v", tc.name, tc.rule, err)
			continue
		}
		if !jsonEqual(t, got, tc.want) {
			t.Errorf("%s: EvaluateString(%q) = %v, want %v", tc.name, tc.rule, conv.String(got), conv.String(tc.want))
		}
	}
}
