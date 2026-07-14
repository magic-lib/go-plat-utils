package activity

import (
	"fmt"
	"time"

	"github.com/magic-lib/go-plat-utils/conv"
)

// ActivityConfig 与 activity.Activity 一一对应，每个属性一列。
// Type 为派生列（act.Type()），作为节点注册 key 与唯一键。
type ActivityConfig struct {
	AutoID       int64
	Type         string
	Id           string
	ActivityType string
	ActNamespace string
	ActName      string
	Arguments    string // []*BindConfig -> JSON
	ArgTemplate  string
	Responses    string // map[string]any -> JSON
	DependsOn    string // []*Activity -> JSON
	Hooks        string // LifecycleHooks -> JSON
	Control      string // ActivityControl -> JSON
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// activityToConfig 摊平为各列。
func activityToConfig(act *Activity) ActivityConfig {
	return ActivityConfig{
		Id:           act.Id,
		ActivityType: act.ActivityType,
		ActNamespace: act.ActNamespace,
		ActName:      act.ActName,
		Arguments:    conv.String(act.Arguments),
		ArgTemplate:  act.ArgTemplate,
		Responses:    conv.String(act.Responses),
		DependsOn:    conv.String(act.DependsOn),
		Hooks:        conv.String(act.Hooks),
		Control:      conv.String(act.Control),
	}
}

// toActivity 从各列还原；空字符串列视为未设置。
func (c ActivityConfig) toActivity() (*Activity, error) {
	act := &Activity{
		Id:           c.Id,
		ActivityType: c.ActivityType,
		ActNamespace: c.ActNamespace,
		ActName:      c.ActName,
		ArgTemplate:  c.ArgTemplate,
	}
	var err error
	if c.Arguments != "" {
		if err = conv.Unmarshal(c.Arguments, &act.Arguments); err != nil {
			return nil, fmt.Errorf("arguments: %w", err)
		}
	}
	if c.Responses != "" {
		if err = conv.Unmarshal(c.Responses, &act.Responses); err != nil {
			return nil, fmt.Errorf("responses: %w", err)
		}
	}
	if c.DependsOn != "" {
		if err = conv.Unmarshal(c.DependsOn, &act.DependsOn); err != nil {
			return nil, fmt.Errorf("depends_on: %w", err)
		}
	}
	if c.Hooks != "" {
		if err = conv.Unmarshal(c.Hooks, &act.Hooks); err != nil {
			return nil, fmt.Errorf("hooks: %w", err)
		}
	}
	if c.Control != "" {
		if err = conv.Unmarshal(c.Control, &act.Control); err != nil {
			return nil, fmt.Errorf("control: %w", err)
		}
	}
	return act, nil
}
