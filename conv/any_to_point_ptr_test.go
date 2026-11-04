package conv_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

func TestPointerFromOtherPointer(t *testing.T) {
	i64 := int64(5)
	if p := conv.Pointer[int](&i64); p == nil || *p != 5 {
		t.Fatalf("Pointer[int](&int64) = %v", p)
	}
	str := "7"
	if p := conv.Pointer[int64](&str); p == nil || *p != 7 {
		t.Fatalf("Pointer[int64](&string) = %v", p)
	}
	var nilPtr *int64
	if p := conv.Pointer[int](nilPtr); p != nil {
		t.Fatalf("Pointer[int](nil *int64) = %v, want nil", p)
	}
	n := 3
	if p := conv.Pointer[int](&n); p != &n {
		t.Fatal("同类型指针应原样返回")
	}
}
