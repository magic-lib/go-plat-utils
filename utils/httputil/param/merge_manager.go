package param

import (
	"log"
	"reflect"
)

type DynamicConfigManager struct {
	rules map[string]SourcePolicy
}

func NewDynamicConfigManager(keyFieldConfigs map[string]KeySourcePolicy) *DynamicConfigManager {
	rules := make(map[string]SourcePolicy)
	for key, cfg := range keyFieldConfigs {
		policy, err := ParseKeyPolicy(cfg)
		if err != nil {
			log.Printf("无效的策略: %s", cfg)
			continue
		}
		rules[key] = policy
	}
	return &DynamicConfigManager{rules: rules}
}

func (m *DynamicConfigManager) MergeMap(frontend, backend, defaults map[string]any) map[string]any {
	if len(frontend) == 0 {
		frontend = make(map[string]any)
	}
	if len(backend) == 0 {
		backend = make(map[string]any)
	}
	if len(defaults) == 0 {
		defaults = make(map[string]any)
	}
	newMap := make(map[string]any)

	// 收集所有需要处理的 key（来自 frontend、backend、defaults 三个 map）
	allKeys := make(map[string]bool)
	for k := range frontend {
		allKeys[k] = true
	}
	for k := range backend {
		allKeys[k] = true
	}
	for k := range defaults {
		allKeys[k] = true
	}

	// 遍历所有 key，根据配置的规则进行合并
	for key := range allKeys {
		// 获取三个来源的值
		vFront, hasFront := frontend[key]
		vBack, hasBack := backend[key]
		vDef, hasDef := defaults[key]

		// 将值转换为 reflect.Value 用于策略执行
		var frontReflect, backReflect, defReflect reflect.Value
		if hasFront {
			frontReflect = reflect.ValueOf(vFront)
		}
		if hasBack {
			backReflect = reflect.ValueOf(vBack)
		}
		if hasDef {
			defReflect = reflect.ValueOf(vDef)
		}

		// 检查是否配置了该 key 的策略规则
		policy, hasRule := m.rules[key]
		if hasRule {
			// 有配置规则，使用对应的策略执行函数
			executor := strategyExecutors[policy]
			finalVal := executor(frontReflect, backReflect, defReflect)
			if finalVal.IsValid() {
				newMap[key] = finalVal.Interface()
				continue
			}
		}

		// 没有配置规则时，使用默认行为：后端 > 前端 > 默认值
		if hasBack && (backReflect.IsValid() && !backReflect.IsZero()) {
			newMap[key] = vBack
			continue
		} else if hasFront && (frontReflect.IsValid() && !frontReflect.IsZero()) {
			newMap[key] = vFront
			continue
		} else if hasDef && (defReflect.IsValid() && !defReflect.IsZero()) {
			newMap[key] = vDef
			continue
		}
		// 直接使用谁的值即可，无论是否非法
		if hasBack {
			newMap[key] = vBack
			continue
		} else if hasFront {
			newMap[key] = vFront
			continue
		} else if hasDef {
			newMap[key] = vDef
			continue
		}
	}
	return newMap
}
