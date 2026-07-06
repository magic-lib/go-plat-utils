package activity

import (
	"context"
)

type (
	LifecycleEvent string
	LifecycleHooks map[LifecycleEvent]*Activity
)

const (
	LifecycleEventOnStart    LifecycleEvent = "start"
	LifecycleEventOnComplete LifecycleEvent = "complete"
	LifecycleEventOnSuccess  LifecycleEvent = "success"
	LifecycleEventOnError    LifecycleEvent = "error"
	//LifecycleEventOnTimeout  LifecycleEvent = "timeout"
)

type (
	Executable interface {
		Execute(ctx context.Context, vars map[string]any) (map[string]any, error)
	}
)

type (
	ActivityControl struct {
		When          string `yaml:"when" json:"when,omitempty"`                     // 执行该action的前提条件
		Timeout       int    `yaml:"timeout" json:"timeout,omitempty"`               // 秒，设置超时时间
		CtxCacheable  bool   `yaml:"ctx_cacheable" json:"ctx_cacheable,omitempty"`   // 流程中临时缓存：true=当前流程中会缓存，false=同一流程中不缓存结果
		DelayDuration int    `yaml:"delay_duration" json:"delay_duration,omitempty"` // 延时多长时间执行 ==0 立即执行，> 0 延时执行，<0 异步执行
	}
)

var (
	_ Executable = (*Activity)(nil)
)
