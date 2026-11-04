package conv_test

import (
	"strings"
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

type lazyErrT struct {
	Name string `json:"name"`
}

// 转换失败时错误信息仍包含目标类型和原值；成功时不再为错误信息做序列化
func TestConvertForTypeLazyError(t *testing.T) {
	_, err := conv.Convert[int64](lazyErrT{Name: "abc"})
	if err == nil || !strings.Contains(err.Error(), "int64") || !strings.Contains(err.Error(), `"name":"abc"`) {
		t.Fatalf("错误信息应包含目标类型和原值, got %v", err)
	}

	src := map[string]any{"name": "abc"}
	allocs := testing.AllocsPerRun(50, func() {
		_, _ = conv.Convert[lazyErrT](src)
	})
	t.Logf("Convert[struct](map) 每次分配 %.0f 次", allocs)
}

func BenchmarkConvertMapToStruct(b *testing.B) {
	src := map[string]any{"name": "abc"}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = conv.Convert[lazyErrT](src)
	}
}
