package conv

import "time"

// FormatFromUnixTime 将unix时间戳格式化为YYYYMMDD HH:MM:SS格式字符串
// FormatFromUnixTime(FullDate, 12321312)
func FormatFromUnixTime(formatStr ...interface{}) string {
	format := fullTimeForm
	var timeNum int64 = 0

	if len(formatStr) == 1 {
		if times, ok := formatStr[0].(int64); ok {
			timeNum = times
		} else if strTemp, ok := formatStr[0].(string); ok {
			if strTemp != "" {
				format = strTemp
			}
		}
	} else if len(formatStr) == 2 {
		if times, ok := formatStr[0].(int64); ok {
			timeNum = times
		} else if strTemp, ok := formatStr[0].(string); ok {
			if strTemp != "" {
				format = strTemp
			}
		}
		if times, ok := formatStr[1].(int64); ok {
			timeNum = times
		} else if strTemp, ok := formatStr[1].(string); ok {
			if strTemp != "" {
				format = strTemp
			}
		}
	}

	if timeNum > 0 {
		return time.Unix(timeNum, 0).Format(format)
	}
	return time.Now().Format(format)
}
