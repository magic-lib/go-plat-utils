package action

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/plugins"
)

// Register 注册全局Action方法
func Register(ai Actor) error {
	if ai == nil {
		return fmt.Errorf("actor is nil")
	}
	am := ai.MetaData()
	if am == nil || am.Name() == "" {
		return fmt.Errorf("activity name is empty")
	}
	return plugins.Register(am.Namespace, ai)
}

// Unregister 注销全局Action方法
func Unregister(ns string, action string) error {
	return plugins.Unregister(ns, action) // 假设 plugins 包有 Unregister
}

// ListActions 列出指定命名空间下所有已注册的Action
func ListActions(ns string) ([]*ActMetaData, error) {
	list, err := plugins.LoadAll(ns)
	if err != nil {
		return nil, err
	}
	result := make([]*ActMetaData, 0, len(list))
	for _, p := range list {
		if a, ok := p.(Actor); ok {
			result = append(result, a.MetaData())
		}
	}
	return result, nil
}

// GetAction 获取Action方法
func GetAction(ns string, action string) (Actor, error) {
	ai, err := plugins.Load(ns, action)
	if err != nil {
		return nil, fmt.Errorf("action err: %v is not registered, ns:%s, action:%s", err, ns, action)
	}
	am, ok := ai.(Actor)
	if ok {
		return am, nil
	}
	return nil, fmt.Errorf("action %s is not registered, ns:%s", action, ns)
}
