package ruleengine

import (
	"fmt"
	"strings"
	"time"

	"github.com/magic-lib/go-plat-utils/conv"
)

// 本文件实现时间相关的内置函数：Now / DateFormat / DateAdd / DateSub /
// DateDiff / DateIsAfter / DateIsBefore。
//
// 约定：
//  1. 时间入参可以是 time.Time，也可以是能被 conv.Convert[time.Time] 识别的字符串，
//     例如 "2006-01-02 15:04:05"、"2006-01-02"、"2006-01-02T15:04:05Z"、
//     "20060102"、1700000000（秒级时间戳）等；
//  2. Now / DateFormat / DateAdd 默认返回秒级时间戳（数字，便于直接参与算术运算），
//     传了格式化参数时才按参数输出，例如 Now('2006-01-02')、DateFormat(t, '2006-01-02') 返回字符串；
//  3. DateSub 默认返回字符串，格式 "2006-01-02 15:04:05"。
//     这样在表达式里可以直接用 > < 比较：govaluate 对字符串按字典序比较，
//     而该格式的字典序正好等价于时间先后；
//  4. 单位 unit 支持 year/month/week/day/hour/minute/second，大小写与单复数都认，默认 day；
//  5. 格式化参数 layout 用 Go 的时间布局，另外额外支持 unix（秒时间戳）、
//     unixmilli（毫秒时间戳）两个特殊值；
//  6. 偏移量 delta 按整数处理（小数会被取整），传负数等于往反方向移。

const (
	// defaultDateLayout DateSub 默认的输出格式
	defaultDateLayout = time.DateTime
	// nowDefaultLayout Now 的默认输出：秒级时间戳
	nowDefaultLayout = layoutUnix
	// dateFormatDefaultLayout DateFormat 的默认输出：秒级时间戳
	dateFormatDefaultLayout = layoutUnix
	// dateAddDefaultLayout DateAdd 的默认输出：秒级时间戳
	dateAddDefaultLayout = layoutUnix
	// layoutUnix / layoutUnixMilli 两个特殊的格式化值
	layoutUnix      = "unix"
	layoutUnixMilli = "unixmilli"
)

// Now 当前时间：Now() 默认返回秒级时间戳（数字），
// 传了格式化参数时按参数输出，例如 Now('2006-01-02') 返回字符串、Now('unixmilli') 返回毫秒时间戳
func (r *customerFunc) Now(args ...any) (any, error) {
	return formatTimeValue(time.Now(), layoutFromArgs(args, 0, nowDefaultLayout))
}

// DateFormat 格式化时间：DateFormat(time, '2006-01-02')
// layout 省略时返回秒级时间戳（数字），也支持 unix / unixmilli
func (r *customerFunc) DateFormat(args ...any) (any, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("参数数量不对：%v", args)
	}
	timeVal, err := toTimeValue(args[0])
	if err != nil {
		return "", err
	}
	return formatTimeValue(timeVal, layoutFromArgs(args, 1, dateFormatDefaultLayout))
}

// DateAdd 时间加法：DateAdd(time, 1, 'day', '2006-01-02')
// unit 默认 day；layout 省略时返回秒级时间戳（数字），传了才按指定格式输出
func (r *customerFunc) DateAdd(args ...any) (any, error) {
	return r.dateShift(args, 1, dateAddDefaultLayout)
}

// DateDiff 两个时间的差值：DateDiff(t1, t2, 'hour')
// 返回 t1 - t2，单位默认 second；不足一个单位时按比例返回小数
func (r *customerFunc) DateDiff(args ...any) (any, error) {
	if len(args) < 2 {
		return 0, fmt.Errorf("参数数量不对：%v", args)
	}
	timeOne, err := toTimeValue(args[0])
	if err != nil {
		return 0, err
	}
	timeTwo, err := toTimeValue(args[1])
	if err != nil {
		return 0, err
	}
	unit, err := dateUnitFromArgs(args, 2, "second")
	if err != nil {
		return 0, err
	}
	return diffTime(timeOne, timeTwo, unit)
}

// DateIsAfter 第一个时间是否晚于第二个时间：DateIsAfter(t1, t2)
func (r *customerFunc) DateIsAfter(args ...any) (any, error) {
	return r.dateCompare(args, true)
}

// DateIsBefore 第一个时间是否早于第二个时间：DateIsBefore(t1, t2)
func (r *customerFunc) DateIsBefore(args ...any) (any, error) {
	return r.dateCompare(args, false)
}

// dateShift 时间平移，sign 为 1 表示往后加，-1 表示往前减；defaultLayout 为省略 layout 时的输出格式
func (r *customerFunc) dateShift(args []any, sign int, defaultLayout string) (any, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("参数数量不对：%v", args)
	}
	timeVal, err := toTimeValue(args[0])
	if err != nil {
		return "", err
	}
	delta, err := conv.Convert[int](args[1])
	if err != nil {
		return "", fmt.Errorf("偏移量不是数字：%v", args[1])
	}
	unit, err := dateUnitFromArgs(args, 2, "day")
	if err != nil {
		return "", err
	}
	return formatTimeValue(addTime(timeVal, sign*delta, unit), layoutFromArgs(args, 3, defaultLayout))
}

