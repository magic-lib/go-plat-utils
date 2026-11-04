package conv_test

import (
	"sync"
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

// map[any]any 及混合 key 的 sync.Map 转字符串结果稳定且按 key 排序
func TestStringAnyMapStable(t *testing.T) {
	for i := 0; i < 30; i++ {
		if got, want := conv.String(map[any]any{"b": 1, "a": 2, 3: "c", "d": int64(4)}),
			`{"3":"c","a":2,"b":1,"d":4}`; got != want {
			t.Fatalf("map[any]any = %s, want %s", got, want)
		}
		var m sync.Map
		m.Store("b", 1)
		m.Store(2, "x")
		if got, want := conv.String(&m), `{"2":"x","b":1}`; got != want {
			t.Fatalf("sync.Map = %s, want %s", got, want)
		}
	}
}
