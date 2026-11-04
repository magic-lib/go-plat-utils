package conv_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

// map 转字符串输出紧凑 JSON 且按 key 排序，结果稳定
func TestStringMapSortedCompact(t *testing.T) {
	for i := 0; i < 20; i++ { // map 遍历顺序随机，多跑几次
		if got, want := conv.String(map[string]string{"name": "zhangsan", "age": "55", "b": "x"}),
			`{"age":"55","b":"x","name":"zhangsan"}`; got != want {
			t.Fatalf("map[string]string = %q, want %q", got, want)
		}
		// 含 int 等 jcs 不支持的类型时走兜底分支，也要按 key 排序
		if got, want := conv.String(map[string]any{"name": "a", "age": 55, "b": int64(1)}),
			`{"age":55,"b":1,"name":"a"}`; got != want {
			t.Fatalf("map[string]any = %q, want %q", got, want)
		}
	}
}
