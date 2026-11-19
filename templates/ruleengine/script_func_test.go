package ruleengine

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestJsFunction(t *testing.T) {
	r := NewEngineLogic()
	params := map[string]any{
		"nid": "151024/92/1", "a": 5, "name": "张三", "list": []any{1, 2, 3},
		"user":  map[string]any{"nid": "151024/92/1", "age": 20},
		"level": map[string]string{"grade": "VIP"},
		"other": map[string]any{"age": 30},
	}
	cases := []struct {
		rule string
		want any
	}{
		{`JavaScript('1 == \"1\"')`, true},   // JS 宽松相等
		{`JavaScript('1 === \"1\"')`, false}, // 严格相等
		{`JavaScript('\"a\" + 5')`, "a5"},    // 字符串拼接，脚本里的引号要转义
		{`JavaScript('\'a\' + 5')`, "a5"},    // 字符串拼接
		{`JavaScript('arg0.substr(-6, 1) == 4', nid) && a > 2`, true},
		{`JavaScript('arg0.slice(-6, -5) * 2', nid) == 8`, true}, // 整数结果转成 float64，可与数字字面量比较
		{`JavaScript('arg0 + arg1', a, 3)`, float64(8)},
		{`JavaScript('args.length', a, 3, nid)`, float64(3)},
		{`JavaScript('arg0.length', name)`, float64(2)},
		{`JavaScript('arg0[1]', list)`, float64(2)},
		{`JavaScript('0.1 + 0.2 > 0.3')`, true}, // JS 宽松语义，0.1+0.2=0.30000000000000004
		{`JavaScript('1+2==3')`, true},
		// map 参数的 key 直接作为变量
		{`JavaScript('nid.substr(-6, 1) == 4 && age >= 18', user)`, true},
		{`JavaScript('grade == \"VIP\"', level)`, true},
		{`JavaScript('arg0.age', user)`, float64(20)},   // arg0 写法仍然可用
		{`JavaScript('age', user, other)`, float64(30)}, // 多个 map 有相同 key 时后面的覆盖前面的
		{`JavaScript('typeof nid', a)`, "undefined"},    // 非 map 参数不会展开
		{`JavaScript('undefined')`, nil},
		{`JavaScript('null')`, nil},
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

func TestJsFunctionErrors(t *testing.T) {
	r := NewEngineLogic()
	for _, rule := range []string{`JavaScript()`, `JavaScript(1)`, `JavaScript('a.b.c')`, `JavaScript('throw new Error(\"x\")')`} {
		if _, err := r.EvaluateString(rule, nil); err == nil {
			t.Errorf("%s: expected error", rule)
		}
	}
}

// 死循环脚本会在超时后被中断
func TestJsFunctionTimeout(t *testing.T) {
	start := time.Now()
	_, err := NewEngineLogic().EvaluateString(`JavaScript('while(true){}')`, nil)
	if err == nil || !strings.Contains(err.Error(), "执行超过") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if cost := time.Since(start); cost > 3*jsRunTimeout {
		t.Fatalf("timeout took too long: %s", cost)
	}
}

// 每次调用独立的 goja.Runtime，可并发执行
func TestJsFunctionConcurrent(t *testing.T) {
	r := NewEngineLogic()
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			ret, err := r.EvaluateString(`JavaScript('arg0 * 2', n)`, map[string]any{"n": i})
			if err != nil || ret != float64(i*2) {
				t.Errorf("n=%d: got %v %v", i, ret, err)
			}
		}(i)
	}
	wg.Wait()
}
