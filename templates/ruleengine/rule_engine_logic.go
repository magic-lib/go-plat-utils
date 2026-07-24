package ruleengine

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/internal/govaluate-3.0.0"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/shopspring/decimal"
	"strings"
	"sync"
)

// EngineLogic 规则引擎判断逻辑
type EngineLogic struct {
	functions     map[string]govaluate.ExpressionFunction
	retRulePrefix string
	preString     string //变量前的字符
	afterString   string //变量后的字符
}

var (
	defaultRulePrefix = "RET_RULE_"

	expressCache sync.Map // 缓存表达式,提高执行效率
)

// RuleInfo 定义好的规则，包括简单规则和负责规则
// 因子码，操作符，值，返回值
type RuleInfo struct {
	Key        string //该规则唯一key
	RuleString string //规则字符串
}

// NewEngineLogic 初始化
func NewEngineLogic() *EngineLogic {
	// 定自定义函数映射
	ruleLogic := &EngineLogic{
		retRulePrefix: defaultRulePrefix,
	}
	ruleLogicFunc := new(customerFunc)
	// 内置方法
	ruleLogic.functions = map[string]govaluate.ExpressionFunction{
		// 数字相关
		"Add":          ruleLogicFunc.Add,
		"Sub":          ruleLogicFunc.Sub,
		"Mul":          ruleLogicFunc.Mul,
		"Div":          ruleLogicFunc.Div,
		"Max":          ruleLogicFunc.Max,
		"Min":          ruleLogicFunc.Min,
		"CeilToDigit":  ruleLogicFunc.CeilToDigit,
		"FloorToDigit": ruleLogicFunc.FloorToDigit,
		// 数组区间相关
		"Has":     ruleLogicFunc.Has,
		"In":      ruleLogicFunc.In,
		"Between": ruleLogicFunc.Between, // 新增区间判断
		// 类型相关
		"Is": ruleLogicFunc.Is,
		"As": ruleLogicFunc.As,
		// 逻辑判断
		"If":         ruleLogicFunc.If,
		"Switch":     ruleLogicFunc.Switch,
		"SwitchExpr": ruleLogicFunc.SwitchExpr,
		// 字符串相关
		"Replace":  ruleLogicFunc.Replace,
		"Split":    ruleLogicFunc.Split,
		"Contains": ruleLogicFunc.Contains,
		"JsonGet":  ruleLogicFunc.JsonGet,
		"Join":     ruleLogicFunc.Join,
	}
	return ruleLogic
}

// WithRetRulePrefix 设置返回值的key前缀
func (r *EngineLogic) WithRetRulePrefix(prefix string) *EngineLogic {
	if prefix == "" {
		prefix = defaultRulePrefix
	}
	r.retRulePrefix = prefix
	return r
}

// WithDelimiter 设置变量包裹字符串
func (r *EngineLogic) WithDelimiter(pre, after string) *EngineLogic {
	if pre == "" || after == "" {
		return r
	}
	r.preString = pre
	r.afterString = after
	return r
}

// SetCustomerFunctions 设置自定义函数
func (r *EngineLogic) SetCustomerFunctions(functions map[string]govaluate.ExpressionFunction) error {
	if functions == nil {
		return nil
	}
	repeatKeyList := make([]string, 0)
	for key, val := range functions {
		if _, ok := r.functions[key]; ok {
			repeatKeyList = append(repeatKeyList, key)
			continue
		}
		r.functions[key] = val
	}
	if len(repeatKeyList) > 0 {
		return fmt.Errorf("自定义函数有重复的key: %v", repeatKeyList)
	}
	return nil
}

func (r *EngineLogic) getExpressionByRuleString(ruleString string) (*govaluate.EvaluableExpression, error) {
	var expression *govaluate.EvaluableExpression
	var err error

	if expressionTemp, ok := expressCache.Load(ruleString); ok {
		if expression, ok = expressionTemp.(*govaluate.EvaluableExpression); ok {
			return expression, nil
		}
	}

	if r.functions != nil && len(r.functions) > 0 {
		expression, err = govaluate.NewEvaluableExpressionWithFunctions(ruleString, r.functions)
	} else {
		expression, err = govaluate.NewEvaluableExpression(ruleString)
	}

	if err != nil {
		errHasOccurred := "RunString: " + ruleString
		if containsVariableDot(ruleString) {
			errHasOccurred = fmt.Sprintf("如果变量含有'.'符号，则需要用[]将条件变量包括起来，或者用\\\\进行转义.符号，不要使用{{}}括起来，这样会被替换掉该变量: RunString: %s",
				ruleString)
		} else if err.Error() == "Unbalanced parenthesis" {
			errHasOccurred = fmt.Sprintf("格式错误: 括号不匹配, RunString: %s", ruleString)
		}
		return nil, fmt.Errorf("err: 【%w】, 【%s】", err, errHasOccurred)
	}

	expressCache.Store(ruleString, expression)
	return expression, nil
}

