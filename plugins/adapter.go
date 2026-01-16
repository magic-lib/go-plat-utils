package plugins

import (
	"context"
	"fmt"
)

var _ Plugin = (*CommPlugin)(nil)

var (
	commPluginManager = NewPluginManager()
)

type CommPlugin struct {
	Namespace  string `yaml:"namespace" json:"namespace"`
	PluginName string `yaml:"plugin_name" json:"plugin_name"`
	method     Method
}

func NewCommPlugin(namespace, name string, method Method) (*CommPlugin, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}
	if method == nil {
		return nil, fmt.Errorf("method cannot be empty")
	}
	return &CommPlugin{
		Namespace:  namespace,
		PluginName: name,
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

func Register(onePlugin Plugin) error {
	return commPluginManager.Register(onePlugin)
}

func Load(name string) (Plugin, error) {
	onePlugin, err := commPluginManager.Load(name)
	if err != nil {
		return nil, err
	}
	if onePlugin == nil {
		return nil, fmt.Errorf("插件 %s 未注册", name)
	}
	return onePlugin, nil
}
