package param

import (
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/samber/lo"
)

// BindConfig 参数绑定配置
type BindConfig struct {
	Key    string          `yaml:"key" json:"key"`
	Value  any             `yaml:"value" json:"value"`
	Type   string          `json:"type,omitempty"` //值类型：string/int64/int/bool
	Policy KeySourcePolicy `yaml:"policy" json:"policy"`
}

// MergeArgumentsByBinding 通过参数绑定进行参数合并
func MergeArgumentsByBinding(args map[string]any, binds []*BindConfig) map[string]any {
	if len(binds) == 0 { //没有默认配置，则直接返回
		return args
	}
	keyFieldConfigs := make(map[string]KeySourcePolicy)
	backendConfig := make(map[string]any)
	lo.ForEach(binds, func(bind *BindConfig, _ int) {
		if bind.Policy == "" {
			bind.Policy = KeyPolicyFrontendPriority
		}
		keyFieldConfigs[bind.Key] = bind.Policy
		// 根据 Type 将配置值转换为最终类型后写入后端配置；Type 为空则原样保留
		backendConfig[bind.Key], _ = conv.ConvertForTypeString(bind.Type, bind.Value)
	})
	configManager := NewDynamicConfigManager(keyFieldConfigs)
	return configManager.MergeMap(args, backendConfig, backendConfig)
}
