package ruleengine_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
)

// TestMapFunction 验证 Map 既支持表达式映射，也支持按字段名取值映射
func TestMapFunction(t *testing.T) {
	userList := []any{
		map[string]any{"name": "tom", "age": 12, "addr": map[string]any{"city": "bj"}},
		map[string]any{"name": "jerry", "age": 20, "addr": map[string]any{"city": "sh"}},
		map[string]any{"name": "jack", "age": 30, "addr": map[string]any{"city": "gz"}},
	}
	userArgs := map[string]any{"users": userList}

	testCases := []struct {
		name  string
		rule  string
		args  map[string]any
		want  any
		isErr bool
	}{
		//表达式映射
		{"表达式-标量四则", `Map(Array(1,2,3), 'item + 1')`, nil, []any{2.0, 3.0, 4.0}, false},
		{"表达式-标量乘除", `Map(Array(1,2,3), 'item * 10')`, nil, []any{10.0, 20.0, 30.0}, false},
		{"表达式-字段运算", `Map(users, 'item.age * 2')`, userArgs, []any{24.0, 40.0, 60.0}, false},
		{"表达式-字段直接暴露", `Map(users, 'age + 1')`, userArgs, []any{13.0, 21.0, 31.0}, false},
		{"表达式-中括号写法", `Map(users, '[item.age] + 1')`, userArgs, []any{13.0, 21.0, 31.0}, false},
		{"表达式-拼接字符串", `Map(users, 'item.name + \'!\'' )`, userArgs, []any{"tom!", "jerry!", "jack!"}, false},
		{"表达式-返回布尔", `Map(users, 'age > 18')`, userArgs, []any{false, true, true}, false},
		{"表达式-取元素本身", `Map(users, 'item')`, userArgs, userList, false},

		//按字段名映射
		{"字段名-取一级字段", `Map(users, 'name')`, userArgs, []any{"tom", "jerry", "jack"}, false},
		{"字段名-取数字字段", `Map(users, 'age')`, userArgs, []any{12.0, 20.0, 30.0}, false},
		{"字段名-取嵌套字段", `Map(users, 'addr.city')`, userArgs, []any{"bj", "sh", "gz"}, false},
		{"字段名-取对象字段", `Map(users, 'addr')`, userArgs, []any{
			map[string]any{"city": "bj"},
			map[string]any{"city": "sh"},
			map[string]any{"city": "gz"},
		}, false},
		{"字段名-不存在则填nil", `Map(users, 'notExist')`, userArgs, []any{nil, nil, nil}, false},

		//长度与边界
		{"空数组", `Map(Array(), 'item + 1')`, nil, []any{}, false},
		{"长度保持不变", `Len(Map(users, 'name'))`, userArgs, 3.0, false},
		{"可嵌套组合", `Map(Filter(users, 'age > 18'), 'name')`, userArgs, []any{"jerry", "jack"}, false},
		{"嵌套Filter后取长度", `Len(Filter(Map(users, 'age'), 'item > 18'))`, userArgs, 2.0, false},

		//异常
		{"第二个参数不是字符串", `Map(users, 123)`, userArgs, nil, true},
		{"漏传第二个参数", `Map(Array(1))`, nil, nil, true},
	}

	for _, tc := range testCases {
		got, err := ruleengine.NewEngineLogic().EvaluateString(tc.rule, tc.args)
		if tc.isErr {
			if err == nil {
				t.Errorf("%s: 期望报错，实际返回 %v", tc.name, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: EvaluateString(%q) err = %v", tc.name, tc.rule, err)
			continue
		}
		if !jsonEqual(t, got, tc.want) {
			t.Errorf("%s: EvaluateString(%q) = %v, want %v", tc.name, tc.rule, conv.String(got), conv.String(tc.want))
		}
	}
}
