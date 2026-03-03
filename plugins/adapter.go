package plugins

import (
	"context"
	"fmt"
	"github.com/magic-lib/go-plat-utils/utils"
)

var _ Plugin = (*CommPlugin)(nil)

type CommPlugin struct {
	Namespace  string `yaml:"namespace" json:"namespace"`
	PluginName string `yaml:"plugin_name" json:"plugin_name"`
	method     utils.ContextAnyHandler
}

func NewCommPlugin(namespace, pluginName string, method utils.ContextAnyHandler) (*CommPlugin, error) {
	if pluginName == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if method == nil {
		return nil, fmt.Errorf("method cannot be empty")
	}
	return &CommPlugin{
		Namespace:  namespace,
		PluginName: pluginName,
		method:     method,
	}, nil
}

func (m *CommPlugin) Name() string {
	if m.Namespace == "" {
		return m.PluginName
	}
	return fmt.Sprintf("%s/%s", m.Namespace, m.PluginName)
}
func (m *CommPlugin) Execute(ctx context.Context, args any) (any, error) {
	return m.method(ctx, args)
}
