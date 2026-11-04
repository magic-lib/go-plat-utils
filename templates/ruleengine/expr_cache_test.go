package ruleengine_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/internal/govaluate-3.0.0"
	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
)

// 注册了同名自定义函数的两个引擎，不能共用对方编译好的表达式
func TestExpressionCacheScopeByFunctions(t *testing.T) {
	newEngine := func(ret float64) *ruleengine.EngineLogic {
		e := ruleengine.NewEngineLogic()
		if err := e.SetCustomerFunctions(map[string]govaluate.ExpressionFunction{
			"CacheScopeFoo": func(args ...any) (any, error) { return ret, nil },
		}); err != nil {
			t.Fatalf("SetCustomerFunctions: %v", err)
		}
		return e
	}
	e1, e2 := newEngine(1), newEngine(2)
	for i := 0; i < 3; i++ {
		got1, err1 := e1.EvaluateString("CacheScopeFoo() + 0", nil)
		got2, err2 := e2.EvaluateString("CacheScopeFoo() + 0", nil)
		if err1 != nil || err2 != nil || got1 != float64(1) || got2 != float64(2) {
			t.Fatalf("e1 = %v(%v), e2 = %v(%v); want 1, 2", got1, err1, got2, err2)
		}
	}
	// 未注册自定义函数的引擎仍可正常使用内置函数
	if got, err := ruleengine.NewEngineLogic().EvaluateString("Add(1, 2)", nil); err != nil || got != float64(3) {
		t.Fatalf("Add(1,2) = %v, %v", got, err)
	}
}
