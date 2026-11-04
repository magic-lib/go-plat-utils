package conv_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

func TestConvertSlice(t *testing.T) {
	// 含 nil 元素
	got, err := conv.Convert[[]any]([]any{nil, 1, "a"})
	if err != nil || len(got) != 3 || got[0] != nil || got[1] != 1 || got[2] != "a" {
		t.Fatalf("Convert[[]any]([]any{nil,1,a}) = %v, %v", got, err)
	}
	ptrs, err := conv.Convert[[]*int]([]any{nil})
	if err != nil || len(ptrs) != 1 || ptrs[0] != nil {
		t.Fatalf("Convert[[]*int]([]any{nil}) = %v, %v", ptrs, err)
	}
	// 已是目标类型直接返回
	bs := []byte("hello")
	if got, err := conv.Convert[[]byte](bs); err != nil || string(got) != "hello" {
		t.Fatalf("Convert[[]byte] = %q, %v", got, err)
	}
	// 跨类型仍逐个转换
	if got, err := conv.Convert[[]int]([]string{"1", "2"}); err != nil || len(got) != 2 || got[1] != 2 {
		t.Fatalf("Convert[[]int]([]string) = %v, %v", got, err)
	}
}
