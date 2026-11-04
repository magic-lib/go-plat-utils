package conv

import (
	"fmt"
	"testing"
)

// T 为接口类型时不能 panic
func TestConvertInterfaceTarget(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Convert[fmt.Stringer] panic: %v", r)
		}
	}()
	_, _ = Convert[fmt.Stringer](1)
}

// 解引用后为 nil 时返回零值指针，不能 panic
func TestPointerDerefNil(t *testing.T) {
	type item struct{ A int }
	var x any
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Pointer panic: %v", r)
		}
	}()
	p := Pointer[item](&x)
	if p == nil || p.A != 0 {
		t.Fatalf("Pointer = %v, want zero value pointer", p)
	}
}

// dstPoint 为 nil 时返回错误，不能 panic
func TestAssignToNilDst(t *testing.T) {
	if err := AssignTo(map[string]any{"a": 1}, nil); err == nil {
		t.Fatal("AssignTo(nil dst) should return error")
	}
}
