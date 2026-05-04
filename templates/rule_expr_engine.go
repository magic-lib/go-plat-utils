package templates

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
)

// RuleExprEngine 公式计算引擎（变量替换 + 表达式计算 + 函数支持）
type RuleExprEngine struct {
	prefix       string
	suffix       string
	jsonTemplate *JsonMapTemplate
}

// NewRuleExprEngine 创建引擎实例
func NewRuleExprEngine(fixString ...string) *RuleExprEngine {
	prefix := prefixDefault
	suffix := suffixDefault
	if len(fixString) == 1 {
		prefix = fixString[0]
	} else if len(fixString) == 2 {
		prefix = fixString[0]
		suffix = fixString[1]
	}
	if prefix == "" {
		prefix = prefixDefault
	}
	if suffix == "" {
		suffix = suffixDefault
	}
	jsonTemplate := NewJsonMapTemplate(prefix, suffix)
	return &RuleExprEngine{
		prefix:       prefix,
		suffix:       suffix,
		jsonTemplate: jsonTemplate,
	}
}

// RunString
// 执行公式字符串：先替换变量，再计算表达式
// 支持：${var}、四则运算、比较、逻辑、内置函数
func (e *RuleExprEngine) RunString(expr string, args any) (any, error) {
	argMap := make(map[string]any)
	err := conv.Unmarshal(args, &argMap)
	if err != nil {
		fmt.Println("RuleExprEngine RunString Unmarshal expr:", expr, "args:", conv.String(args), "err:", err)
	}

	tmpl := NewTemplate(expr, e.prefix, e.suffix)
	newExpr := tmpl.Replace(argMap)

	newWhen, err := e.jsonTemplate.Replace(newExpr, argMap)
	if err != nil {
		return nil, fmt.Errorf("expr: %s, %v", newExpr, err)
	}
	ruleEngine := ruleengine.NewEngineLogic()
	runStringArg := conv.KeyListFromMap(args)
	newWhenString := conv.String(newWhen)
	retVal, err := ruleEngine.RunString(newWhenString, runStringArg)
	if err != nil {
		// 有可能就没有变量，所以不需要去运行错误，比如一个字符串不加引号，就应该是正确的，如果有表达式，计算的话，就应该报错
		fmt.Println("RuleExprEngine RunString expr:", expr, "return:", newWhenString, "args:", conv.String(args), "err:", err)
		return newWhenString, err
	}
	return retVal, nil
}
