package ruleengine_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/templates"
	"testing"
	"time"

	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/templates/ruleengine"
)

// baseTime 用例里统一使用的时间，时间戳类断言都由它算出来，避免受时区影响
var baseTime = time.Date(2024, 1, 2, 3, 4, 5, 0, time.Local)

// shiftUnix 在 baseTime 基础上按 年/时/天 平移后取秒级时间戳，避免硬编码时区相关的值
func shiftUnix(years, hours, days int) int64 {
	return baseTime.AddDate(years, 0, days).Add(time.Duration(hours) * time.Hour).Unix()
}

// dateFuncArgs 所有时间用例共用的变量表
func dateFuncArgs() map[string]any {
	return map[string]any{
		"t1": "2024-01-02 03:04:05",
		"t2": "2024-03-04 05:06:07",
		"ts": baseTime.Unix(),
	}
}

func TestDateFunction(t *testing.T) {
	testCases := []struct {
		name  string
		rule  string
		want  any
		isErr bool
	}{
		// DateFormat
		{"格式化-默认返回秒时间戳", `DateFormat('2024-01-02 03:04:05')`, baseTime.Unix(), false},
		{"格式化-默认与 unix 一致", `DateFormat(t1) == DateFormat(t1, 'unix')`, true, false},
		{"格式化-自定义", `DateFormat('2024-01-02 03:04:05', '2006/01/02')`, "2024/01/02", false},
		{"格式化-变量", `DateFormat(t1, '2006年01月02日')`, "2024年01月02日", false},
		{"格式化-时间戳入参", `DateFormat(ts, '2006-01-02 15:04:05')`, "2024-01-02 03:04:05", false},
		{"格式化-unix", `DateFormat('2024-01-02 03:04:05', 'unix')`, baseTime.Unix(), false},
		{"格式化-unixmilli", `DateFormat('2024-01-02 03:04:05', 'unixmilli')`, baseTime.UnixMilli(), false},
		{"格式化-非法时间", `DateFormat('abc')`, nil, true},

		// DateAdd（默认返回秒级时间戳）
		{"加-默认按天返回秒时间戳", `DateAdd('2024-01-02 03:04:05', 1)`, shiftUnix(0, 0, 1), false},
		{"加-默认与 unix 一致", `DateAdd(t1, 1) == DateAdd(t1, 1, 'day', 'unix')`, true, false},
		{"加-默认按天可用差值验证", `DateDiff(DateAdd(t1, 1), t1, 'day')`, 1.0, false},
		{"加-指定小时", `DateAdd('2024-01-02 03:04:05', 2, 'hour')`, shiftUnix(0, 2, 0), false},
		{"加-指定月份", `DateAdd('2024-01-02 03:04:05', 1, 'month', '2006-01-02')`, "2024-02-02", false},
		{"加-跨月进位", `DateAdd('2024-01-31 03:04:05', 1, 'month', '2006-01-02')`, "2024-03-02", false},
		{"加-指定年份", `DateAdd('2024-01-02 03:04:05', 1, 'year', '2006-01-02')`, "2025-01-02", false},
		{"加-指定周", `DateAdd('2024-01-02 03:04:05', 1, 'week', '2006-01-02')`, "2024-01-09", false},
		{"加-负数等于减", `DateAdd('2024-01-02 03:04:05', -1, 'day', '2006-01-02')`, "2024-01-01", false},
		{"加-指定输出格式", `DateAdd('2024-01-02 03:04:05', 1, 'day', '2006/01/02')`, "2024/01/03", false},
		{"加-可直接参与算术", `DateAdd(t1, 1, 'day')-DateFormat(t1)`, 86400, false},
		{"加-单位非法", `DateAdd('2024-01-02 03:04:05', 1, 'xx')`, nil, true},

		// DateDiff
		{"差-默认按秒", `DateDiff('2024-01-02 03:04:06', '2024-01-02 03:04:05')`, 1, false},
		{"差-指定小时", `DateDiff('2024-01-02 05:04:05', '2024-01-02 03:04:05', 'hour')`, 2, false},
		{"差-不足一小时为小数", `DateDiff('2024-01-02 04:04:05', '2024-01-02 03:04:05', 'day')`, 1.0 / 24, false},
		{"差-负数表示前者更早", `DateDiff('2024-01-01 03:04:05', '2024-01-02 03:04:05')`, -86400, false},
		{"差-按天", `DateDiff('2024-01-05 03:04:05', '2024-01-02 03:04:05', 'day')`, 3, false},
		{"差-按周", `DateDiff('2024-01-16 03:04:05', '2024-01-02 03:04:05', 'week')`, 2, false},
		{"差-按月", `DateDiff('2024-03-04 05:06:07', '2024-01-02 03:04:05', 'month')`, 2, false},
		{"差-不足一月不计", `DateDiff('2024-03-01 03:04:05', '2024-01-02 03:04:05', 'month')`, 1, false},
		{"差-按年", `DateDiff('2026-01-02 03:04:05', '2024-01-02 03:04:05', 'year')`, 2, false},

		// DateIsAfter / DateIsBefore
		{"晚于-成立", `DateIsAfter('2024-03-04 05:06:07', '2024-01-02 03:04:05')`, true, false},
		{"晚于-不成立", `DateIsAfter('2024-01-02 03:04:05', '2024-03-04 05:06:07')`, false, false},
		{"晚于-相等为false", `DateIsAfter('2024-01-02 03:04:05', '2024-01-02 03:04:05')`, false, false},
		{"早于-成立", `DateIsBefore('2024-01-02 03:04:05', '2024-03-04 05:06:07')`, true, false},
		{"早于-不成立", `DateIsBefore('2024-03-04 05:06:07', '2024-01-02 03:04:05')`, false, false},
		{"比较-入参是变量", `DateIsBefore(t1, t2)`, true, false},
		{"比较-非法时间", `DateIsBefore('abc', t2)`, nil, true},

		// 组合使用
		{"组合-Now往后推一天", `DateDiff(DateAdd(Now(), 1, 'day'), Now(), 'day')`, 1, false},
		{"组合-格式化后比较", `DateFormat('2024-01-02 03:04:05', '2006-01-02') < '2024-02-01'`, true, false},
	}

	for _, tc := range testCases {
		got, err := ruleengine.NewEngineLogic().EvaluateString(tc.rule, dateFuncArgs())
		if tc.isErr {
			if err == nil {
				t.Errorf("%s: 期望报错，实际返回 %v", tc.name, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: EvaluateString(%q) err = %v", tc.name, tc.rule, err)
			continue
		}
		gotStr, wantStr := conv.String(got), conv.String(tc.want)
		if gotStr != wantStr {
			t.Errorf("%s: EvaluateString(%q) = %v, want %v", tc.name, tc.rule, gotStr, wantStr)
		}
	}
}

// TestNowFunction Now 没有固定返回值，单独校验格式与内容范围
func TestNowFunction(t *testing.T) {
	before := time.Now().Add(-1 * time.Minute)
	after := time.Now().Add(1 * time.Minute)

	//默认格式
	got, err := ruleengine.NewEngineLogic().EvaluateString(`Now()`, nil)
	if err != nil {
		t.Fatalf("Now() err = %v", err)
	}
	gotStr := conv.String(got)
	gotTime, err := conv.Convert[time.Time](gotStr)
	if err != nil {
		t.Fatalf("Now() 返回值 %q 不是可解析的时间：%v", gotStr, err)
	}
	if gotTime.Before(before) || gotTime.After(after) {
		t.Errorf("Now() = %v, 期望在 %v ~ %v 之间", gotTime, before, after)
	}

	//指定格式
	got, err = ruleengine.NewEngineLogic().EvaluateString(`Now('2006/01/02')`, nil)
	if err != nil {
		t.Fatalf("Now('2006/01/02') err = %v", err)
	}
	if _, err := time.Parse("2006/01/02", conv.String(got)); err != nil {
		t.Errorf("Now('2006/01/02') = %v, 不符合指定格式：%v", got, err)
	}

	//unix 时间戳
	got, err = ruleengine.NewEngineLogic().EvaluateString(`Now('unix')`, nil)
	if err != nil {
		t.Fatalf("Now('unix') err = %v", err)
	}
	ts, err := conv.Convert[int64](got)
	if err != nil {
		t.Fatalf("Now('unix') = %v, 不是数字：%v", got, err)
	}
	if ts < before.Unix() || ts > after.Unix() {
		t.Errorf("Now('unix') = %v, 期望在 %v ~ %v 之间", ts, before.Unix(), after.Unix())
	}
}
func TestNowFunction1(t *testing.T) {
	ruleExpr := templates.NewRuleExprEngine()
	newArgs, _ := ruleExpr.RunString("Now()", nil)
	fmt.Println(newArgs)

	newArgs2, _ := ruleExpr.RunString("DateFormat('2026-10-05 00:23:45')", nil)
	fmt.Println(newArgs2)
	newArgs3, _ := ruleExpr.RunString("DateAdd('2026-10-05 03:04:05', 1)", nil)
	fmt.Println(newArgs3)
}
