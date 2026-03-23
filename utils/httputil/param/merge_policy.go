package param

import (
	"fmt"
	"reflect"
	"strings"
)

// SourcePolicy 定义配置项的来源优先级策略
type SourcePolicy int
type KeySourcePolicy string

const (
	// PolicyFrontendOnly 前端强制模式
	// 逻辑：前端传参 > 默认值。后台配置会被直接忽略。
	// 场景：纯前端交互逻辑，后端无需干预。
	PolicyFrontendOnly SourcePolicy = iota

	// PolicyBackendOnly 后台强制模式
	// 逻辑：后台配置 > 默认值。前端传参会被直接忽略。
	// 场景：安全限制、系统阈值。
	PolicyBackendOnly

	// PolicyFrontendPriority 前端优先模式
	// 逻辑：前端传参 > 后台配置 > 默认值。
	// 场景：用户个性化设置（分页、排序），后台仅提供默认建议。
	PolicyFrontendPriority

	// PolicyBackendPriority 后台优先模式
	// 逻辑：后台配置 > 前端传参 > 默认值。
	// 场景：业务规则，管理员可强制覆盖用户的临时设置。
	PolicyBackendPriority

	// PolicyDefaultOnly 仅默认值模式
	// 逻辑：只使用代码中的默认值。前端和后台的传参均被忽略。
	// 场景：系统固定常量。
	PolicyDefaultOnly
)

// 策略字符串常量
const (
	KeyPolicyFrontendOnly     KeySourcePolicy = "FRONTEND_ONLY"
	KeyPolicyBackendOnly      KeySourcePolicy = "BACKEND_ONLY"
	KeyPolicyFrontendPriority KeySourcePolicy = "FRONTEND_PRIORITY"
	KeyPolicyBackendPriority  KeySourcePolicy = "BACKEND_PRIORITY"
	KeyPolicyDefaultOnly      KeySourcePolicy = "DEFAULT_ONLY"
)

var policyMap = map[KeySourcePolicy]SourcePolicy{
	KeyPolicyFrontendOnly:     PolicyFrontendOnly,
	KeyPolicyBackendOnly:      PolicyBackendOnly,
	KeyPolicyFrontendPriority: PolicyFrontendPriority,
	KeyPolicyBackendPriority:  PolicyBackendPriority,
	KeyPolicyDefaultOnly:      PolicyDefaultOnly,
}

// ParseKeyPolicy 将配置文件的字符串解析为内部策略
func ParseKeyPolicy(s KeySourcePolicy) (SourcePolicy, error) {
	sTemp := string(s)
	sTemp = strings.ToUpper(strings.TrimSpace(sTemp))
	s = KeySourcePolicy(sTemp)
	policy, ok := policyMap[s]
	if !ok {
		return -1, fmt.Errorf("未知的策略配置: '%s'。有效值: %v", s, getValidKeys())
	}
	return policy, nil
}

func getValidKeys() []KeySourcePolicy {
	keys := make([]KeySourcePolicy, 0, len(policyMap))
	for k := range policyMap {
		keys = append(keys, k)
	}
	return keys
}

// String 方法方便打印调试
func (p SourcePolicy) String() string {
	switch p {
	case PolicyFrontendOnly:
		return "前端强制 (Frontend Only)"
	case PolicyBackendOnly:
		return "后台强制 (Backend Only)"
	case PolicyFrontendPriority:
		return "前端优先 (Frontend Priority)"
	case PolicyBackendPriority:
		return "后台优先 (Backend Priority)"
	case PolicyDefaultOnly:
		return "仅默认值 (Default Only)"
	default:
		return "未知策略"
	}
}

// mergeStrategyFunc 定义通用的合并逻辑函数签名
// 参数: defaultValue, frontendValue, backendValue (均为 reflect.Value)
// 返回: 最终生效的 reflect.Value
type mergeStrategyFunc func(frontVal, backVal, defVal reflect.Value) reflect.Value

// 策略映射表：将枚举映射到具体的执行函数
var strategyExecutors = map[SourcePolicy]mergeStrategyFunc{
	PolicyFrontendOnly: func(front, back, def reflect.Value) reflect.Value {
		if front.IsValid() && !front.IsZero() {
			return front
		}
		return def
	},
	PolicyBackendOnly: func(front, back, def reflect.Value) reflect.Value {
		if back.IsValid() && !back.IsZero() {
			return back
		}
		return def
	},
	PolicyFrontendPriority: func(front, back, def reflect.Value) reflect.Value {
		if front.IsValid() && !front.IsZero() {
			return front
		}
		if back.IsValid() && !back.IsZero() {
			return back
		}
		return def
	},
	PolicyBackendPriority: func(front, back, def reflect.Value) reflect.Value {
		if back.IsValid() && !back.IsZero() {
			return back
		}
		if front.IsValid() && !front.IsZero() {
			return front
		}
		return def
	},
	PolicyDefaultOnly: func(front, back, def reflect.Value) reflect.Value {
		return def
	},
}
