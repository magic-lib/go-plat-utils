package commnode

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/plugins/action"
	"github.com/magic-lib/go-plat-utils/plugins/activity"
	"github.com/magic-lib/go-plat-utils/plugins/paramx"
	"github.com/rulego/rulego"
	"github.com/rulego/rulego/api/types"
	"github.com/rulego/rulego/components/base"
	"github.com/samber/lo"
	"go.uber.org/multierr"
)

type ActivityNode struct {
	Activity  *activity.Activity `json:"activity"`
	Arguments map[string]any     `json:"arguments"` //参数直接映射替换
	Responses map[string]any     `json:"responses"` //参数直接映射替换
}

// RegisterActivityNodes 将传入的 Activity 动态注册为 rulego 节点。
// 若目标 type 已存在则直接返回错误，不会覆盖已有节点。
// 返回注册成功的 type 字符串，可在 rule chain JSON 的 "type" 字段直接使用。
func RegisterActivityNodes(actList ...*activity.Activity) error {
	if len(actList) == 0 {
		return nil
	}
	var retErr error
	for _, act := range actList {
		if err := registerOneActivityNode(act); err != nil {
			retErr = multierr.Append(retErr, err)
		}
	}
	return retErr
}
func registerOneActivityNode(act *activity.Activity) error {
	if act == nil {
		return fmt.Errorf("activity is nil")
	}
	typeName := act.Type()
	if typeName == "" {
		return fmt.Errorf("activity type is empty, please set ActNamespace/ActName or ActivityType")
	}

	for k, _ := range rulego.Registry.GetComponents() {
		if k == typeName {
			return fmt.Errorf("activity node type %q already registered", typeName)
		}
	}
	existing := lo.FindOrElse(Registry.Components(), nil, func(item types.Node) bool {
		return item.Type() == typeName
	})
	if existing != nil {
		return fmt.Errorf("activity node type %q already registered", typeName)
	}

	node := &ActivityNode{Activity: act}
	_ = rulego.Registry.Register(node) // 全局注册表，rulego.New 默认使用
	Registry.Add(node)                 // 本地注册表，与 init 保持一致
	return nil
}

// Type 返回组件类型
// Type returns the component type identifier.
func (x *ActivityNode) Type() string {
	if x.Activity != nil {
		if t := x.Activity.Type(); t != "" {
			return t
		}
	}
	return "activity"
}

// New 创建新实例
// New creates a new instance.
func (x *ActivityNode) New() types.Node {
	if x.Activity == nil {
		return &ActivityNode{}
	}
	oneActivityStr := conv.String(x.Activity)
	newActivity := new(activity.Activity)
	_ = conv.Unmarshal(oneActivityStr, newActivity)
	return &ActivityNode{
		Activity: newActivity,
	}
}

// Init 初始化组件，验证并编译表达式
// Init initializes the component.
func (x *ActivityNode) Init(ruleConfig types.Config, configuration types.Configuration) error {
	if len(configuration) == 0 {
		return nil
	}
	ruleNode := base.NodeUtils.GetSelfDefinition(configuration.Copy())
	newCond := ActivityNode{}
	if err := conv.Unmarshal(ruleNode.Configuration, &newCond); err != nil {
		return fmt.Errorf("activityNode error parsing newCond: %s, %v", conv.String(configuration), err)
	}
	newCond.Activity = &activity.Activity{}
	if err := conv.Unmarshal(ruleNode, newCond.Activity); err != nil {
		return fmt.Errorf("activityNode error parsing configuration: %s, %v", conv.String(configuration), err)
	}

	if newCond.Activity != nil && newCond.Activity.ActName != "" {
		oldAct := x.Activity
		x.Activity = newCond.Activity
		//Arguments    []*param.BindConfig `yaml:"arguments" json:"arguments,omitempty"`
		//Responses    map[string]any      `yaml:"responses" json:"responses"`                 // 返回的参数map，可以自定义添加内容，比如命名转换
		x.Activity.ActivityType = oldAct.ActivityType
		x.Activity.ActNamespace = oldAct.ActNamespace
		x.Activity.ActName = oldAct.ActName
		if x.Activity.ArgTemplate == "" {
			x.Activity.ArgTemplate = oldAct.ArgTemplate
		}
	}

	x.Arguments = newCond.Arguments
	x.Responses = newCond.Responses

	return nil
}

// OnMsg 处理消息，通过评估编译的表达式来过滤消息
// OnMsg processes incoming messages by evaluating the compiled expression.
func (x *ActivityNode) OnMsg(ctx types.RuleContext, msg types.RuleMsg) {
	if x.Activity == nil {
		ctx.TellFailure(msg, fmt.Errorf("activityNode has no activity"))
		return
	}
	newAct := x.Activity.Clone()
	newAct.Id = ctx.GetSelfId()

	actionFun, err := action.GetAction(newAct.ActNamespace, newAct.ActName)
	if err != nil {
		ctx.TellFailure(msg, err)
		return
	}

	actMeta := actionFun.MetaData()
	if msg.Metadata == nil {
		msg.Metadata = types.NewMetadata()
	}
	var meta map[string]any
	if err = conv.Unmarshal(actMeta, &meta); err == nil {
		for k, v := range meta {
			msg.Metadata.PutValue(k, conv.String(v))
		}
	}

	dataMap, data, err := x.getStepArguments(newAct.Id, msg)
	if err != nil {
		ctx.TellFailure(msg, err)
		return
	}

	respData, err := newAct.Execute(ctx.GetContext(), dataMap)
	if err != nil {
		ctx.TellFailure(msg, err)
		return
	}
	if oneFuncData, ok := respData[newAct.Id]; ok {
		oneFuncParam := new(paramx.ParamStruct)
		_ = conv.Unmarshal(oneFuncData, oneFuncParam)
		data.Steps[newAct.Id] = *oneFuncParam
	}

	msg.SetData(conv.String(data))
	ctx.TellSuccess(msg)
}

func (x *ActivityNode) getStepArguments(stepId string, msg types.RuleMsg) (map[string]any, *paramx.ParamCtx, error) {
	data := new(paramx.ParamCtx)
	if err := conv.Unmarshal(msg.GetData(), &data); err != nil {
		return nil, nil, err
	}

	dataMap := data.StepMapsByStepId(stepId)
	if len(x.Arguments) > 0 {
		oneArgs, err := data.TemplateArguments(x.Arguments)
		if err == nil {
			for k, v := range oneArgs {
				dataMap[k] = v
			}
		} else {
			for k, v := range x.Arguments {
				dataMap[k] = v
			}
		}
	}
	return dataMap, data, nil
}

// Desc returns the component description
func (x *ActivityNode) Desc() string {
	return "ActivityNode: " + x.Type() + " to Failure/Success."
}

// Destroy 清理资源
// Destroy cleans up resources.
func (x *ActivityNode) Destroy() {
}
