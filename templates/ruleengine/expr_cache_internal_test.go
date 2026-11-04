package ruleengine

import (
	"strconv"
	"testing"
)

// 不同取值拼出的表达式各占一条缓存，总数不能超过上限
func TestExpressionCacheBounded(t *testing.T) {
	e := NewEngineLogic()
	for i := 0; i < expressCacheSize+500; i++ {
		if _, err := e.EvaluateString(strconv.Itoa(i)+" > 0", nil); err != nil {
			t.Fatalf("EvaluateString: %v", err)
		}
	}
	if n := expressCache.Len(); n > expressCacheSize {
		t.Fatalf("缓存条目 %d 超过上限 %d", n, expressCacheSize)
	}
}
