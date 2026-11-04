package plugins

import (
	"fmt"
	cmapv2 "github.com/orcaman/concurrent-map/v2"
)

var (
	// 所有类型的插件管理器
	commPluginManager = cmapv2.New[*PluginManager]()
)

func getNamespace(namespace string) string {
	if namespace == "" {
		return "default"
	}
	return namespace
}
func getOnePluginManagerFromComm(namespace string) (*PluginManager, error) {
	namespace = getNamespace(namespace)
	if !commPluginManager.Has(namespace) {
		return nil, nil
	}
	onePluginManager, ok := commPluginManager.Get(namespace)
	if !ok {
		return nil, fmt.Errorf("插件管理器 %s 注册错误", namespace)
	}
	return onePluginManager, nil
}

func Register(namespace string, onePlugin Plugin) error {
	// 写入与查找使用同一个 namespace（空值归一为 default），否则空 namespace 注册后永远 Load 不到；
	// Upsert 原子地取出或创建管理器，避免并发注册同一新 namespace 时互相覆盖而丢插件
	namespace = getNamespace(namespace)
	onePluginManager := commPluginManager.Upsert(namespace, nil,
		func(exist bool, valueInMap *PluginManager, _ *PluginManager) *PluginManager {
			if exist && valueInMap != nil {
				return valueInMap
			}
			return NewPluginManager()
		})
	return onePluginManager.Register(onePlugin)
}
func Unregister(namespace string, pluginName string) error {
	onePluginManager, err := getOnePluginManagerFromComm(namespace)
	if err != nil {
		return err
	}
	if onePluginManager == nil {
		return nil
	}
	onePluginManager.Unregister(pluginName)
	return nil
}

func Load(namespace string, pluginName string) (Plugin, error) {
	onePluginManager, err := getOnePluginManagerFromComm(namespace)
	if err != nil {
		return nil, err
	}
	if onePluginManager == nil {
		return nil, fmt.Errorf("插件管理器 %s 未注册", namespace)
	}
	onePlugin, err := onePluginManager.Load(pluginName)
	if err != nil {
		return nil, err
	}
	if onePlugin == nil {
		return nil, fmt.Errorf("插件 %s / %s 未注册", namespace, pluginName)
	}
	return onePlugin, nil
}
func LoadAll(namespace string) ([]Plugin, error) {
	onePluginManager, err := getOnePluginManagerFromComm(namespace)
	if err != nil {
		return nil, err
	}
	if onePluginManager == nil {
		return nil, fmt.Errorf("插件管理器 %s 未注册", namespace)
	}
	return onePluginManager.LoadAll(), nil
}
