// Package paramx 执行流程中方法参数的整个传递
package paramx

import (
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates"
	"log"
)

// ParamCtx 执行流程中方法参数的整个传递
/*
{
	"variables" : {
		"aa" : "11",
		"bb" : "22"
	},
	"func1_params" : {
		"arguments" : {
			"aa" : "11"
		},
        "configs" : {
			"cc" : "dd"
		},
		"responses" : "bb"
	},
	"responses": {
		"resp": 4
	}
}
*/

type ParamStruct struct {
	Arguments map[string]any `json:"arguments"`
	Configs   map[string]any `json:"configs,omitempty"`
	Responses any            `json:"responses"`
}

type ParamCtx struct {
	Variables map[string]any         `json:"variables"`
	Configs   map[string]any         `json:"configs"`
	Responses any                    `json:"responses"`
	Steps     map[string]ParamStruct `json:"steps"`
}

func NewParamCtx() *ParamCtx {
	return &ParamCtx{
		Variables: make(map[string]any),
		Configs:   make(map[string]any),
		Responses: nil,
		Steps:     make(map[string]ParamStruct),
	}
}

func NewParamCtxFromVariables(vars map[string]any) *ParamCtx {
	c := NewParamCtx()
	c.SetVariables(vars)
	return c
}

func (c *ParamCtx) SetVariables(vars map[string]any) {
	if len(vars) == 0 {
		return
	}
	if c.Variables == nil {
		c.Variables = make(map[string]any)
	}
	for k, v := range vars {
		c.SetVariable(k, v)
	}
}

func (c *ParamCtx) SetVariable(key string, value any) {
	if c.Variables == nil {
		c.Variables = make(map[string]any)
	}
	c.Variables[key] = value
}

func (c *ParamCtx) GetVariable(key string) (any, bool) {
	v, ok := c.Variables[key]
	return v, ok
}

func (c *ParamCtx) SetConfig(key string, value any) {
	if c.Configs == nil {
		c.Configs = make(map[string]any)
	}
	c.Configs[key] = value
}

func (c *ParamCtx) GetConfig(key string) (any, bool) {
	v, ok := c.Configs[key]
	return v, ok
}

func (c *ParamCtx) SetResponses(resp any) { c.Responses = resp }
func (c *ParamCtx) GetResponses() any     { return c.Responses }

// SetStepArguments 记录方法 stepId 本次执行"实际拿到的入参"。
// 调用方应只塞该方法真正需要的字段（数量/内容可与全局 Variables 不同）。
func (c *ParamCtx) SetStepArguments(stepId string, params map[string]any) {
	if c.Steps == nil {
		c.Steps = make(map[string]ParamStruct)
	}
	ps, ok := c.Steps[stepId]
	if !ok {
		ps = ParamStruct{
			Arguments: make(map[string]any),
		}
	}
	if ps.Arguments == nil {
		ps.Arguments = make(map[string]any)
	}
	for k, v := range params {
		ps.Arguments[k] = v
	}
	c.Steps[stepId] = ps
}

// GetStepArguments 取方法 stepId 实际入参快照（不存在返回 nil）
func (c *ParamCtx) GetStepArguments(stepId string) map[string]any {
	if ps, ok := c.Steps[stepId]; ok {
		return ps.Arguments
	}
	return nil
}

// SetStepResponse 记录方法 stepId 的返回值
func (c *ParamCtx) SetStepResponse(stepId string, resp any) {
	if c.Steps == nil {
		c.Steps = make(map[string]ParamStruct)
	}
	ps := c.Steps[stepId]
	ps.Responses = resp
	c.Steps[stepId] = ps
}

// GetStepResponse 取方法 stepId 的返回值
func (c *ParamCtx) GetStepResponse(stepId string) (any, bool) {
	ps, ok := c.Steps[stepId]
	return ps.Responses, ok
}

// AllMaps 将参数转换为map
func (c *ParamCtx) AllMaps() map[string]any {
	newMap := map[string]any{
		"variables": c.Variables,
		"configs":   c.Configs,
		"responses": c.Responses,
	}
	for k, v := range c.Steps {
		newMap[k] = v
	}
	return newMap
}

// StepMapsByStepId 取方法 stepId 实际入参快照（不存在返回 nil）
func (c *ParamCtx) StepMapsByStepId(stepId string) map[string]any {
	oneStepMap := make(map[string]any)
	for k, v := range c.Variables {
		oneStepMap[k] = v
	}
	args := c.GetStepArguments(stepId)
	if len(args) > 0 {
		for k, v := range args {
			oneStepMap[k] = v
		}
	}
	newMap, err := c.TemplateArguments(oneStepMap)
	if err == nil {
		return newMap
	}
	log.Print("StepArgumentMap TemplateArguments err:", err.Error())
	return oneStepMap
}

// TemplateArguments 替换变量
func (c *ParamCtx) TemplateArguments(args map[string]any) (map[string]any, error) {
	argsStr := conv.String(args)

	ruleExpr := templates.NewRuleExprEngine()
	returnResult, err := ruleExpr.RunString(argsStr, c.AllMaps())
	if err != nil {
		return args, err
	}
	retMap := make(map[string]any)
	err = conv.Unmarshal(returnResult, &retMap)
	if err != nil {
		return args, err
	}
	return retMap, nil
}
