package paramx_test

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/plugins/paramx"
)

// TestGetStepResponseNotExist 验证 stepId 不存在时不会 panic
func TestGetStepResponseNotExist(t *testing.T) {
	c := paramx.NewFlowContext("f1", "i1", map[string]any{"aa": "11"})
	resp, ok := c.GetStepResponse("not_exist")
	if ok {
		t.Errorf("GetStepResponse 应返回 ok=false, got ok=%v resp=%v", ok, resp)
	}
	if resp != nil {
		t.Errorf("GetStepResponse 应返回 nil, got %v", resp)
	}
}

// TestGetStepResponseExist 验证正常写入后可读取
func TestGetStepResponseExist(t *testing.T) {
	c := paramx.NewFlowContext("f1", "i1", nil)
	c.SetStepResponse("step1", "ok")
	resp, ok := c.GetStepResponse("step1")
	if !ok {
		t.Errorf("GetStepResponse 应返回 ok=true")
	}
	if resp != "ok" {
		t.Errorf("GetStepResponse = %v, want ok", resp)
	}
}

// TestStepArguments 验证步骤参数合并（全局 + 步骤级覆盖）
func TestStepArguments(t *testing.T) {
	c := paramx.NewFlowContext("f1", "i1", map[string]any{"aa": "11", "bb": "22"})
	c.SetStepArguments("step1", map[string]any{"bb": "step-bb", "cc": "33"})
	args := c.GetStepArguments("step1")
	if args["aa"] != "11" {
		t.Errorf("args[aa] = %v, want 11", args["aa"])
	}
	if args["bb"] != "step-bb" {
		t.Errorf("args[bb] = %v, want step-bb", args["bb"])
	}
	if args["cc"] != "33" {
		t.Errorf("args[cc] = %v, want 33", args["cc"])
	}
}

// TestNewFlowContextIsolation 验证 NewFlowContext 拷贝外部 map，外部修改不影响内部
func TestNewFlowContextIsolation(t *testing.T) {
	gp := map[string]any{"aa": "11"}
	c := paramx.NewFlowContext("f1", "i1", gp)
	gp["aa"] = "changed"
	v, _ := c.GetVariable("aa")
	if v != "11" {
		t.Errorf("GetVariable(aa) = %v, want 11（外部 map 修改不应影响内部）", v)
	}
}

// TestConcurrentSafety 并发读写验证无数据竞争（go test -race）
func TestConcurrentSafety(t *testing.T) {
	c := paramx.NewFlowContext("f1", "i1", map[string]any{"k": "v"})
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			stepId := paramx.StepId("step")
			_ = n
			c.SetVariable("kk", n)
			c.SetStepArguments(stepId, map[string]any{"x": n})
			c.SetStepResponse(stepId, n)
			_, _ = c.GetVariable("kk")
			_ = c.GetStepArguments(stepId)
			_, _ = c.GetStepResponse(stepId)
			_ = c.GetResponses()
			_, _ = c.GetConfig("cfg")
		}(i)
	}
	wg.Wait()
}

// TestNewFlowContextMeta 验证 NewFlowContext 初始化的元数据与容器
func TestNewFlowContextMeta(t *testing.T) {
	before := time.Now().UnixMilli()
	c := paramx.NewFlowContext("flow-1", "inst-1", nil)
	after := time.Now().UnixMilli()

	if c.Meta == nil {
		t.Fatal("Meta 不应为 nil")
	}
	if c.Meta.FlowId != "flow-1" {
		t.Errorf("Meta.FlowId = %v, want flow-1", c.Meta.FlowId)
	}
	if c.Meta.InstanceId != "inst-1" {
		t.Errorf("Meta.InstanceId = %v, want inst-1", c.Meta.InstanceId)
	}
	if c.Meta.Status != paramx.FlowStatusInit {
		t.Errorf("Meta.Status = %v, want %v", c.Meta.Status, paramx.FlowStatusInit)
	}
	if c.Meta.CreateTimeMs < before || c.Meta.CreateTimeMs > after {
		t.Errorf("Meta.CreateTimeMs = %v, 不在 [%v, %v] 范围内", c.Meta.CreateTimeMs, before, after)
	}
	if c.Meta.TraceId != "" {
		t.Errorf("Meta.TraceId = %v, want 空", c.Meta.TraceId)
	}
	if c.Arguments == nil {
		t.Error("Arguments 不应为 nil")
	}
	if c.Steps == nil {
		t.Error("Steps 不应为 nil")
	}
	if c.Responses != nil {
		t.Errorf("Responses = %v, want nil", c.Responses)
	}
}

