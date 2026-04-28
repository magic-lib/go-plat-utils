package templates_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/templates"
	"testing"
)

func TestMaxOrMinFunction(t *testing.T) {
	exprEngine := templates.NewRuleExprEngine()
	result, err := exprEngine.RunString(`Max('{{.dfss}}',2,3.4,Min(14,13),10.6)`, map[string]any{
		"dfss": 12,
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)

	result, err = exprEngine.RunString(`Min('{{.dfss}}',2,3.4,2.1,7.0,10.6)`, map[string]any{
		"dfss": 123,
	})

	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(result)

	return
}
