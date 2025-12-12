package conv_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
	"time"
)

func TestAnyToBool(t *testing.T) {
	iPtr := 90
	testCases := []*utils.TestStruct{
		{"bool true", []any{true}, []any{true}, nil},
		{"bool false", []any{false}, []any{false}, nil},
		{"int -1", []any{int(-1)}, []any{true}, nil},
		{"int 1", []any{int(1)}, []any{true}, nil},
		{"int 0", []any{int(0)}, []any{false}, nil},
		{"int8 1", []any{int8(1)}, []any{true}, nil},
		{"int8 0", []any{int8(0)}, []any{false}, nil},
		{"int16 1", []any{int16(1)}, []any{true}, nil},
		{"int16 0", []any{int16(0)}, []any{false}, nil},
		{"int32 1", []any{int32(1)}, []any{true}, nil},
		{"int32 0", []any{int32(0)}, []any{false}, nil},
		{"int64 1", []any{int64(1)}, []any{true}, nil},
		{"int64 100", []any{int64(100)}, []any{true}, nil},
		{"int64 0", []any{int64(0)}, []any{false}, nil},
		{"uint 1", []any{uint(1)}, []any{true}, nil},
		{"uint 0", []any{uint(0)}, []any{false}, nil},
		{"uint8 1", []any{uint8(1)}, []any{true}, nil},
		{"uint8 0", []any{uint8(0)}, []any{false}, nil},
		{"uint16 1", []any{uint16(1)}, []any{true}, nil},
		{"uint16 0", []any{uint16(0)}, []any{false}, nil},
		{"uint32 1", []any{uint32(1)}, []any{true}, nil},
		{"uint32 0", []any{uint32(0)}, []any{false}, nil},
		{"uint64 1", []any{uint64(1)}, []any{true}, nil},
		{"uint64 0", []any{uint64(0)}, []any{false}, nil},
		{"float32 1.0", []any{float32(1.0)}, []any{true}, nil},
		{"float32 0.0", []any{float32(0.0)}, []any{false}, nil},
		{"float64 1.0", []any{float64(1.0)}, []any{true}, nil},
		{"float64 0.0", []any{float64(0.0)}, []any{false}, nil},
		{"string abc", []any{"abc"}, []any{false}, nil},
		{"string true", []any{"true"}, []any{true}, nil},
		{"string false", []any{"false"}, []any{false}, nil},
		{"empty string", []any{""}, []any{false}, nil},
		{"nil value", []any{nil}, []any{false}, nil},
		{"complex64 1+1i", []any{complex64(1 + 1i)}, []any{true}, nil},
		{"complex64 0+0i", []any{complex64(0 + 0i)}, []any{false}, nil},
		{"complex128 1+1i", []any{complex128(1 + 1i)}, []any{true}, nil},
		{"complex128 0+0i", []any{complex128(0 + 0i)}, []any{false}, nil},
		{"nil pointer", []any{(*int)(nil)}, []any{false}, nil},
		{"non-nil pointer", []any{&iPtr}, []any{true}, nil},
		{"empty slice no value", []any{[]int{}}, []any{false}, nil},
		{"empty slice has value", []any{[]int{1, 2, 3}}, []any{true}, nil},
	}
	utils.TestFunction(t, testCases, conv.Convert[bool])
}

func FormatFullTime(t time.Time) string {
	// 自定义布局：覆盖年月日时分秒毫秒 + 时区偏移 + 时区缩写
	const fullLayout = "2006-01-02 15:04:05.000 -07:00 (MST)"
	return t.Format(fullLayout)
}
func TestAnyToTime(t *testing.T) {
	timeFunc := func(val string) bool {
		tt, err := conv.Convert[time.Time](val)
		fmt.Println("当前完整时间：", val)
		fmt.Println("当前完整时间：", FormatFullTime(tt))
		if err == nil {
			return true
		}
		return false
	}

	testCases := []*utils.TestStruct{
		{"2025-01-02T15:04:05Z07:00", []any{"2025-01-02T18:04:05Z07:00"}, []any{true}, nil},
		{"02/01/2006", []any{"02/01/2036"}, []any{true}, nil},
		{"02/1/2006", []any{"02/1/2036"}, []any{true}, nil},
		{"02/1/2006 15:04:05", []any{"02/1/2026 18:04:05"}, []any{true}, nil},
		{time.DateOnly, []any{time.DateOnly}, []any{true}, nil},
		{"2006.01", []any{"2026.01"}, []any{true}, nil},
		{"2006/1/02 15:04:05", []any{"2026/1/02 18:04:05"}, []any{true}, nil},
		{"2006/1/02", []any{"2026/1/02"}, []any{true}, nil},
		{"2006-01-02 15:04:05", []any{"2026-01-02 18:04:05"}, []any{true}, nil},
		{"2006/1/02 15:04", []any{"2026/1/02 18:04"}, []any{true}, nil},
		{"2006/1/2 15:04", []any{"2016/10/20 18:04"}, []any{true}, nil},
		{"02-Jan-2006", []any{"24-Jan-2016"}, []any{true}, nil},
		{"2-Jan-2006", []any{"24-Jan-2026"}, []any{true}, nil},
		{"2006/1/02 15:04:05:00", []any{"2017/10/02 15:04:05:00"}, []any{true}, nil},
		{"Jan 02, 2006", []any{"Jan 25, 2016"}, []any{true}, nil},
		{"2025/12/5", []any{"2025/12/5"}, []any{true}, nil},
		{"2025/9/5 11:28", []any{"2025/9/5 11:28"}, []any{true}, nil},
		{"2025/09/05 11:28", []any{"2025/09/05 11:28"}, []any{true}, nil},
		{"2025/9/5 9:13", []any{"2025/9/5 9:13"}, []any{true}, nil},
		{"2025-9-12 12:04:07", []any{"2025-9-12 12:04:07"}, []any{true}, nil},
		{"2025/12/5", []any{"2025/12/5"}, []any{true}, nil},
		{"2025-12", []any{"2025-12"}, []any{true}, nil},
	}
	utils.TestFunction(t, testCases, timeFunc)
}

func TestAnyToNumber(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"string 1 to int 1", []any{"1"}, []any{1}, conv.Convert[int]},
		{"bool to int 0", []any{false}, []any{0}, conv.Convert[int]},
		{"int 0 to int8 0", []any{0}, []any{int8(0)}, conv.Convert[int8]},
		{"int 0 to int16 0", []any{0}, []any{int16(0)}, conv.Convert[int16]},
		{"int 0 to int32 0", []any{0}, []any{int32(0)}, conv.Convert[int32]},
		{"int -5 to int64", []any{-5}, []any{int64(-5)}, conv.Convert[int64]},
		{"int 5 to uint", []any{5}, []any{uint(5)}, conv.Convert[uint]},
		{"int 5 to uint8", []any{5}, []any{uint8(5)}, conv.Convert[uint8]},
		{"int 5 to uint16", []any{5}, []any{uint16(5)}, conv.Convert[uint16]},
		{"int 5 to uint32", []any{5}, []any{uint32(5)}, conv.Convert[uint32]},
		{"int 5 to uint64", []any{5}, []any{uint64(5)}, conv.Convert[uint64]},
		{"int -5 to float32", []any{-5}, []any{float32(-5)}, conv.Convert[float32]},
		{"int -5 to float64", []any{-5}, []any{float64(-5)}, conv.Convert[float64]},
		{"int -5 to string", []any{-5}, []any{"-5"}, conv.Convert[string]},
	}
	utils.TestFunction(t, testCases, nil)
}
