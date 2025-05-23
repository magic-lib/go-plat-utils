package ruleengine

import (
	"fmt"
	"github.com/Knetic/govaluate"
	"github.com/magic-lib/go-plat-utils/utils"
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
	ruleLogicFunc := &customerFunc{}
	// 内置方法
	ruleLogic.functions = map[string]govaluate.ExpressionFunction{
		"AddByNumber": ruleLogicFunc.AddByNumber,
		"SubByNumber": ruleLogicFunc.SubByNumber,
		"MulByNumber": ruleLogicFunc.MulByNumber,
		"DivByNumber": ruleLogicFunc.DivByNumber,
		"Has":         ruleLogicFunc.Has,
		"In":          ruleLogicFunc.In,
		"Is":          ruleLogicFunc.Is,
		"If":          ruleLogicFunc.If,
	}
	return ruleLogic
}

// SetRetRulePrefix 设置返回值的key前缀
func (r *EngineLogic) SetRetRulePrefix(prefix string) *EngineLogic {
	if prefix == "" {
		prefix = defaultRulePrefix
	}
	r.retRulePrefix = prefix
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

// SetDelimitedString 设置变量包裹字符串
func (r *EngineLogic) SetDelimitedString(pre, after string) *EngineLogic {
	if pre == "" || after == "" {
		return r
	}
	r.preString = pre
	r.afterString = after
	return r
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
func (r *EngineLogic) runOneRuleString(ruleString string, parameters map[string]interface{}) (interface{}, error) {
	ruleString = r.replaceRuleString(ruleString)

	expression, err := r.getExpressionByRuleString(ruleString)

	if err != nil {
		return nil, err
	}

	return expression.Evaluate(parameters)
}

func (r *EngineLogic) getRetValueKey(key string) string {
	return fmt.Sprintf("%s%s", r.retRulePrefix, key)
}

// Vars 获取变量列表
func (r *EngineLogic) Vars(ruleString string) ([]string, error) {
	if ruleString == "" {
		return []string{}, nil
	}
	exp, err := r.getExpressionByRuleString(ruleString)
	if err != nil {
		return nil, err
	}
	return exp.Vars(), nil
}

// RunOneRuleString 一个规则，返回所有规则的结果
func (r *EngineLogic) RunOneRuleString(ruleString string, parameters map[string]interface{}) (interface{}, error) {
	if ruleString == "" {
		return nil, nil
	}
	retVal, err := r.runOneRuleString(ruleString, parameters)
	if err != nil {
		return nil, err
	}
	return retVal, nil
}

// RunRuleList 一个规则组列表，返回所有规则的结果
func (r *EngineLogic) RunRuleList(ruleList []*RuleInfo, allData map[string]interface{}) (map[string]interface{}, error) {
	if len(ruleList) == 0 {
		return allData, nil
	}
	for _, rule := range ruleList {
		retVal, err := r.runOneRuleString(rule.RuleString, allData)
		if err != nil {
			return allData, err
		}
		allData[r.getRetValueKey(rule.Key)] = retVal
	}
	return allData, nil
}

// CheckLastRuleByList 一个规则组列表，返回最后一条的结果，最后一条必须返回 true or false
func (r *EngineLogic) CheckLastRuleByList(ruleList []*RuleInfo, allData map[string]interface{}) (bool, error) {
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
func (r *EngineLogic) CheckAllRuleList(ruleList []*RuleInfo, operator string, allData map[string]interface{}) (bool, error) {
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
	retVal, err := r.RunOneRuleString(checkRuleString, allRetData)
	if err != nil {
		return false, err
	}
	if retValBool, ok := retVal.(bool); ok {
		return retValBool, nil
	}

	return false, fmt.Errorf("规则结果不是bool类型: %s, %v", checkRuleString, retVal)
}
