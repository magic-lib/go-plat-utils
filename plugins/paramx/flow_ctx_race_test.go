package paramx_test

import (
	"fmt"
	"sync"
	"testing"

	"github.com/magic-lib/go-plat-utils/plugins/paramx"
)

// 步骤不存在时设置错误，需新建步骤并保留错误信息
func TestSetStepErrorNewStep(t *testing.T) {
	c := paramx.NewFlowContext("f1", "i1", nil)
	c.SetStepError("s1", &paramx.ErrorInfo{})
	if step := c.GetStep("s1"); step == nil || step.Error == nil {
		t.Fatalf("SetStepError 后应能取到步骤及错误, got %+v", step)
	}
}

// 并发写元信息与配置，配合 -race 检查数据竞争
func TestFlowContextConcurrentMeta(t *testing.T) {
	c := paramx.NewFlowContext("f1", "i1", nil)
	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.SetTraceId(fmt.Sprint("t", i))
			c.SetFlowId(fmt.Sprint("f", i))
			c.SetInstanceId(fmt.Sprint("in", i))
			c.SetConfig(fmt.Sprint("k", i), i)
			_, _ = c.GetConfig("k0")
		}(i)
	}
	wg.Wait()
	for i := 0; i < 20; i++ {
		if v, ok := c.GetConfig(fmt.Sprint("k", i)); !ok || v != i {
			t.Fatalf("配置 k%d 丢失: %v %v", i, v, ok)
		}
	}
}
