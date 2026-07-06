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
	onePluginManager, err := getOnePluginManagerFromComm(namespace)
	if err != nil {
		return err
	}
	if onePluginManager == nil {
		onePluginManager = NewPluginManager()
	}
	err = onePluginManager.Register(onePlugin)
	if err != nil {
		return err
	}
	commPluginManager.Set(namespace, onePluginManager)
	return nil
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