func (r *EngineLogic) replaceRuleString(ruleString string) string {
	if r.preString == "" || r.afterString == "" {
		return ruleString
	}
	return utils.ReplaceDynamicVariables(ruleString, r.preString, r.afterString, "[", "]")
}

// runOneRuleString 一个规则进行判断
func (r *EngineLogic) runOneRuleString(ruleString string, parameters map[string]any) (any, error) {
	ruleString = r.replaceRuleString(ruleString)

	expression, err := r.getExpressionByRuleString(ruleString)

	if err != nil {
		return nil, err
	}
	//需要对parameters里的Decimal类型特殊处理
	parameters = r.convertParameters(parameters)

	retVal, err := expression.Evaluate(parameters)
	if err != nil {
		//有可能对象里存在key深层嵌套的情况
		if strings.Contains(err.Error(), "No parameter") {
			newParameters := conv.KeyListFromMap(parameters)
			retVal1, err1 := expression.Evaluate(newParameters)
			if err1 == nil {
				return retVal1, nil
			}
		}
	}

	return retVal, err
}
func (r *EngineLogic) convertParameters(parameters map[string]any) map[string]any {
	newParameters := make(map[string]any)
	for key, val := range parameters {
		if valDecimal, ok := val.(decimal.Decimal); ok {
			valDecimal.CoefficientInt64()
			if valDecimal.IsInteger() {
				newParameters[key] = valDecimal.CoefficientInt64()
				continue
			}
			newParameters[key] = valDecimal.InexactFloat64()
			continue
		}
		newParameters[key] = val
	}
	return newParameters
}

func (r *EngineLogic) getRetValueKey(key string) string {
	return fmt.Sprintf("%s%s", r.retRulePrefix, key)
}

// Vars 获取变量列表
func (r *EngineLogic) Vars(ruleString string) ([]string, error) {
	if ruleString == "" {
		return []string{}, nil
	}
	ruleString = r.replaceRuleString(ruleString)

	exp, err := r.getExpressionByRuleString(ruleString)
	if err != nil {
		return nil, err
	}

	var varList []string
	for _, val := range exp.Tokens() {
		if val.Kind == govaluate.VARIABLE {
			//oneVarList := r.varCheckList(val.Value.(string))
			//需要判断是否是字符串值，如果是，则需要跳过
			varList = utils.AppendUniq(varList, val.Value.(string))
		}
	}
	return varList, nil
}

//func (r *EngineLogic) varCheckList(varString string) []string {
//	varString = strings.TrimSpace(varString)
//	//不包含"和'
//	if !(strings.HasPrefix(varString, "\"") || strings.HasPrefix(varString, "'")) {
//		return []string{varString}
//	}
//	//如果是单个字符串，而不是变量，则返回空
//	oneVar := regexp.MustCompile(`^('([^']+)')$|^("([^"]+)")$`)
//	isOneStr := oneVar.MatchString(varString)
//	if isOneStr {
//		return []string{}
//	}
//	varList := strings.Split(varString, ",")
//	allVarList := make([]string, 0)
//	for _, val := range varList {
//		oneVarList := r.varCheckList(val)
//		allVarList = append(allVarList, oneVarList...)
//	}
//	return allVarList
//}

// RunString 一个规则，返回规则的结果
// Deprecated: 请使用 ruleExpr := templates.NewRuleExprEngine();
// newArgs, _ := ruleExpr.RunString(ac.ArgTemplate, inputParams)
func (r *EngineLogic) RunString(ruleString string, parameters map[string]any) (any, error) {
	if ruleString == "" {
		return nil, nil
	}
	retVal, err := r.runOneRuleString(ruleString, parameters)
	if err != nil {
		return nil, err
	}
	return retVal, nil
}

