package utils

import (
	"context"
	"testing"
)

// TResp 为接口时返回 (nil, nil) 是合法结果
func TestContextMethodToAnyHandlerNilResp(t *testing.T) {
	h, err := ContextMethodToAnyHandler[int, any](func(ctx context.Context, a int) (any, error) {
		return nil, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	ret, err := h(context.Background(), 1)
	if err != nil || ret != nil {
		t.Fatalf("got (%v, %v), want (nil, nil)", ret, err)
	}
}

// 整数参数转字符串应得到数字文本，而不是码点字符
func TestFuncArgListIntToString(t *testing.T) {
	list, err := FuncArgList(func(s string) {}, 65)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0] != "65" {
		t.Fatalf("got %#v, want [\"65\"]", list)
	}
}

// 切片转数组长度不足时不能 panic
func TestFuncArgListSliceToArray(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("FuncArgList panic: %v", r)
		}
	}()
	_, _ = FuncArgList(func(a [3]int) {}, []int{1})
}

func TestSliceDiffNilElement(t *testing.T) {
	diff := SliceDiff([]any{nil, 1, "a"}, []any{1})
	if len(diff) != 2 || diff[0] != nil || diff[1] != "a" {
		t.Fatalf("got %#v", diff)
	}
}

func TestGetStructInfoNil(t *testing.T) {
	info := GetStructInfo(nil)
	if info.TypeName != "" || len(info.Fields) != 0 {
		t.Fatalf("got %#v", info)
	}
}
