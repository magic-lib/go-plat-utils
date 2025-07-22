package ruleengine

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	govaluate "github.com/magic-lib/go-plat-utils/internal/govaluate-3.0.0"
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
		"Add":     ruleLogicFunc.Add,
		"Sub":     ruleLogicFunc.Sub,
		"Mul":     ruleLogicFunc.Mul,
		"Div":     ruleLogicFunc.Div,
		"Has":     ruleLogicFunc.Has,
		"In":      ruleLogicFunc.In,
		"Is":      ruleLogicFunc.Is,
		"If":      ruleLogicFunc.If,
		"As":      ruleLogicFunc.As,
		"Replace": ruleLogicFunc.Replace,
		"Split":   ruleLogicFunc.Split,
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
		return nil, err
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

	return expression.Evaluate(parameters)
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
