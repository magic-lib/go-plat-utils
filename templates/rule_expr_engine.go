package templates

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
	"log"
	"regexp"
)

// RuleExprEngine 公式计算引擎（变量替换 + 表达式计算 + 函数支持）
type RuleExprEngine struct {
	prefix       string
	suffix       string
	jsonTemplate *JsonMapTemplate
}

// NewRuleExprEngine 创建引擎实例
func NewRuleExprEngine(fixString ...string) *RuleExprEngine {
	prefix := DefaultPrefix
	suffix := DefaultSuffix
	if len(fixString) == 1 {
		prefix = fixString[0]
	} else if len(fixString) == 2 {
		prefix = fixString[0]
		suffix = fixString[1]
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
// noEvaluate 不计算，默认是需要计算的，如果不需要计算，则传true
func (e *RuleExprEngine) RunString(expr string, args any, noEvaluate ...bool) (any, error) {
	var argAny any
	var argMap map[string]any

	if argMapTemp, ok := args.(map[string]any); ok {
		argMap = argMapTemp
		argAny = argMap
	} else {
		argMap = make(map[string]any)
		err := conv.Unmarshal(args, &argMap)
		argAny = argMap
		if err != nil {
			fmt.Println("RuleExprEngine RunString Unmarshal expr:", expr, "args:", conv.String(args), "err:", err)
			if !cond.IsJsonMap(conv.String(args)) {
				argAny = args //如果不是json格式，就用原始格式
			}
		}
	}

	tmpl := NewTemplate(expr, e.prefix, e.suffix)
	newExpr := tmpl.Replace(argAny)

	newWhen, err := e.jsonTemplate.Replace(newExpr, argMap)
	if err != nil {
		return nil, fmt.Errorf("expr: %s, %v", newExpr, err)
	}
	if cond.IsNumeric(newWhen) { //如果全是数字，则直接返回，解决"0123",执行后，会返回123的问题
		return newWhen, nil
	}
	// 字符串只包含数字和算术运算符（如 "2026-08-14"、"1+2"）时直接返回原值，
	// 避免日期、编号等被表达式引擎误算（如 "2026-08-14" 会被算成 2026-8-14=2004）
	if _, ok := newWhen.(string); ok {
		if isReplaceString(expr, e.prefix, e.suffix) {
			return newWhen, nil
		}
		//if isPureNumericExpr(newWhenStr) && newWhenStr == expr {
		//	return newWhen, nil
		//}
	}
	noUseEvaluate := false
	if len(noEvaluate) > 0 {
		noUseEvaluate = noEvaluate[0]
	}
	if noUseEvaluate {
		return newWhen, nil
	}

	ruleEngine := ruleengine.NewEngineLogic()
	newWhenString := conv.String(newWhen)
	retVal, err := ruleEngine.EvaluateString(newWhenString, argMap)
	if err != nil {
		//如果错误是json格式，则直接返回即可，证明不是表达式，可能只是进行变量替换而已
		if cond.IsJson(newWhenString) {
			return newWhen, nil
		}
		//err: No parameter 'tianlin0' found., RunString: tianlin0
		if isParameterNotFoundError(err) {
			log.Println("RunString:", err)
			return newWhen, nil
		}

		// 有可能就没有变量，所以不需要去运行错误，比如一个字符串不加引号，就应该是正确的，如果有表达式，计算的话，就应该报错
		log.Println("RuleExprEngine RunString expr:", expr, "return:", newWhenString, "args:", conv.String(args), " RunStringErr:", err)
		return newWhen, err
	}
	return retVal, nil
}

// RenderObject 传入一个对象，通过 expr 模板表达式渲染后返回新的对象。
// 若 expr 为空则原样返回 args；若结果可解析为 JSON map 则返回 map[string]any。
// 主要用于处理方法参数或返回值中的模板替换场景。
func (e *RuleExprEngine) RenderObject(expr string, args any) (any, error) {
	if expr == "" {
		return args, nil
	}
	newArgs, err := e.RunString(expr, args)
	if err != nil {
		return nil, err
	}
	if cond.IsJsonMap(conv.String(newArgs)) {
		var argMap map[string]any
		_ = conv.Unmarshal(newArgs, &argMap)
		return argMap, nil
	}
	return newArgs, nil
}

// isParameterNotFoundError 判断错误是否是因为参数未找到（模板中使用了不存在的变量）
func isParameterNotFoundError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := err.Error()
	pattern := `No parameter '[^']*' found\.`
	matched, _ := regexp.MatchString(pattern, errMsg)
	return matched
}

// pureNumericExprRe 匹配只包含数字、小数点、算术运算符（+ - * / % ^ ( )）和空白的字符串
var pureNumericExprRe = regexp.MustCompile(`^[0-9.+\-*/%^()\s]+$`)

// isPureNumericExpr 判断字符串是否只由数字和算术符号组成（无字母、无变量占位符残留）。
func isPureNumericExpr(s string) bool {
	if s == "" {
		return false
	}
	return pureNumericExprRe.MatchString(s)
}

// exprOperatorRe 匹配表达式运算符特征字符。
// 占位符外的文本若包含这些字符，说明字符串需要表达式计算，而非纯变量替换。
var exprOperatorRe = regexp.MustCompile(`[+\-*/%^()<>=!&|?;'",\[\]{}@#\\]`)

// isReplaceString 判断字符串是否"仅仅只是需要替换变量"：
// 整个字符串仅由模板占位符（prefix...suffix）和普通文本组成，占位符外不含任何
// 表达式运算符（+ - * / % 等）。若为 true，说明变量替换完成后即可直接返回结果，
// 无需再进入表达式引擎计算（避免替换出的日期、编号、纯文本被引擎误算或误报错，
// 如 {{a}} 替换成 "2026-08-14" 后不再被算成 2004）。
func isReplaceString(s, prefix, suffix string) bool {
	if s == "" || prefix == "" || suffix == "" {
		return false
	}
	pat := regexp.MustCompile(regexp.QuoteMeta(prefix) + `.*?` + regexp.QuoteMeta(suffix))
	rest := pat.ReplaceAllString(s, "")
	if rest == "" {
		return true // 整个字符串就是占位符，如 {{a}}
	}
	// 占位符外的文本必须不含表达式运算符
	return !exprOperatorRe.MatchString(rest)
}
