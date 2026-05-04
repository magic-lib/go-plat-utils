package templates_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates"
	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
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
func TestRuleStringFunction(t *testing.T) {
	args := map[string]any{
		"dfss": map[string]any{
			"aaa": "aa",
		},
	}
	newWhen := `{{dfss.aaa}}=='aa'`

	exprEngine := templates.NewRuleExprEngine()
	result, err := exprEngine.RunString(newWhen, args)

	if err != nil {
		fmt.Println(result, err)
	}
	fmt.Println(result)

	newWhen2 := "[dfss.aaa]=='aa'"

	ruleEngine := ruleengine.NewEngineLogic()
	runStringArg := conv.KeyListFromMap(args)
	newWhenString := conv.String(newWhen2)

	fmt.Println(newWhenString)

	retVal, err := ruleEngine.RunString(newWhenString, runStringArg)

	fmt.Println(retVal, err)

	newWhen2 = "dfss\\.aaa=='aa'"

	ruleEngine = ruleengine.NewEngineLogic()
	runStringArg = conv.KeyListFromMap(args)
	newWhenString = conv.String(newWhen2)

	fmt.Println(newWhenString)

	retVal, err = ruleEngine.RunString(newWhenString, runStringArg)

	fmt.Println(retVal, err)

	return
}
