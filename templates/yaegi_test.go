package templates_test

import (
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
	"testing"
)

func TestTemplates4(t *testing.T) {
	// 1. 创建解释器
	i := interp.New(interp.Options{})
	// 2. 加载标准库
	i.Use(stdlib.Symbols)

	// 3. 执行Go代码
	_, _ = i.Eval(`import "fmt"`)
	_, _ = i.Eval(`fmt.Println("Yaegi v0.17+ 运行中")`)

	// 4. 动态函数调用
	_, _ = i.Eval(`func add(a,b int) int { return a+b }`)
	v, _ := i.Eval(`add(3,5)`)
	println("add(3,5) =", v.Int())
}
