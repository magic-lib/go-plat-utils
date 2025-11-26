package cond

import (
	"regexp"
	"time"
)

// IsTime 是否是时间格式
func IsTime(dateTime string) bool {
	regPattern := "((([0-9]{3}[1-9]|[0-9]{2}[1-9][0-9]{1}|[0-9]{1}[1-9][0-9]{2}|[1-9][0-9]{3})-(((0[13578]|1[02])-"
	regPattern += "(0[1-9]|[12][0-9]|3[01]))|((0[469]|11)-(0[1-9]|[12][0-9]|30))|(02-(0[1-9]|[1][0-9]|2[0-8]))))|"
	regPattern += "((([0-9]{2})(0[48]|[2468][048]|[13579][26])|((0[48]|[2468][048]|[3579][26])00))-02-29))\\s"
	regPattern += "([0-1][0-9]|2[0-3]):([0-5][0-9]):([0-5][0-9])$"
	matched, err := regexp.Match(regPattern, []byte(dateTime))
	if err == nil {
		return matched
	}
	return false
}

// IsTimeEmpty 是否为空时间
func IsTimeEmpty(timeParam time.Time) bool {
	nilTime := time.Time{} //赋零值
	if timeParam == nilTime {
		return true
	}
	//因为有时区会影响到判断，这里会忽略时区的影响：0001-01-01 00:00:00 这样的格式，都为零值时间
	// 0001-01-01 00:00:00  0000-00-00 00:00:00  这两种情况
	return ((timeParam.Year() == 1 &&
		timeParam.Month() == 1 &&
		timeParam.Day() == 1) ||
		(timeParam.Year() == 0 &&
			timeParam.Month() == 0 &&
			timeParam.Day() == 0)) &&
		timeParam.Hour() == 0 &&
		timeParam.Minute() == 0 &&
		timeParam.Second() == 0 &&
		timeParam.Nanosecond() == 0
}
