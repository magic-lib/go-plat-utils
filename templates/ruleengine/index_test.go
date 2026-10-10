package ruleengine_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
)

// TestIndexFunction 验证 Index 按下标取元素的各种场景
func TestIndexFunction(t *testing.T) {
	userList := newUserList()

	testCases := []struct {
		name  string
		rule  string
		args  map[string]any
		want  any
		isErr bool
	}{
		{"字面量数组-取首个", `Index(Array(1,2,3), 0)`, nil, 1, false},
		{"字面量数组-取中间", `Index(Array(1,2,3), 1)`, nil, 2, false},
		{"字面量数组-取末个", `Index(Array(1,2,3), 2)`, nil, 3, false},
		{"字面量数组-下标越界", `Index(Array(1,2,3), 3)`, nil, nil, false},
		{"字面量数组-负数下标", `Index(Array(1,2,3), -1)`, nil, nil, false},
		{"空数组", `Index(Array(), 0)`, nil, nil, false},
		{"中括号字面量", `Index([10,20,30], 2)`, nil, 30, false},

		{"变量数组-取元素", `Index(users, 1)`, map[string]any{"users": userList}, userList[1], false},
		{"变量数组-越界", `Index(users, 9)`, map[string]any{"users": userList}, nil, false},
		{"nil 数组", `Index(nilSlice, 0)`, map[string]any{"nilSlice": []any(nil)}, nil, false},
		{"带类型切片", `Index(strSlice, 1)`, map[string]any{"strSlice": []string{"a", "b"}}, "b", false},
		{"字符串形式的数组", `Index(strList, 1)`, map[string]any{"strList": `["a","b"]`}, "b", false},

		{"嵌套使用", `Index(Filter(users, 'age > 18'), 0)`, map[string]any{"users": userList}, userList[1], false},
		{"下标来自表达式", `Index(users, Len(users) - 1)`, map[string]any{"users": userList}, userList[2], false},
		{"取元素后再取字段", `Map(Array(Index(users, 2)), 'name')`, map[string]any{"users": userList}, []any{"jack"}, false},

		{"At 别名", `At(Array(1,2,3), 2)`, nil, 3, false},

		{"下标非整数", `Index(Array(1,2,3), 'x')`, nil, nil, true},
		// Index(Array(1,2,3)) 少写下标时，摊平后无法和"最后一个元素当下标"区分，会当成越界返回 nil
		{"缺少下标", `Index(Array(1,2,3))`, nil, nil, false},
	}

	for _, tc := range testCases {
		got, err := ruleengine.NewEngineLogic().EvaluateString(tc.rule, tc.args)
		if tc.isErr {
			if err == nil {
				t.Errorf("%s: 期望报错，实际返回 %v", tc.name, conv.String(got))
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: EvaluateString(%q) err = %v", tc.name, tc.rule, err)
			continue
		}
		if tc.want == nil {
			if got != nil {
				t.Errorf("%s: EvaluateString(%q) = %v, want nil", tc.name, tc.rule, conv.String(got))
			}
			continue
		}
		if !jsonEqual(t, got, tc.want) {
			t.Errorf("%s: EvaluateString(%q) = %v, want %v", tc.name, tc.rule, conv.String(got), conv.String(tc.want))
		}
	}
}
