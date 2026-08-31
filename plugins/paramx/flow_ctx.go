// Package paramx 执行流程中方法参数的整个传递
package paramx

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates"
	"log"
	"sync"
	"time"
)

// FlowContext 执行流程中方法参数的整个传递
/*
{
	"arguments" : {
		"aa" : "11",
		"bb" : "22"
	},
	"responses": {
		"resp": 4
	},
	"steps": {
		"func1_params" : {
			"arguments" : {
				"aa" : "11"
			},
			"configs" : {
				"cc" : "dd"
			},
			"responses" : "bb"
		}
	}
}
*/

// StepId 流程ID
type StepId string

// FlowStatus 流程整体状态类型
type FlowStatus string

// StepStatus 单个步骤状态类型
type StepStatus string

// 流程状态常量
const (
	FlowStatusInit    FlowStatus = "init"
	FlowStatusRunning FlowStatus = "running"
	FlowStatusSuccess FlowStatus = "success"
	FlowStatusFail    FlowStatus = "failure"
)

// 步骤状态常量
const (
	StepStatusPending StepStatus = "pending"
	StepStatusSuccess StepStatus = "success"
	StepStatusFail    StepStatus = "failure"
	StepStatusSkip    StepStatus = "skip"
)

type FlowMeta struct {
	FlowId       string         `json:"flow_id"`     // 流程id
	InstanceId   string         `json:"instance_id"` // 实例id
	TraceId      string         `json:"trace_id"`    // 链路追踪ID
	Status       FlowStatus     `json:"status"`      // init/running/success/failure
	Configs      map[string]any `json:"configs,omitempty"`
	CreateTimeMs int64          `json:"create_time_ms,omitempty"`
	StartTimeMs  int64          `json:"start_time_ms,omitempty"`
	EndTimeMs    int64          `json:"end_time_ms,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

type Step struct {
	Arguments   map[string]any `json:"arguments"`
	Responses   any            `json:"responses"`
	Status      StepStatus     `json:"status"` // pending/success/fail/skip
	Error       *ErrorInfo     `json:"error,omitempty"`
	StartTimeMs int64          `json:"start_time_ms"`
	EndTimeMs   int64          `json:"end_time_ms"`
}

type FlowContext struct {
	mux       sync.RWMutex
	Meta      *FlowMeta        `json:"meta"`            //附带的元数据
	Arguments map[string]any   `json:"arguments"`       //全局的参数
	Responses any              `json:"responses"`       //全局的返回
	Steps     map[StepId]*Step `json:"steps,omitempty"` //每一步的配置，key是stepId
}

func NewFlowContext(flowId, instanceId string, globalArguments map[string]any, stepKeyFunc ...func(key string, val any) (any, bool)) *FlowContext {
	// 拷贝一份，避免外部修改传入的 map 影响内部状态
	args := make(map[string]any, len(globalArguments))
	steps := make(map[StepId]*Step)
	for k, v := range globalArguments {
		if len(stepKeyFunc) > 0 {
			if newVal, ok := stepKeyFunc[0](k, v); ok {
				newV := make(map[string]any)
				_ = conv.Unmarshal(newVal, &newV)
				if len(newV) > 0 {
					steps[StepId(k)] = &Step{
						Arguments: newV,
						Status:    StepStatusPending,
					}
					continue
				}
			}
		}
		args[k] = v
	}
	return &FlowContext{
		Meta: &FlowMeta{
			FlowId:       flowId,
			InstanceId:   instanceId,
			Status:       FlowStatusInit,
			CreateTimeMs: time.Now().UnixMilli(),
		},
		Arguments: args,
		Steps:     steps,
	}
}

func (c *FlowContext) SetVariables(vars map[string]any) {
	if len(vars) == 0 {
		return
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.Arguments == nil {
		c.Arguments = make(map[string]any)
	}
	for k, v := range vars {
		c.Arguments[k] = v
	}
}
func (c *FlowContext) SetTraceId(traceId string) {
	if traceId == "" {
		return
	}
	if c.Meta == nil {
		c.Meta = &FlowMeta{}
	}
	c.Meta.TraceId = traceId
}
func (c *FlowContext) SetFlowId(flowId string) {
	if flowId == "" {
		return
	}
	if c.Meta == nil {
		c.Meta = &FlowMeta{}
	}
	c.Meta.FlowId = flowId
}
func (c *FlowContext) SetInstanceId(instanceId string) {
	if instanceId == "" {
		return
	}
	if c.Meta == nil {
		c.Meta = &FlowMeta{}
	}
	c.Meta.InstanceId = instanceId
}

func (c *FlowContext) SetConfig(key string, value any) {
	if c.Meta == nil {
		c.Meta = &FlowMeta{}
	}
	if len(c.Meta.Configs) == 0 {
		c.Meta.Configs = make(map[string]any)
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	c.Meta.Configs[key] = value
}

func (c *FlowContext) GetConfig(key string) (any, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	if c.Meta == nil || len(c.Meta.Configs) == 0 {
		return nil, false
	}
	v, ok := c.Meta.Configs[key]
	return v, ok
}

func (c *FlowContext) SetVariable(key string, value any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.Arguments == nil {
		c.Arguments = make(map[string]any)
	}
	c.Arguments[key] = value
}

func (c *FlowContext) GetVariable(key string) (any, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	v, ok := c.Arguments[key]
	return v, ok
}

func (c *FlowContext) SetResponses(resp any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.Responses = resp
}
func (c *FlowContext) GetResponses() any {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.Responses
}

func (c *FlowContext) SetStep(stepId StepId, oneStep *Step) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.Steps == nil {
		c.Steps = make(map[StepId]*Step)
	}
	if oneStep != nil {
		c.Steps[stepId] = oneStep
	}
}
func (c *FlowContext) GetStep(stepId StepId) *Step {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.Steps == nil {
		c.Steps = make(map[StepId]*Step)
	}
	if oneStep, ok := c.Steps[stepId]; ok {
		return oneStep
	}
	return nil
}
func (c *FlowContext) SetStepError(stepId StepId, err *ErrorInfo) {
	if err == nil {
		return
	}

	c.mux.Lock()
	defer c.mux.Unlock()
	if c.Steps == nil {
		c.Steps = make(map[StepId]*Step)
	}
	ps, ok := c.Steps[stepId]
	if !ok {
		ps = &Step{}
	}
	ps.Error = err
}

// SetStepArguments 记录方法 stepId 本次执行"实际拿到的入参"。
// 调用方应只塞该方法真正需要的字段（数量/内容可与全局 Variables 不同）。
func (c *FlowContext) SetStepArguments(stepId StepId, params map[string]any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.Steps == nil {
		c.Steps = make(map[StepId]*Step)
	}
	ps, ok := c.Steps[stepId]
	if !ok {
		ps = &Step{
			Arguments: make(map[string]any),
			Status:    StepStatusPending,
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
func (c *FlowContext) GetStepArguments(stepId StepId) map[string]any {
	c.mux.RLock()
	oneStepMap := make(map[string]any)
	for k, v := range c.Arguments {
		oneStepMap[k] = v
	}
	if ps, ok := c.Steps[stepId]; ok && ps != nil && len(ps.Arguments) > 0 {
		for k, v := range ps.Arguments {
			oneStepMap[k] = v
		}
	}
	c.mux.RUnlock()
	newMap, err := c.TemplateArguments(oneStepMap)
	if err != nil {
		log.Print("StepArgumentMap TemplateArguments err:", err.Error())
		return oneStepMap
	}
	return newMap
}

// SetStepResponse 记录方法 stepId 的返回值
func (c *FlowContext) SetStepResponse(stepId StepId, resp any) {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.Steps == nil {
		c.Steps = make(map[StepId]*Step)
	}
	ps, ok := c.Steps[stepId]
	if !ok || ps == nil {
		ps = &Step{
			Status: StepStatusPending,
		}
	}
	ps.Responses = resp
	c.Steps[stepId] = ps
}

// GetStepResponse 取方法 stepId 的返回值
func (c *FlowContext) GetStepResponse(stepId StepId) (any, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	ps, ok := c.Steps[stepId]
	if !ok || ps == nil {
		return nil, false
	}
	return ps.Responses, true
}
func (c *FlowContext) StepMaps(stepId StepId) (map[string]any, bool) {
	c.mux.RLock()
	defer c.mux.RUnlock()
	ps, ok := c.Steps[stepId]
	if !ok || ps == nil {
		return nil, false
	}
	stepMap, err := ps.toMaps()
	if err != nil {
		log.Print("GetStepMap ToMaps err:", err.Error())
		return stepMap, false
	}
	return stepMap, true
}
func (s *Step) toMaps() (map[string]any, error) {
	newMap := make(map[string]any)
	err := conv.Unmarshal(s, &newMap)
	if err != nil {
		log.Print("ToMaps copy c err:", err.Error())
		return newMap, err
	}
	return newMap, nil
}

// ToMaps 将参数转换为map
func (c *FlowContext) ToMaps() (map[string]any, error) {
	newMap := make(map[string]any)
	// 深拷贝一个 c，避免并发错误（*c 浅拷贝仍会共享 Steps/Arguments 等内部 map）
	newC := new(FlowContext)
	c.mux.RLock()
	err := conv.Unmarshal(c, newC)
	c.mux.RUnlock()
	if err != nil {
		log.Print("ToMaps copy c err:", err.Error())
		return newMap, err
	}
	err = conv.Unmarshal(newC, &newMap)
	if err != nil {
		log.Print("ToMaps conv.Unmarshal err:", err.Error())
		return newMap, err
	}
	if len(newMap) == 0 {
		return newMap, fmt.Errorf("ToMaps newMap is empty")
	}
	return newMap, nil
}

// TemplateArguments 替换变量
func (c *FlowContext) TemplateArguments(argsTemplate map[string]any) (map[string]any, error) {
	argsStr := conv.String(argsTemplate)
	ruleExpr := templates.NewRuleExprEngine()
	allMaps, _ := c.ToMaps()
	retResult, err := ruleExpr.RenderObject(argsStr, allMaps)
	if err != nil {
		return argsTemplate, err
	}
	if retMap, ok := retResult.(map[string]any); ok {
		return retMap, nil
	}
	return argsTemplate, fmt.Errorf("TemplateArguments retResult is not map[string]any")
}
