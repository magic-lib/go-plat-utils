package templates

import (
	"github.com/dop251/goja"
	"github.com/magic-lib/go-plat-utils/conv"
)

type Goja struct {
	rt *goja.Runtime
}

// GojaConfig Goja 执行配置
type GojaConfig struct {
	// ArgsName 把 RunString 传入的整个 args 对象注入 JS 环境时所用的变量名。
	// 例如设置为 "args"，脚本里即可用 args.xxx 访问该对象；
	// 为空字符串表示不注入整个对象（此时 RunString 的 args 参数会被忽略）。
	ArgsName string
}

func NewGoja() *Goja {
	return &Goja{rt: goja.New()}
}

func (g *Goja) WithFuncsAndVars(funcs map[string]any, vars map[string]any) error {
	for k, v := range funcs {
		err := g.rt.Set(k, v)
		if err != nil {
			return err
		}
	}
	for k, v := range vars {
		err := g.rt.Set(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// RunString 执行 expr 并把结果导出为 Go 值
//   - expr  JS 代码或表达式
//   - args  需要注入 JS 环境的对象，注入后的变量名由 cfg 的 ArgsName 指定
//   - cfg   可选配置，多个时只取第一个非 nil 的
//
// 注意：goja.Runtime 不是并发安全的，同一个 Goja 实例不可并发调用
func (g *Goja) RunString(expr string, args any, cfg ...*GojaConfig) (any, error) {
	var conf *GojaConfig
	if len(cfg) > 0 {
		conf = cfg[0]
	}

	// 将map映射上，这样就可以直接使用{"name":1,"age":2} name 和 age
	argMaps := make(map[string]any)
	_ = conv.Unmarshal(args, &argMaps)
	if len(argMaps) > 0 {
		for k, v := range argMaps {
			err := g.rt.Set(k, v)
			if err != nil {
				return nil, err
			}
		}
	}

	// 把整个 args 对象挂到 ArgsName 指定的全局变量上
	if conf != nil && conf.ArgsName != "" && args != nil {
		if err := g.rt.Set(conf.ArgsName, args); err != nil {
			return nil, err
		}
	}

	value, err := g.rt.RunString(expr)
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	return value.Export(), nil
}
