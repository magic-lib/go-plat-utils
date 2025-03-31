package plugins

import (
	"context"
	"fmt"
	"reflect"
)

// 插件注册表，存储插件类型
//var pluginRegistry = make(map[string]reflect.Type)

// PluginManager 插件管理器
type PluginManager struct {
	plugins map[string]Plugin
}

// NewPluginManager 创建新的插件管理器
func NewPluginManager() *PluginManager {
	return &PluginManager{plugins: make(map[string]Plugin)}
}

// RegisterPlugin 注册插件
func (pm *PluginManager) RegisterPlugin(plugin Plugin) error {
	if plugin == nil {
		return fmt.Errorf("插件不能为空")
	}
	if plugin.Name() == "" {
		return fmt.Errorf("插件Name不能为空")
	}
	_, exists := pm.plugins[plugin.Name()]
	if exists {
		return fmt.Errorf("插件 %s 已注册", plugin.Name())
	}
	//pluginType := reflect.TypeOf(plugin).Elem()
	pm.plugins[plugin.Name()] = plugin
	return nil
}

// LoadPlugin 加载插件
func (pm *PluginManager) LoadPlugin(name string) error {
	onePlugin, exists := pm.plugins[name]
	if !exists {
		return fmt.Errorf("插件 %s 未注册", name)
	}
	pluginType := reflect.TypeOf(onePlugin).Elem()
	pluginInstance := reflect.New(pluginType).Interface().(Plugin)
	pm.plugins[name] = pluginInstance
	return nil
}

// ExecutePlugin 执行插件
func (pm *PluginManager) ExecutePlugin(ctx context.Context, name string, args any) (any, error) {
	plugin, exists := pm.plugins[name]
	if !exists {
		return nil, fmt.Errorf("插件 %s 未加载", name)
	}
	return plugin.Execute(ctx, args)
}
