package param

import (
	"github.com/samber/lo"
)

// BindConfig 参数绑定配置
type BindConfig struct {
	Key    string          `yaml:"key" json:"key"`
	Value  any             `yaml:"value" json:"value"`
	Policy KeySourcePolicy `yaml:"policy" json:"policy"`
}

// MergeArgumentsByBinding 通过参数绑定进行参数合并
func MergeArgumentsByBinding(args map[string]any, binds []*BindConfig) map[string]any {
	if len(binds) == 0 { //没有默认配置，则直接返回
		return args
	}
	keyFieldConfigs := make(map[string]KeySourcePolicy)
	backendConfig := make(map[string]any)
	lo.ForEach(binds, func(arg *BindConfig, _ int) {
		if arg.Policy == "" {
			arg.Policy = KeyPolicyFrontendPriority
		}
		keyFieldConfigs[arg.Key] = arg.Policy
		backendConfig[arg.Key] = arg.Value
	})
	configManager := NewDynamicConfigManager(keyFieldConfigs)
	return configManager.MergeMap(args, backendConfig, backendConfig)
}
