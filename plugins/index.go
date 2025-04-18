package plugins

import (
	"context"
)

type Plugin interface {
	Name() string                                       //插件的英文名，唯一性
	MustArgList() []string                              //必须的参数列表
	Execute(ctx context.Context, args any) (any, error) //需要执行的插件方法
}
