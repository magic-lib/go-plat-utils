package action

import (
	"context"
	"errors"
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/plugins"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/samber/lo"
	"github.com/tidwall/gjson"
	"log"
	"reflect"
)

type (
	Actor interface {
		plugins.Plugin
		ActMeta() *ActionMeta //定义具体的动作属性
	}
)

const (
	namespaceLinkNameChar = "." // 命名之间的默认连接字符
)

type (
	// ActionMeta 动作的元数据配置（集中管理所有描述性字段）
	ActionMeta struct {
		Namespace    string                  `yaml:"namespace" json:"namespace,omitempty"` // 避免相同的activity名称冲突，默认为空，则为顶级方法
		Action       string                  `yaml:"action" json:"action"`                 // 活动名,对应执行的相应方法,可以自定义名
		Desc         string                  `yaml:"desc" json:"desc"`                     // 动作描述
		RequiredArgs []string                `yaml:"required_args" json:"required_args"`   // 必传参数键，主要是做前期检查用，避免调用方法后才报错, 是jsonPath
		RequiredResp []string                `yaml:"required_resp" json:"required_resp"`   // 返回参数元数据，做检查用
		ArgumentType reflect.Type            `yaml:"-" json:"-"`                           // 输入参数类型，这样可以保证使用golang的类型
		actionMethod utils.ContextAnyHandler //action执行的具体方法
		linkChar     string
	}
)

// 检查输入参数是否符合要求
func (am *ActionMeta) checkArguments(arguments any) error {
	if am.Name() == "" {
		return errors.New("action name cannot be empty")
	}

	// 检查参数类型是否匹配
	if am.ArgumentType != nil {
		if _, err1 := conv.ConvertForType(am.ArgumentType, arguments); err1 != nil {
			return fmt.Errorf("arguments type does not match required type: %v", am.ArgumentType)
		}
	}

	// 检查必填参数：优先用 gjson（支持模板变量 {{.xxx}}），
	// 如果 arguments 是结构体，转为 JSON 后再检查
	if len(am.RequiredArgs) > 0 {
		missingArgs := am.findMissingRequiredArgs(arguments)
		if len(missingArgs) > 0 {
			return fmt.Errorf("required: %s", conv.String(missingArgs))
		}
	}

	return nil
}

// 检查返回数据是否符合要求
func (am *ActionMeta) checkResponses(retData any) error {
	if len(am.RequiredResp) == 0 {
		return nil
	}

	missingFields := am.findMissingRequiredFields(retData)
	if len(missingFields) > 0 {
		return fmt.Errorf("%s missing required response fields: %v", am.FullName(), missingFields)
	}
	return nil
}

func (am *ActionMeta) Name() string {
	return am.Action
}
func (am *ActionMeta) ActMeta() *ActionMeta {
	return am
}
func (am *ActionMeta) FullName() string {
	if am.Namespace == "" {
		return am.Name()
	}
	if am.linkChar == "" {
		am.linkChar = namespaceLinkNameChar
	}
	return am.Namespace + am.linkChar + am.Name()
}
func (am *ActionMeta) SetMethod(mt utils.ContextAnyHandler) *ActionMeta {
	am.actionMethod = mt
	return am
}
func (am *ActionMeta) SetLinkChar(linkChar string) *ActionMeta {
	am.linkChar = linkChar
	return am
}

// Execute 执行动作并返回结果
func (am *ActionMeta) Execute(ctx context.Context, arguments any) (any, error) {
	// 1. 验证输入参数
	if err := am.checkArguments(arguments); err != nil {
		return nil, fmt.Errorf("%s invalid arguments: %w", am.FullName(), err)
	}

	if am.actionMethod == nil {
		return nil, fmt.Errorf("%s failed to load method", am.FullName())
	}

	// 3. 转换参数并执行动作
	retData, err := am.executeAction(ctx, am.actionMethod, arguments)
	if err != nil {
		log.Printf("%s execution failed: %s", am.FullName(), err.Error())
		return retData, err
	}

	// 4. 验证返回数据
	if err = am.checkResponses(retData); err != nil {
		return retData, fmt.Errorf("%s invalid response data: %w", am.FullName(), err)
	}

	return retData, nil
}

// 执行动作的内部方法，处理参数转换
func (am *ActionMeta) executeAction(ctx context.Context, action utils.ContextAnyHandler, arguments any) (any, error) {
	// 如果指定了参数类型且可以转换，则使用转换后的参数
	if am.ArgumentType != nil {
		if convertedArgs, err := conv.ConvertForType(am.ArgumentType, arguments); err == nil {
			return action(ctx, convertedArgs)
		}
	}
	// 使用原始参数执行
	return action(ctx, arguments)
}

// 查找缺失的必填参数
func (am *ActionMeta) findMissingRequiredArgs(arguments any) []string {
	return findMissingRequiredFields(arguments, am.RequiredArgs)
}

// 查找缺失的必填返回字段
func (am *ActionMeta) findMissingRequiredFields(retData any) []string {
	return findMissingRequiredFields(retData, am.RequiredResp)
}
func findMissingRequiredFields(retData any, requiredList []string) []string {
	jsonStr := conv.String(retData)
	if !cond.IsJson(jsonStr) || len(requiredList) == 0 {
		return nil
	}

	missing := make([]string, 0)
	lo.ForEach(requiredList, func(config string, _ int) {
		if config == "" {
			return
		}

		if !gjson.Get(jsonStr, config).Exists() {
			missing = append(missing, config)
		}
	})

	return missing
}
