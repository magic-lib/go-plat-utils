package templates

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

// 表达式本身是参数里的变量名时取变量值；不是变量名的普通文本仍原样返回
func TestRunStringBareVariable(t *testing.T) {
	e := NewRuleExprEngine()
	ret, err := e.RunString("a", map[string]any{"a": 5})
	if err != nil {
		t.Fatal(err)
	}
	if conv.String(ret) != "5" {
		t.Fatalf("RunString(a) = %v, want 5", ret)
	}
	ret, err = e.RunString("hello", map[string]any{"a": 5})
	if err != nil || ret != "hello" {
		t.Fatalf("RunString(hello) = %v, %v", ret, err)
	}
}
