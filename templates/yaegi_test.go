package templates_test

import (
	"fmt"
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
	_, _ = i.Eval(`import "fmt"\nfmt.Println("Yaegi v0.17+ 运行中")`)
	_, _ = i.Eval(`func add(a,b int) int { return a+b }`)
	v, err := i.Eval(`add(3,5)`)
	fmt.Println("add(3,5) =", v, err)
}
