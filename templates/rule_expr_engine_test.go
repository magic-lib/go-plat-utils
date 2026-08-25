package templates_test

import (
	"fmt"
	"testing"

	"github.com/magic-lib/go-plat-utils/templates"
)

// TestRunStringPureNumericExpr 验证：字符串只包含数字和 +-*/ 等符号时直接返回原值，不做表达式计算
func TestRunStringPureNumericExpr(t *testing.T) {
	e := templates.NewRuleExprEngine()
	cases := []struct {
		name string
		expr string
		want any
	}{
		{"日期字符串", "2026-08-14", "2026-08-14"},
		{"纯加减", "1+2", "1+2"},
		{"带空格运算", "3 * 4", "3 * 4"},
		{"小数乘除", "1.5*2/0.5", "1.5*2/0.5"},
		{"括号", "(1+2)*3", "(1+2)*3"},
		{"取余", "10%3", "10%3"},
	}
	for _, c := range cases {
		got, err := e.RunString(c.expr, map[string]any{})
		if err != nil {
			t.Errorf("%s: RunString(%q) err = %v", c.name, c.expr, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: RunString(%q) = %v (%T), want %v (%T)", c.name, c.expr, got, got, c.want, c.want)
		}
	}
}

// TestRunStringNonExpr 验证：包含字母、变量等非纯数字运算串仍走原有逻辑
func TestRunStringNonExpr(t *testing.T) {
	e := templates.NewRuleExprEngine()

	// 纯字母串：无变量错误后原样返回
	got, err := e.RunString("abc", map[string]any{})
	if err != nil {
		t.Errorf("RunString(abc) err = %v", err)
	}
	if got != "abc" {
		t.Errorf("RunString(abc) = %v, want abc", got)
	}

	// 变量替换：纯数字结果直接返回
	got, err = e.RunString("{{a}}", map[string]any{"a": "5"})
	if err != nil {
		t.Errorf("RunString({{a}}) err = %v", err)
	}
	if got != "5" {
		t.Errorf("RunString({{a}}) = %v, want 5", got)
	}
}

// TestRunStringPureReplace 验证 isReplaceString：仅需替换变量的字符串直接返回替换结果，
// 不再进入表达式引擎（避免日期/编号/纯文本被误算或误报错）
func TestRunStringPureReplace(t *testing.T) {
	e := templates.NewRuleExprEngine()
	cases := []struct {
		name string
		expr string
		args map[string]any
		want any
	}{
		// 单个占位符：替换出日期，不应被引擎算成 2026-8-14=2004
		{"占位符替换日期", "{{a}}", map[string]any{"a": "2026-08-14"}, "2026-08-14"},
		// 单个占位符：替换出纯文本，不应报错
		{"占位符替换文本", "{{a}}", map[string]any{"a": "hello world"}, "hello world"},
		// 占位符+普通文本
		{"占位符加文本", "姓名:{{name}}", map[string]any{"name": "张三"}, "姓名:张三"},
		// 多个占位符拼接
		{"多个占位符", "{{a}}{{b}}", map[string]any{"a": "2026-", "b": "08-14"}, "2026-08-14"},
		// 占位符替换出编号，不应被误算
		{"占位符替换编号", "{{a}}", map[string]any{"a": "0012"}, "0012"},
	}
	for _, c := range cases {
		got, err := e.RunString(c.expr, c.args)
		if err != nil {
			t.Errorf("%s: RunString(%q) err = %v", c.name, c.expr, err)
			continue
		}
		if got != c.want {
			t.Errorf("%s: RunString(%q) = %v (%T), want %v (%T)", c.name, c.expr, got, got, c.want, c.want)
		}
	}
}

// TestRunStringRealExpr 验证：真正的表达式（占位符外含运算符）仍会进入引擎计算或按原逻辑返回
func TestRunStringRealExpr(t *testing.T) {
	e := templates.NewRuleExprEngine()

	// 占位符+比较运算符：isReplaceString 返回 false，走引擎计算
	got, err := e.RunString("{{a}}>5", map[string]any{"a": 3})
	if err != nil {
		t.Errorf("RunString({{a}}>5) err = %v", err)
	}
	if got != false {
		t.Errorf("RunString({{a}}>5) = %v, want false", got)
	}

	// 占位符+加法（替换后为纯数字表达式，由 isPureNumericExpr 兜底直接返回原值）
	got, err = e.RunString("{{a}}+1", map[string]any{"a": 2})
	if err != nil {
		t.Errorf("RunString({{a}}+1) err = %v", err)
	}
	if got != "2+1" {
		t.Errorf("RunString({{a}}+1) = %v, want 2+1", got)
	}
}
func TestJsonMapTemplate(t *testing.T) {
	aa := templates.NewJsonMapTemplate("*", "")
	mm, err := aa.ReplacePath("/api/project/:project/env/*env/redis-config", map[string]any{"project": "123456", "env": "test"})
	fmt.Print(mm)
	fmt.Print(err)
}
