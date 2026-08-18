package paramx_test

import (
	"sync"
	"testing"

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
