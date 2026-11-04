package cond_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/cond"
)

func TestIsNumericEdge(t *testing.T) {
	cases := map[any]bool{
		"1n": false, "rt5": false, "..5": false, "1.2.3": false,
		"\t1\n": true, " 3.5 ": true, ".5": true, "-2e3": true, "0x1F": true,
		int64(3): true, uint8(1): true, 1.5: true,
	}
	for in, want := range cases {
		if got := cond.IsNumeric(in); got != want {
			t.Errorf("IsNumeric(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestIsZeroComplex(t *testing.T) {
	if !cond.IsZero(complex64(0)) || !cond.IsZero(complex128(0)) || cond.IsZero(complex128(1i)) {
		t.Fatal("IsZero 复数判断错误")
	}
}

func TestIsTimeAnchored(t *testing.T) {
	if !cond.IsTime("2023-01-01 00:00:00") || cond.IsTime("xx2023-01-01 00:00:00") {
		t.Fatal("IsTime 应完整匹配")
	}
	if !cond.IsUUID("123e4567-e89b-12d3-a456-426614174000") || cond.IsUUID("123E4567-e89b-12d3-a456-426614174000") {
		t.Fatal("IsUUID 判断错误")
	}
}
