package plugins

import (
	"fmt"
	cmapv2 "github.com/orcaman/concurrent-map/v2"
)

var (
	// 所有类型的插件管理器
	commPluginManager = cmapv2.New[*PluginManager]()
)

func Register(namespace string, onePlugin Plugin) error {
	if namespace == "" {
		namespace = "default"
	}
	var onePluginManager *PluginManager
	if !commPluginManager.Has(namespace) {
		onePluginManager = NewPluginManager()
	} else {
		var ok bool
		onePluginManager, ok = commPluginManager.Get(namespace)
		if !ok {
			return fmt.Errorf("插件管理器 %s 注册错误", namespace)
		}
	}

	err := onePluginManager.Register(onePlugin)
	if err != nil {
		return err
	}
	commPluginManager.Set(namespace, onePluginManager)
	return nil
}

func Load(namespace string, pluginKey string) (Plugin, error) {
	if namespace == "" {
		namespace = "default"
	}
	if !commPluginManager.Has(namespace) {
		return nil, fmt.Errorf("插件管理器 %s 未注册", namespace)
	}
	onePluginManager, ok := commPluginManager.Get(namespace)
	if !ok {
		return nil, fmt.Errorf("插件管理器 %s 注册错误", namespace)
	}
	onePlugin, err := onePluginManager.Load(pluginKey)
	if err != nil {
		return nil, err
	}
	if onePlugin == nil {
		return nil, fmt.Errorf("插件 %s 未注册", pluginKey)
	}
	return onePlugin, nil
}
func LoadAll(namespace string) ([]Plugin, error) {
	if namespace == "" {
		namespace = "default"
	}
	if !commPluginManager.Has(namespace) {
		return nil, fmt.Errorf("插件管理器 %s 未注册", namespace)
	}
	onePluginManager, ok := commPluginManager.Get(namespace)
	if !ok {
		return nil, fmt.Errorf("插件管理器 %s 注册错误", namespace)
	}
	return onePluginManager.LoadAll(), nil
}