// containsVariableDot 判断字符串中是否存在“变量路径中的点”（而非小数点的点）。
// 仅当 "." 前后均为数字时才视为小数点并忽略；其余情况（如 user.age、a.b.c）视为变量路径点。
func containsVariableDot(s string) bool {
	if !strings.Contains(s, ".") {
		return false
	}

	bracketDepth := 0 // 同时统计 [ ] 与 { } 的包裹层数
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\':
			// 转义符：跳过其后一个字符（\. 或 \[），避免被误判
			i++
			continue
		case '[':
			bracketDepth++
			continue
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
			continue
		case '.':
			// 1) 小数点：前后均为数字，忽略
			prevDigit := i > 0 && s[i-1] >= '0' && s[i-1] <= '9'
			nextDigit := i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '9'
			if prevDigit && nextDigit {
				continue
			}
			// 2) 反斜杠转义（\.）或 3) 处于 []/{{}} 包裹内：已正确处理，忽略
			if bracketDepth > 0 {
				continue
			}
			return true // 裸写的变量路径点，需要提示
		}
	}
	return false
}

// RunStringAny 一个规则，返回规则的结果
func (r *EngineLogic) RunStringAny(ruleString string, parameters any) (any, error) {
	if ruleString == "" {
		return nil, nil
	}
	if parameters == nil {
		return nil, fmt.Errorf("input nil")
	}
	var parameterMap map[string]any

	if parameterMapTemp, ok := parameters.(map[string]any); ok {
		parameterMap = parameterMapTemp
	} else {
		parameterMap = make(map[string]any)
		err := conv.Unmarshal(parameters, &parameterMap)
		if err != nil {
			return nil, fmt.Errorf("input change to map error:%w", err)
		}
	}

	return r.RunString(ruleString, parameterMap)
}

// RunRuleList 一个规则组列表，返回所有规则的结果
func (r *EngineLogic) RunRuleList(ruleList []*RuleInfo, allData map[string]any) (map[string]any, error) {
	if len(ruleList) == 0 {
		return allData, nil
	}
	for _, rule := range ruleList {
		retVal, err := r.runOneRuleString(rule.RuleString, allData)
		if err != nil {
			return allData, err
		}
		retKey := r.getRetValueKey(rule.Key)
		allData[retKey] = retVal
	}
	return allData, nil
}

// CheckLastRuleList 一个规则组列表，返回最后一条的结果，最后一条必须返回 true or false
func (r *EngineLogic) CheckLastRuleList(ruleList []*RuleInfo, allData map[string]any) (bool, error) {
	if len(ruleList) == 0 {
		return true, nil
	}
	allRetData, err := r.RunRuleList(ruleList, allData)
	if err != nil {
		return false, err
	}
	// 最后一条一定是判断是否是true或false
	lastRule := ruleList[len(ruleList)-1]
	retKey := r.getRetValueKey(lastRule.Key)
	if retVal, ok := allRetData[retKey]; ok {
		if retValBool, ok := retVal.(bool); ok {
			return retValBool, nil
		}
	}
	return false, fmt.Errorf("最后一条规则不是bool类型: key: %s, str: %s", lastRule.Key, lastRule.RuleString)
}

// CheckAllRuleList 一个规则组列表，通过 operator 将所有Rule连起来，返回结果
func (r *EngineLogic) CheckAllRuleList(ruleList []*RuleInfo, operator string, allData map[string]any) (bool, error) {
	if len(ruleList) == 0 {
		return true, nil
	}

	if operator != "&&" && operator != "||" {
		return false, fmt.Errorf("operator must: &&、||")
	}

	allRetData, err := r.RunRuleList(ruleList, allData)
	if err != nil {
		return false, err
	}
	//需要检查所有是否是bool类型
	for _, rule := range ruleList {
		retKey := r.getRetValueKey(rule.Key)
		if retVal, ok := allRetData[retKey]; ok {
			if _, ok := retVal.(bool); !ok {
				return false, fmt.Errorf("ruleString return not bool: key:%s, str: %s, real return: %v",
					rule.Key, rule.RuleString, retVal)
			}
		}
	}
	//将所有的返回通过operator连起来
	checkRuleList := make([]string, 0)
	for _, rule := range ruleList {
		checkRuleList = append(checkRuleList, r.getRetValueKey(rule.Key))
	}

	checkRuleString := fmt.Sprintf("(%s)", strings.Join(checkRuleList, fmt.Sprintf(" %s ", operator)))
	retVal, err := r.RunString(checkRuleString, allRetData)
	if err != nil {
		return false, err
	}
	if retValBool, ok := retVal.(bool); ok {
		return retValBool, nil
	}

	return false, fmt.Errorf("规则结果不是bool类型: %s, %v", checkRuleString, retVal)
}
