package cond_test

import (
	"math"
	"testing"

	"github.com/magic-lib/go-plat-utils/cond"
)

func TestIsEqualLargeInteger(t *testing.T) {
	testCases := []struct {
		name string
		a, b any
		want bool
	}{
		{"超过 2^53 的不同 ID", int64(9007199254740993), int64(9007199254740992), false},
		{"不同类型的大整数", int64(9007199254740993), uint64(9007199254740992), false},
		{"大整数字符串与整数", "9007199254740993", int64(9007199254740992), false},
		{"大整数字符串与整数相等", "9007199254740993", int64(9007199254740993), true},
		{"int64 与 uint64 最大值", int64(math.MaxInt64), uint64(math.MaxInt64), true},
		{"负数与 uint", int64(-1), uint64(math.MaxUint64), false},
		// 原有宽松相等语义保持不变
		{"字符串与整数", "1", 1, true},
		{"带空格字符串", " 7 ", int32(7), true},
		{"小数字符串与整数", "1.0", 1, true},
		{"整数与浮点", int64(1), 1.0, true},
		{"浮点误差", 1.0, float32(1.0000000001), true},
		{"不等", "2", 1, false},
	}
	for _, tc := range testCases {
		if got := cond.IsEqual(tc.a, tc.b); got != tc.want {
			t.Errorf("%s: IsEqual(%v, %v) = %v, want %v", tc.name, tc.a, tc.b, got, tc.want)
		}
	}
}
