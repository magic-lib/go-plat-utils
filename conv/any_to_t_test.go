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

	testCases := []*utils.TestStruct{
		{"string 1 to int 1", []any{"1"}, []any{1, true}, conv.Convert[int]},
		{"bool to int 0", []any{false}, []any{0, true}, conv.Convert[int]},
		{"int -5 to int64", []any{-5}, []any{int64(-5), true}, conv.Convert[int64]},
		{"string to time", []any{timeStr}, []any{parsedTime, true}, conv.Convert[time.Time]},
	}
	utils.TestFunction(t, testCases, conv.Bool)
}