// dateCompare 两个时间比较先后，after 为 true 判“晚于”，false 判“早于”
func (r *customerFunc) dateCompare(args []any, after bool) (any, error) {
	if len(args) < 2 {
		return false, fmt.Errorf("参数数量不对：%v", args)
	}
	timeOne, err := toTimeValue(args[0])
	if err != nil {
		return false, err
	}
	timeTwo, err := toTimeValue(args[1])
	if err != nil {
		return false, err
	}
	if after {
		return timeOne.After(timeTwo), nil
	}
	return timeOne.Before(timeTwo), nil
}

// toTimeValue 把入参转成 time.Time，转不了时报错
func toTimeValue(val any) (time.Time, error) {
	if val == nil {
		return time.Time{}, fmt.Errorf("时间参数不能为空")
	}
	if timeVal, ok := val.(time.Time); ok {
		return timeVal, nil
	}
	timeVal, err := conv.Convert[time.Time](val)
	if err != nil {
		return time.Time{}, fmt.Errorf("参数不是时间类型：%v", val)
	}
	return timeVal, nil
}

// formatTimeValue 按 layout 输出时间，layout 为 unix / unixmilli 时输出时间戳。
// 注意：时间戳必须返回 float64 而不是 int64——govaluate 的 isFloat64 只认 float64，
// 返回 int64 会让 DateFormat(t) - 1、Now() > 100 这类运算直接报
// "it is not a number"。
func formatTimeValue(timeVal time.Time, layout string) (any, error) {
	switch strings.ToLower(layout) {
	case layoutUnix:
		return float64(timeVal.Unix()), nil
	case layoutUnixMilli:
		return float64(timeVal.UnixMilli()), nil
	}
	return timeVal.Format(layout), nil
}

// layoutFromArgs 取 args[idx] 作为格式化参数，不存在或为空时用 defaultLayout
func layoutFromArgs(args []any, idx int, defaultLayout string) string {
	if idx >= len(args) {
		return defaultLayout
	}
	if layout := strings.TrimSpace(conv.String(args[idx])); layout != "" {
		return layout
	}
	return defaultLayout
}

// dateUnitFromArgs 取 args[idx] 作为时间单位并归一化，不存在或为空时用 defaultUnit
func dateUnitFromArgs(args []any, idx int, defaultUnit string) (string, error) {
	if idx >= len(args) {
		return defaultUnit, nil
	}
	unitStr := strings.TrimSpace(conv.String(args[idx]))
	if unitStr == "" {
		return defaultUnit, nil
	}
	return dateUnit(unitStr)
}

// dateUnit 归一化时间单位，返回 year/month/week/day/hour/minute/second 之一
func dateUnit(unitStr string) (string, error) {
	switch strings.ToLower(unitStr) {
	case "y", "year", "years":
		return "year", nil
	case "month", "months":
		return "month", nil
	case "w", "week", "weeks":
		return "week", nil
	case "d", "day", "days":
		return "day", nil
	case "h", "hour", "hours":
		return "hour", nil
	case "min", "mins", "minute", "minutes":
		return "minute", nil
	case "s", "sec", "secs", "second", "seconds":
		return "second", nil
	}
	return "", fmt.Errorf("不支持的时间单位：%s", unitStr)
}

// addTime 按单位平移时间：日历单位走 AddDate，时钟单位走 Add
func addTime(timeVal time.Time, delta int, unit string) time.Time {
	switch unit {
	case "year":
		return timeVal.AddDate(delta, 0, 0)
	case "month":
		return timeVal.AddDate(0, delta, 0)
	case "week":
		return timeVal.AddDate(0, 0, delta*7)
	case "day":
		return timeVal.AddDate(0, 0, delta)
	case "hour":
		return timeVal.Add(time.Duration(delta) * time.Hour)
	case "minute":
		return timeVal.Add(time.Duration(delta) * time.Minute)
	case "second":
		return timeVal.Add(time.Duration(delta) * time.Second)
	}
	return timeVal
}

// diffTime 计算 timeOne - timeTwo 的差值：日历单位按日历差取整数，
// 时钟单位按绝对时长取比例（可以是小数）
func diffTime(timeOne, timeTwo time.Time, unit string) (any, error) {
	switch unit {
	case "year":
		years := timeOne.Year() - timeTwo.Year()
		if timeOne.Before(timeTwo.AddDate(years, 0, 0)) {
			years--
		}
		return float64(years), nil
	case "month":
		months := (timeOne.Year()-timeTwo.Year())*12 + int(timeOne.Month()) - int(timeTwo.Month())
		if timeOne.Before(timeTwo.AddDate(0, months, 0)) {
			months--
		}
		return float64(months), nil
	}

	var oneUnit time.Duration
	switch unit {
	case "week":
		oneUnit = 7 * 24 * time.Hour
	case "day":
		oneUnit = 24 * time.Hour
	case "hour":
		oneUnit = time.Hour
	case "minute":
		oneUnit = time.Minute
	case "second":
		oneUnit = time.Second
	default:
		return 0, fmt.Errorf("不支持的时间单位：%s", unit)
	}
	return float64(timeOne.Sub(timeTwo)) / float64(oneUnit), nil
}
