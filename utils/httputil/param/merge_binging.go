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
	lo.ForEach(binds, func(arg *BindConfig, _ int) {
		if arg.Policy == "" {
			arg.Policy = KeyPolicyFrontendPriority
		}
		keyFieldConfigs[arg.Key] = arg.Policy
		// 根据 Type 将配置值转换为最终类型后写入后端配置；Type 为空则原样保留
		backendConfig[arg.Key] = convertByType(arg.Value, arg.Type)
	})
	configManager := NewDynamicConfigManager(keyFieldConfigs)
	return configManager.MergeMap(args, backendConfig, backendConfig)
}

// convertByType 按声明的类型把任意值转换成最终类型。
// 支持的 type：string / number / boolean / json；type 为空时原样返回。
// 转换失败时回退为转换前的原始值，保证不丢数据。
func convertByType(raw any, typ string) any {
	if typ == "" {
		return raw
	}
	switch typ {
	case "string":
		return conv.String(raw)
	case "int64":
		num, err := conv.Convert[int64](raw)
		if err == nil {
			return num
		}
	case "bool":
		boolTemp, err := conv.Convert[bool](raw)
		if err == nil {
			return boolTemp
		}
	}
	return raw
}
