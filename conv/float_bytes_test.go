package conv_test

import (
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

func TestFloatConvertEdge(t *testing.T) {
	cases := []struct {
		in   any
		want float64
	}{
		{float32(0.1), 0.1},
		{float32(1e-7), 1e-7},
		{" 3.5 ", 3.5},
		{"1234567890.123456", 1234567890.123456},
	}
	for _, c := range cases {
		if got, err := conv.Convert[float64](c.in); err != nil || got != c.want {
			t.Errorf("Convert[float64](%v) = %v, %v; want %v", c.in, got, err, c.want)
		}
	}
}

// 长度为 8 的 []byte 也按数字文本解析
func TestInt64FromBytes(t *testing.T) {
	for in, want := range map[string]int64{"20240101": 20240101, "12": 12, "123456789": 123456789} {
		if got, err := conv.Convert[int64]([]byte(in)); err != nil || got != want {
			t.Errorf("Convert[int64]([]byte(%q)) = %v, %v; want %d", in, got, err, want)
		}
	}
}
