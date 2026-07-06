package activity

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/plugins/action"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/magic-lib/go-plat-utils/utils/httputil/param"
	"github.com/samber/lo"
)

func (ac *Activity) initArguments() {
	// 将 Arguments 中的 类型没有设置的，默认设置为仅default，前端不能进行设置
	lo.ForEach(ac.Arguments, func(arg *param.BindConfig, index int) {
		if arg.Policy == "" {
			arg.Policy = param.KeyPolicyDefaultOnly
		}
	})
}

// extractDependenciesFromArguments 从 Arguments 中自动提取依赖的 activity IDs
func (ac *Activity) extractDependenciesFromArguments(keyPrefix string) []*Activity {
	deps := make(map[string]bool)

	for _, arg := range ac.Arguments {
		if arg.Value != nil {
			valueStr := conv.String(arg.Value)
			matches := ExtractDependsActivityIds(valueStr, keyPrefix)
			for _, match := range matches {
				deps[match] = true
			}
		}
	}

	// 也从 ArgTemplate 中提取
	if ac.ArgTemplate != "" {
		matches := ExtractDependsActivityIds(ac.ArgTemplate, keyPrefix)
		for _, match := range matches {
			deps[match] = true
		}
	}

	// 条件中提取
	if ac.Control.When != "" {
		matches := ExtractDependsActivityIds(ac.Control.When, keyPrefix)
		for _, match := range matches {
			deps[match] = true
		}
	}

	var result = make([]*Activity, 0)
	for dep := range deps {
		result = utils.AppendUniq(result, &Activity{
			Id: dep,
		})
	}
	return result
}

// getAllDependencies 获取有效的依赖列表
func (ac *Activity) getAllDependencies() []*Activity {
	keyPrefix := ReturnKeyPrefix()
	oneDependsOnIdList := ac.extractDependenciesFromArguments(keyPrefix)
	if len(ac.DependsOn) == 0 {
		return oneDependsOnIdList
	}
	lo.ForEach(ac.DependsOn, func(dep *Activity, index int) {
		oneDependsOnIdList = utils.AppendUniq(oneDependsOnIdList, dep)
	})
	return oneDependsOnIdList
}

func (ac *Activity) getActionParamKeyId(inputParams map[string]any) (string, error) {
	if ac.ActionName == "" {
		return "", nil
	}
	actionFun, err := action.GetAction(ac.ActionNamespace, ac.ActionName)
	if err != nil {
		return "", err
	}

	actionParam := ac.getActionParam(inputParams)

	actData := actionFun.ActMeta()
	if actData.ArgumentType != nil {
		var data any
		var err1 error
		if data, err1 = conv.ConvertForType(actData.ArgumentType, actionParam); err1 != nil {
			return "", fmt.Errorf("arguments type does not match required type: %v", actData.ArgumentType)
		}
		return utils.UniqueJsonId(data)
	}
	return utils.UniqueJsonId(actionParam)
}
