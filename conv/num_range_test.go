package conv_test

import (
	"math"
	"testing"

	"github.com/magic-lib/go-plat-utils/conv"
)

func checkConvert[T comparable](t *testing.T, name string, in any, want T) {
	t.Helper()
	got, err := conv.Convert[T](in)
	if err != nil || got != want {
		t.Errorf("%s: Convert(%v) = %v, %v; want %v", name, in, got, err, want)
	}
}

// 越界取边界值：超过最大值取最大值，低于最小值取最小值（无符号最小值为 0）
func TestConvertClampOnOverflow(t *testing.T) {
	checkConvert[int8](t, "int8 上界", 300, math.MaxInt8)
	checkConvert[int8](t, "int8 下界", -300, math.MinInt8)
	checkConvert[int16](t, "int16 上界", "70000", math.MaxInt16)
	checkConvert[int32](t, "int32 下界", int64(math.MinInt64), math.MinInt32)
	checkConvert[int32](t, "int32 正常", 123, 123)

	checkConvert[int64](t, "uint64 超出 int64", uint64(math.MaxUint64), math.MaxInt64)
	checkConvert[int64](t, "float 超出 int64", 1e30, math.MaxInt64)
	checkConvert[int64](t, "float 低于 int64", -1e30, math.MinInt64)
	checkConvert[int64](t, "+Inf", math.Inf(1), math.MaxInt64)
	checkConvert[int64](t, "超长数字字符串", "99999999999999999999", math.MaxInt64)
	checkConvert[int64](t, "超长负数字符串", "-99999999999999999999", math.MinInt64)
	checkConvert[int64](t, "整数小数字符串越界", "99999999999999999999.00", math.MaxInt64)
	checkConvert[int](t, "int 上界", uint64(math.MaxUint64), math.MaxInt)

	checkConvert[uint8](t, "uint8 上界", 1000, math.MaxUint8)
	checkConvert[uint8](t, "uint8 负数", -5, 0)
	checkConvert[uint16](t, "uint16 上界", "99999999", math.MaxUint16)
	checkConvert[uint32](t, "uint32 负数字符串", "-5", 0)
	checkConvert[uint32](t, "uint32 带正号", "+7", 7)
	checkConvert[uint64](t, "uint64 负浮点", -1.5, 0)
	checkConvert[uint64](t, "uint64 超大浮点", 1e30, math.MaxUint64)
	checkConvert[uint64](t, "uint64 超长字符串", "999999999999999999999999", math.MaxUint64)
	checkConvert[uint](t, "uint 正常", "42", 42)

	checkConvert[float32](t, "float32 上界", 1e300, math.MaxFloat32)
	checkConvert[float32](t, "float32 下界", -1e300, -math.MaxFloat32)
	checkConvert[float32](t, "float32 +Inf", math.Inf(1), float32(math.Inf(1)))
}

// NaN 无法比较大小，转整数仍返回失败
func TestConvertNaNToInt(t *testing.T) {
	if v, err := conv.Convert[int64](math.NaN()); err == nil {
		t.Errorf("NaN 转 int64 应失败, got %v", v)
	}
	if v, err := conv.Convert[uint32](math.NaN()); err == nil {
		t.Errorf("NaN 转 uint32 应失败, got %v", v)
	}
}

func TestStrToIntClamp(t *testing.T) {
	if v := conv.StrToInt8("300"); v != math.MaxInt8 {
		t.Errorf("StrToInt8(300) = %d", v)
	}
	if v := conv.StrToInt64("-99999999999999999999"); v != math.MinInt64 {
		t.Errorf("StrToInt64 = %d", v)
	}
	if v := conv.StrToUint16("70000"); v != math.MaxUint16 {
		t.Errorf("StrToUint16(70000) = %d", v)
	}
	if v := conv.StrToUint32("-1"); v != 0 {
		t.Errorf("StrToUint32(-1) = %d", v)
	}
	if v := conv.StrToInt32("abc"); v != 0 {
		t.Errorf("StrToInt32(abc) = %d, 非数字仍返回 0", v)
	}
}
