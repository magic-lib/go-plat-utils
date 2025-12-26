package plugins

import (
	"context"
)

type Plugin interface {
	// Name 返回插件的名称，应保证唯一。
	Name() string                                       //插件的英文名，唯一性
	Execute(ctx context.Context, args any) (any, error) //需要执行的插件方法
}
