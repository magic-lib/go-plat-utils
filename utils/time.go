package utils

import (
	"github.com/magic-lib/go-plat-utils/conf"
	"time"

	"github.com/magic-lib/go-plat-utils/conv"
)

// GetSinceMilliTime 取得相差时间
func GetSinceMilliTime(timeStart time.Time) int64 {
	return time.Since(timeStart.In(conf.TimeLocation())).Milliseconds()
}

// NextDayDuration 得到当前时间到下一天零点的延时
func NextDayDuration() time.Duration {
	var sysTimeLocationTemp = conf.TimeLocation()
	year, month, day := time.Now().In(sysTimeLocationTemp).Add(time.Hour * 24).Date()
	next := time.Date(year, month, day, 0, 0, 0, 0, sysTimeLocationTemp)
	return next.Sub(time.Now())
}

// MilliTime 毫秒
func MilliTime() int64 {
	return time.Now().UnixMilli()
}

// NowUnix 当前时间的时间戳
func NowUnix() int {
	return int(time.Now().In(conf.TimeLocation()).Unix())
}

// DateSub 日期之间进行比较
func DateSub(oneTime time.Time, towTime time.Time) (time.Duration, bool) {
	newOneTime, ok1 := conv.Time(oneTime.Format("2006-01-02") + " 00:00:00")
	newTwoTime, ok2 := conv.Time(towTime.Format("2006-01-02") + " 00:00:00")
	if ok1 && ok2 {
		return newOneTime.Sub(newTwoTime), true
	}
	return 0, false
}
