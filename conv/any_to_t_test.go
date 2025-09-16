package conv_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
	"time"
)

func TestAnyToAny(t *testing.T) {

	// 目标时间字符串
	timeStr := "2025-9-12 12:04:07"
	// 布局模板（与字符串格式严格匹配）
	layout := "2006-1-2 3:04:05"

	// 解析时间
	parsedTime, err := time.Parse(layout, timeStr)
	if err != nil {
		fmt.Printf("解析失败：%v\n", err)
		panic(err)
	}
	// 布局模板（与字符串格式严格匹配）
	layout1 := "2025/9/5 11:28"
	layout2 := "2025/09/05 11:28"
	layout3 := "2025/9/5 9:13"

	timeStr1 := "2006/1/2 15:04"

	// 解析时间
	parsedTime1, err := time.Parse(timeStr1, layout1)
	if err != nil {
		fmt.Printf("解析失败：%v\n", err)
		panic(err)
	}
	parsedTime2, err := time.Parse(timeStr1, layout2)
	if err != nil {
		fmt.Printf("解析失败：%v\n", err)
		panic(err)
	}
	parsedTime3, err := time.Parse(timeStr1, layout3)
	if err != nil {
		fmt.Printf("解析失败：%v\n", err)
		panic(err)
	}

	testCases := []*utils.TestStruct{
		{"string 1 to int 1", []any{"1"}, []any{1, true}, conv.Convert[int]},
		{"bool to int 0", []any{false}, []any{0, true}, conv.Convert[int]},
		{"int -5 to int64", []any{-5}, []any{int64(-5), true}, conv.Convert[int64]},
		{"string to time", []any{timeStr}, []any{parsedTime, true}, conv.Convert[time.Time]},
		{"string to time " + layout1, []any{layout1}, []any{parsedTime1, true}, conv.Convert[time.Time]},
		{"string to time " + layout2, []any{layout2}, []any{parsedTime2, true}, conv.Convert[time.Time]},
		{"string to time " + layout3, []any{layout3}, []any{parsedTime3, true}, conv.Convert[time.Time]},
	}
	utils.TestFunction(t, testCases, conv.Bool)
}
