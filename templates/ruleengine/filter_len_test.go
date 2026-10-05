package ruleengine_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
)

func newUserList() []any {
	return []any{
		map[string]any{"name": "tom", "age": 12, "addr": map[string]any{"city": "bj"}},
		map[string]any{"name": "jerry", "age": 20, "addr": map[string]any{"city": "sh"}},
		map[string]any{"name": "jack", "age": 30, "addr": map[string]any{"city": "gz"}},
	}
}

// TestLenFunction 验证 Len 取长度的各种场景
func TestLenFunction(t *testing.T) {
	var nilSlice []any

	testCases := []struct {
		name  string
		rule  string
		args  map[string]any
		want  any
		isErr bool
	}{
		{"字面量数组", `Len(Array(1,2,3))`, nil, 3, false},
		{"变量数组", `Len(users)`, map[string]any{"users": newUserList()}, 3, false},
		{"单元素", `Len(Array(1))`, nil, 1, false},
		{"空数组", `Len(Array())`, nil, 0, false},
		{"嵌套取长度", `Len(Filter(users, 'age > 18'))`, map[string]any{"users": newUserList()}, 2, false},
		{"nil 切片", `Len(nilSlice)`, map[string]any{"nilSlice": nilSlice}, 0, false},
		{"非数组按单个参数算", `Len(123)`, nil, 1, false},
		{"参数不存在", `Len(notExist)`, nil, nil, true},
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
		gotStr, wantStr := conv.String(got), conv.String(tc.want)
		if gotStr != wantStr {
			t.Errorf("%s: EvaluateString(%q) = %v, want %v", tc.name, tc.rule, gotStr, wantStr)
		}
	}
}

// TestFilterFunction 验证 Filter 既支持表达式过滤，也支持原来的相等匹配
func TestFilterFunction(t *testing.T) {
	userList := newUserList()
	jerryAndJack := []any{userList[1], userList[2]}
	jackOnly := []any{userList[2]}

	testCases := []struct {
		name string
		rule string
		args map[string]any
		want any
	}{
		{"表达式-标量", `Filter(Array(1,2,3), 'item > 1')`, nil, []any{2.0, 3.0}},
		{"表达式-无命中", `Filter(Array(1,2,3), 'item > 9')`, nil, []any{}},

		{"表达式-字段直接暴露", `Filter(users, 'age > 18')`, map[string]any{"users": userList}, jerryAndJack},
		{"表达式-item取字段", `Filter(users, 'item.age > 18')`, map[string]any{"users": userList}, jerryAndJack},
		{"表达式-中括号写法", `Filter(users, '[item.age] > 18')`, map[string]any{"users": userList}, jerryAndJack},
		{"表达式-嵌套字段", `Filter(users, 'item.addr.city == \'gz\'')`, map[string]any{"users": userList}, jackOnly},
		{"表达式-逻辑组合", `Filter(users, 'age >= 20 && name != \'jerry\'')`, map[string]any{"users": userList}, jackOnly},

		{"相等匹配-保留重复项", `Filter(Array(1,2,2,3), 2)`, nil, []any{2.0, 2.0}},
		{"相等匹配-无命中", `Filter(users, 999)`, map[string]any{"users": userList}, []any{}},
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

// jsonEqual 按 JSON 语义比较两个值（map 序列化后的 key 顺序是不确定的）
func jsonEqual(t *testing.T, got, want any) bool {
	t.Helper()
	if reflect.DeepEqual(got, want) {
		return true
	}
	return reflect.DeepEqual(toJSONObj(got), toJSONObj(want))
}

func toJSONObj(src any) any {
	if src == nil {
		return nil
	}
	bytes, err := json.Marshal(src)
	if err != nil {
		return nil
	}
	var obj any
	if err := json.Unmarshal(bytes, &obj); err != nil {
		return nil
	}
	return obj
}
