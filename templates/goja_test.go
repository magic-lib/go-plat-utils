package templates_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/templates"
)

// gojaArgs 用于验证 struct 注入
type gojaArgs struct {
	Name string
	Age  int
}

// gojaArgsWithMethod 简单方法对象
type gojaArgsWithMethod struct {
	Base int
}

func (a *gojaArgsWithMethod) Double() int {
	return a.Base * 2
}

// TestGojaRunStringWithoutArgs 未配置 ArgsName 时，args 不注入，脚本独立运行
func TestGojaRunStringWithoutArgs(t *testing.T) {
	g := templates.NewGoja()

	got, err := g.RunString("1+2", nil)
	if err != nil {
		t.Fatalf("RunString err = %v", err)
	}
	if got != int64(3) {
		t.Errorf("RunString(1+2) = %v (%T), want 3", got, got)
	}
}

// TestGojaRunStringWithArgsName 配置 ArgsName 后，整个对象以该变量名注入
func TestGojaRunStringWithArgsName(t *testing.T) {
	g := templates.NewGoja()
	conf := &templates.GojaConfig{ArgsName: "args"}

	cases := []struct {
		name string
		expr string
		args any
		want any
	}{
		{
			name: "map 取值",
			expr: "args.a + args.b",
			args: map[string]any{"a": 2, "b": 3},
			want: int64(5),
		},
		{
			name: "map 嵌套",
			expr: "args.user.name",
			args: map[string]any{"user": map[string]any{"name": "tom"}},
			want: "tom",
		},
		{
			name: "struct 字段",
			expr: "args.Name",
			args: gojaArgs{Name: "tom", Age: 18},
			want: "tom",
		},
		{
			name: "整对象传参",
			expr: "JSON.stringify(args)",
			args: map[string]any{"k": "v"},
			want: `{"k":"v"}`,
		},
		{
			name: "对象作为函数实参",
			expr: `(function(o){ return o.Age > 10; })(args)`,
			args: gojaArgs{Name: "tom", Age: 18},
			want: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := g.RunString(c.expr, c.args, conf)
			if err != nil {
				t.Fatalf("RunString(%q) err = %v", c.expr, err)
			}
			if got != c.want {
				t.Errorf("RunString(%q) = %v (%T), want %v (%T)", c.expr, got, got, c.want, c.want)
			}
		})
	}
}

// TestGojaRunStringArgsCases 边界：ArgsName 为空、args 为 nil、cfg 为空
func TestGojaRunStringArgsCases(t *testing.T) {
	g := templates.NewGoja()

	// ArgsName 为空：即使传了 args 也不注入
	if got, err := g.RunString("typeof args", map[string]any{"a": 1}, &templates.GojaConfig{}); err != nil {
		t.Errorf("ArgsName 为空 err = %v", err)
	} else if got != "undefined" {
		t.Errorf("ArgsName 为空时应未注入, got %v", got)
	}

	// args 为 nil：不注入
	if got, err := g.RunString("typeof args", nil, &templates.GojaConfig{ArgsName: "args"}); err != nil {
		t.Errorf("args 为 nil err = %v", err)
	} else if got != "undefined" {
		t.Errorf("args 为 nil 时应未注入, got %v", got)
	}

	// 不传 cfg
	if got, err := g.RunString("typeof args", map[string]any{"a": 1}); err != nil {
		t.Errorf("不传 cfg err = %v", err)
	} else if got != "undefined" {
		t.Errorf("不传 cfg 时应未注入, got %v", got)
	}

	// cfg 显式传 nil
	if got, err := g.RunString("1+1", map[string]any{"a": 1}, nil); err != nil {
		t.Errorf("cfg 为 nil err = %v", err)
	} else if got != int64(2) {
		t.Errorf("cfg 为 nil 时 = %v, want 2", got)
	}
}

// TestGojaRunStringRetainsState 同一实例内后一次执行可复用前一次定义的变量，且 ArgsName 可被覆盖
func TestGojaRunStringRetainsState(t *testing.T) {
	g := templates.NewGoja()

	if _, err := g.RunString("var counter = 1", nil); err != nil {
		t.Fatalf("定义变量 err = %v", err)
	}
	if got, err := g.RunString("counter + 1", nil); err != nil {
		t.Errorf("复用变量 err = %v", err)
	} else if got != int64(2) {
		t.Errorf("counter + 1 = %v, want 2", got)
	}

	// 同名变量被新的 args 覆盖
	conf := &templates.GojaConfig{ArgsName: "args"}
	if got, err := g.RunString("args.v", map[string]any{"v": "first"}, conf); err != nil {
		t.Fatalf("第一次注入 err = %v", err)
	} else if got != "first" {
		t.Errorf("第一次注入 = %v, want first", got)
	}
	if got, err := g.RunString("args.v", map[string]any{"v": "second"}, conf); err != nil {
		t.Fatalf("第二次注入 err = %v", err)
	} else if got != "second" {
		t.Errorf("第二次注入 = %v, want second", got)
	}
}

// TestGojaRunStringError 脚本异常时返回 error
func TestGojaRunStringError(t *testing.T) {
	g := templates.NewGoja()

	if _, err := g.RunString("this is not valid js", nil); err == nil {
		t.Error("语法错误应返回 error")
	}
	if _, err := g.RunString("undefinedFuncCall()", nil); err == nil {
		t.Error("调用未定义函数应返回 error")
	}
}

// TestGojaWithFuncsAndVars 函数与变量注入
func TestGojaWithFuncsAndVars(t *testing.T) {
	g := templates.NewGoja()

	err := g.WithFuncsAndVars(
		map[string]any{
			"add": func(a, b int) int { return a + b },
		},
		map[string]any{
			"prefix": "hello",
		},
	)
	if err != nil {
		t.Fatalf("WithFuncsAndVars err = %v", err)
	}

	got, err := g.RunString("prefix + ' ' + add(3, 5)", nil)
	if err != nil {
		t.Fatalf("RunString err = %v", err)
	}
	if got != "hello 8" {
		t.Errorf("got %v, want %q", got, "hello 8")
	}
}