// TestNewFlowContextStepKeyFunc 验证 stepKeyFunc 匹配的 key 分流到 Steps、其余进 Arguments
func TestNewFlowContextStepKeyFunc(t *testing.T) {
	stepFn := func(key string) bool { return strings.HasPrefix(key, "step_") }
	c := paramx.NewFlowContext("f1", "i1", map[string]any{
		"step_a": map[string]any{"x": 1},
		"normal": "v",
	}, stepFn)

	if len(c.Steps) != 1 {
		t.Fatalf("Steps 长度 = %d, want 1", len(c.Steps))
	}
	step, ok := c.Steps[paramx.StepId("step_a")]
	if !ok {
		t.Fatal("Steps 中应有 step_a")
	}
	if step.Arguments == nil {
		t.Fatal("step_a.Arguments 不应为 nil")
	}
	if v, ok := step.Arguments["x"].(int); !ok || v != 1 {
		t.Errorf("step_a.Arguments[x] = %v (%T), want int(1)", step.Arguments["x"], step.Arguments["x"])
	}
	// 步骤未执行，状态应为零值 pending
	if step.Status != paramx.StepStatusPending {
		t.Errorf("step_a.Status = %v, want %v", step.Status, paramx.StepStatusPending)
	}
	if v, ok := c.Arguments["normal"]; !ok || v != "v" {
		t.Errorf("Arguments[normal] = (%v, %v), want (v, true)", v, ok)
	}
	if _, ok := c.Arguments["step_a"]; ok {
		t.Error("Arguments 中不应包含 step_a")
	}
}

// TestNewFlowContextStepKeyFuncNotMap 验证 stepKeyFunc 命中但值不是 map 时回退到 Arguments
func TestNewFlowContextStepKeyFuncNotMap(t *testing.T) {
	stepFn := func(key string) bool { return key == "step_b" }
	c := paramx.NewFlowContext("f1", "i1", map[string]any{
		"step_b": "not-a-map",
	}, stepFn)

	if len(c.Steps) != 0 {
		t.Errorf("Steps 长度 = %d, want 0", len(c.Steps))
	}
	if v, ok := c.Arguments["step_b"]; !ok || v != "not-a-map" {
		t.Errorf("Arguments[step_b] = (%v, %v), want (not-a-map, true)", v, ok)
	}
}

// TestNewFlowContextWithoutStepKeyFunc 验证不传 stepKeyFunc 时所有 key 都进 Arguments
func TestNewFlowContextWithoutStepKeyFunc(t *testing.T) {
	c := paramx.NewFlowContext("f1", "i1", map[string]any{"aa": "11", "bb": "22"})

	if len(c.Steps) != 0 {
		t.Errorf("Steps 长度 = %d, want 0", len(c.Steps))
	}
	if len(c.Arguments) != 2 {
		t.Errorf("Arguments 长度 = %d, want 2", len(c.Arguments))
	}
	for k, want := range map[string]any{"aa": "11", "bb": "22"} {
		if v, ok := c.Arguments[k]; !ok || v != want {
			t.Errorf("Arguments[%s] = (%v, %v), want (%v, true)", k, v, ok, want)
		}
	}
}
