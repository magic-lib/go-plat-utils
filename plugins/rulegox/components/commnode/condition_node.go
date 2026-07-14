package commnode

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates"
	"github.com/rulego/rulego"
	"github.com/rulego/rulego/api/types"
)

// init 注册ExprFilterNode组件
// init registers the ExprFilterNode component with the default registry.
func init() {
	Registry.Add(&CondRouterNode{})
	_ = rulego.Registry.Register(&CondRouterNode{})
}

type CondRouterNode struct {
	Condition string `json:"condition"`
	ruleObj   *templates.RuleExprEngine
}

// Type 返回组件类型
// Type returns the component type identifier.
func (x *CondRouterNode) Type() string {
	return "condition"
}

// New 创建新实例
// New creates a new instance.
func (x *CondRouterNode) New() types.Node {
	return &CondRouterNode{}
}

// Init 初始化组件，验证并编译表达式
// Init initializes the component.
func (x *CondRouterNode) Init(ruleConfig types.Config, configuration types.Configuration) error {
	newCond := new(CondRouterNode)
	err := conv.Unmarshal(configuration, newCond)
	if err != nil {
		return fmt.Errorf("condRouter error parsing configuration: %s, %v", conv.String(configuration), err)
	}
	x.Condition = newCond.Condition
	x.ruleObj = templates.NewRuleExprEngine()
	return nil
}

// OnMsg 处理消息，通过评估编译的表达式来过滤消息
// OnMsg processes incoming messages by evaluating the compiled expression.
func (x *CondRouterNode) OnMsg(ctx types.RuleContext, msg types.RuleMsg) {
	dataStr := msg.Data.String()
	dataMap := map[string]any{}
	err := conv.Unmarshal(dataStr, &dataMap)
	if err != nil {
		ctx.TellFailure(msg, err)
		return
	}
	condStr := x.Condition
	if condStr == "" {
		ctx.TellNext(msg) //结束流程了
		return
	}

	conResult, err := x.ruleObj.RunString(condStr, dataMap)
	if err != nil {
		ctx.TellFailure(msg, err)
		return
	}

	isBool := cond.IsBool(conResult)
	if isBool {
		boolResult, err := conv.Convert[bool](conResult)
		if err != nil {
			ctx.TellFailure(msg, err)
			return
		}
		if boolResult {
			ctx.TellNext(msg, types.True)
		} else {
			ctx.TellNext(msg, types.False)
		}
		return
	}
	if retStr, ok := conResult.(string); ok {
		ctx.TellNext(msg, retStr)
		return
	}
	// 默认使用Success和Failure
	changeBool, err := conv.Convert[bool](conResult)
	if err != nil {
		// 如果出错，则采用转为字符串来进行处理
		ctx.TellNext(msg, conv.String(conResult))
		return
	}
	// 默认采用Success和Failure
	if changeBool {
		ctx.TellSuccess(msg)
	} else {
		ctx.TellNext(msg, types.Failure)
	}
}

// Desc returns the component description
func (x *CondRouterNode) Desc() string {
	return "Routes to True/False. Variables: id, ts, data, msg, metadata, type, dataType"
}

// Destroy 清理资源
// Destroy cleans up resources.
func (x *CondRouterNode) Destroy() {
}
