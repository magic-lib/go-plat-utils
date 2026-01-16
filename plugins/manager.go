package plugins

import (
	"context"
	"fmt"
	cmapv2 "github.com/orcaman/concurrent-map/v2"
)

// 插件注册表，存储插件类型
//var pluginRegistry = make(map[string]reflect.Type)

// PluginManager 插件管理器
type PluginManager struct {
	plugins cmapv2.ConcurrentMap[string, Plugin]
}

// NewPluginManager 创建新的插件管理器
func NewPluginManager() *PluginManager {
	onePlugins := cmapv2.New[Plugin]()
	return &PluginManager{
		plugins: onePlugins,
	}
}

// Register 注册插件
func (pm *PluginManager) Register(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("插件不能为空")
	}
	pluginKey := plugin.Name()
	if pluginKey == "" {
		return fmt.Errorf("插件Name不能为空")
	}
	// 不能重复注册，避免覆盖
	if pm.plugins.Has(pluginKey) {
		return fmt.Errorf("PluginManager %s is already registered", pluginKey)
	}
	pm.plugins.Set(pluginKey, plugin)
	return nil
}

// Load 加载插件
func (pm *PluginManager) Load(name string) (Plugin, error) {
	onePlugin, exists := pm.plugins.Get(name)
	if !exists {
		return nil, fmt.Errorf("插件 %s 未注册", name)
	}
	return onePlugin, nil
}

// Execute 执行插件
func (pm *PluginManager) Execute(ctx context.Context, name string, args any) (any, error) {
	onePlugin, err := pm.Load(name)
	if err != nil {
		return nil, err
	}
	return onePlugin.Execute(ctx, args)
}
