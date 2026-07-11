// Package paramx 执行流程中方法参数的整个传递
package paramx

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
	Variables    map[string]any         `json:"variables"`
	Configs      map[string]any         `json:"configs"`
	Responses    any                    `json:"responses"`
	AllFuncParam map[string]ParamStruct `json:"all_func_param"`
}

func NewParamCtx() *ParamCtx {
	return &ParamCtx{
		Variables:    make(map[string]any),
		Configs:      make(map[string]any),
		Responses:    nil,
		AllFuncParam: make(map[string]ParamStruct),
	}
}

func NewParamCtxFromVariables(vars map[string]any) *ParamCtx {
	c := NewParamCtx()
	for k, v := range vars {
		c.Variables[k] = v
	}
	return c
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

// SetFuncParam 记录方法 funcId 本次执行"实际拿到的入参"。
// 调用方应只塞该方法真正需要的字段（数量/内容可与全局 Variables 不同）。
func (c *ParamCtx) SetFuncParam(funcId string, params map[string]any) {
	if c.AllFuncParam == nil {
		c.AllFuncParam = make(map[string]ParamStruct)
	}
	ps, ok := c.AllFuncParam[funcId]
	if !ok || ps.Arguments == nil {
		ps.Arguments = make(map[string]any)
	}
	for k, v := range params {
		ps.Arguments[k] = v
	}
	c.AllFuncParam[funcId] = ps
}

// GetFuncParam 取方法 funcId 实际入参快照（不存在返回 nil）
func (c *ParamCtx) GetFuncParam(funcId string) map[string]any {
	if ps, ok := c.AllFuncParam[funcId]; ok {
		return ps.Arguments
	}
	return nil
}

// SetFuncResponse 记录方法 funcId 的返回值
func (c *ParamCtx) SetFuncResponse(funcId string, resp any) {
	if c.AllFuncParam == nil {
		c.AllFuncParam = make(map[string]ParamStruct)
	}
	ps := c.AllFuncParam[funcId]
	ps.Responses = resp
	c.AllFuncParam[funcId] = ps
}

// GetFuncResponse 取方法 funcId 的返回值
func (c *ParamCtx) GetFuncResponse(funcId string) (any, bool) {
	ps, ok := c.AllFuncParam[funcId]
	return ps.Responses, ok
}

// ToAllMap 将参数转换为map
func (c *ParamCtx) ToAllMap() map[string]any {
	newMap := map[string]any{
		"variables": c.Variables,
		"configs":   c.Configs,
		"responses": c.Responses,
	}
	for k, v := range c.AllFuncParam {
		newMap[k] = v
	}
	return newMap
}

// ToOneMap 将参数转换为map
func (c *ParamCtx) ToOneMap(funcId string) map[string]any {
	oneFuncMap := make(map[string]any)
	for k, v := range c.Variables {
		oneFuncMap[k] = v
	}
	if ps, ok := c.AllFuncParam[funcId]; ok {
		for k, v := range ps.Arguments {
			oneFuncMap[k] = v
		}
	}
	return oneFuncMap
}

// MergeFuncResponseToVariables 仅当上游输出需作为下游入参时调用：
// 把 funcId 的返回值（若为 map）合并进全局 Variables，供后续方法读取。
func (c *ParamCtx) MergeFuncResponseToVariables(funcId string, keyList ...string) {
	resp, ok := c.GetFuncResponse(funcId)
	if !ok {
		return
	}
	if m, ok := resp.(map[string]any); ok {
		if len(keyList) > 0 {
			for _, k := range keyList {
				if v, ok := m[k]; ok {
					c.SetVariable(k, v)
				}
			}
		} else {
			for k, v := range m {
				c.SetVariable(k, v)
			}
		}
	}
}
