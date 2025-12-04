package conv_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
	"time"
)

func FormatFullTime(t time.Time) string {
	// 自定义布局：覆盖年月日时分秒毫秒 + 时区偏移 + 时区缩写
	const fullLayout = "2006-01-02 15:04:05.000 -07:00 (MST)"
	return t.Format(fullLayout)
}
func TestAnyToTime(t *testing.T) {
	timeFunc := func(val string) bool {
		tt, ok := conv.Time(val)
		fmt.Println("当前完整时间：", val)
		fmt.Println("当前完整时间：", FormatFullTime(tt))
		if ok {
			return ok
		}
		return false
	}

	testCases := []*utils.TestStruct{
		{"2025-01-02T15:04:05Z07:00", []any{"2025-01-02T18:04:05Z07:00"}, []any{true}, timeFunc},
		{"02/01/2006", []any{"02/01/2036"}, []any{true}, timeFunc},
		{"02/1/2006", []any{"02/1/2036"}, []any{true}, timeFunc},
		{"02/1/2006 15:04:05", []any{"02/1/2026 18:04:05"}, []any{true}, timeFunc},
		{time.DateOnly, []any{time.DateOnly}, []any{true}, timeFunc},
		{"2006.01", []any{"2026.01"}, []any{true}, timeFunc},
		{"2006/1/02 15:04:05", []any{"2026/1/02 18:04:05"}, []any{true}, timeFunc},
		{"2006/1/02", []any{"2026/1/02"}, []any{true}, timeFunc},
		{"2006-01-02 15:04:05", []any{"2026-01-02 18:04:05"}, []any{true}, timeFunc},
		{"2006/1/02 15:04", []any{"2026/1/02 18:04"}, []any{true}, timeFunc},
		{"2006/1/2 15:04", []any{"2016/10/20 18:04"}, []any{true}, timeFunc},
		{"02-Jan-2006", []any{"24-Jan-2016"}, []any{true}, timeFunc},
		{"2-Jan-2006", []any{"24-Jan-2026"}, []any{true}, timeFunc},
		{"2006/1/02 15:04:05:00", []any{"2017/10/02 15:04:05:00"}, []any{true}, timeFunc},
		{"Jan 02, 2006", []any{"Jan 25, 2016"}, []any{true}, timeFunc},
		{"2025/12/5", []any{"2025/12/5"}, []any{true}, timeFunc},
		{"2025/9/5 11:28", []any{"2025/9/5 11:28"}, []any{true}, timeFunc},
		{"2025/09/05 11:28", []any{"2025/09/05 11:28"}, []any{true}, timeFunc},
		{"2025/9/5 9:13", []any{"2025/9/5 9:13"}, []any{true}, timeFunc},
		{"2025-9-12 12:04:07", []any{"2025-9-12 12:04:07"}, []any{true}, timeFunc},
		{"2025/12/5", []any{"2025/12/5"}, []any{true}, timeFunc},
	}
	utils.TestFunction(t, testCases, timeFunc)

}

func TestAnyToAny(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"string 1 to int 1", []any{"1"}, []any{1, true}, conv.Convert[int]},
		{"bool to int 0", []any{false}, []any{0, true}, conv.Convert[int]},
		{"int -5 to int64", []any{-5}, []any{int64(-5), true}, conv.Convert[int64]},
	}
	utils.TestFunction(t, testCases, conv.Bool)
}
