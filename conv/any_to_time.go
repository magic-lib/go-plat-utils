package conv

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"reflect"
	"regexp"
	"strings"
	"time"
)

const (
	fullTimeForm     = "2006-01-02 15:04:05"
	FullDateForm     = "2006-01-02"
	ShortTimeForm10  = "0102150405"
	ShortTimeForm12  = "060102150405"
	ShortTimeForm14  = "20060102150405"
	ShortDateForm08  = "20060102"
	ShortMonthForm06 = "200601"
)

// Time 转换为Time
func Time(val interface{}) (time.Time, bool) {
	timeRet := time.Time{}
	if val == nil {
		return timeRet, true
	}
	reValue := reflect.ValueOf(val)
	for reValue.Kind() == reflect.Ptr {
		reValue = reValue.Elem()
		if !reValue.IsValid() {
			return timeRet, true
		}
		val = reValue.Interface()
		if val == nil {
			return timeRet, true
		}
		reValue = reflect.ValueOf(val)
	}
	if val == "" {
		return timeRet, true
	}

	if v, ok := val.(time.Time); ok {
		return v, true
	}

	valTemp := String(val)
	if timeTemp, ok := toTimeFromString(valTemp); ok {
		return timeTemp, ok
	}

	return timeRet, true
}

func milliTime() int64 {
	return time.Now().UnixMilli()
}

func toTimeFromNormal(v string) (time.Time, error) {
	tLen := len(v)
	if tLen == 0 {
		return time.Time{}, nil
	} else if tLen == 8 {
		return time.ParseInLocation(ShortDateForm08, v, time.Local)
	} else if tLen == len(time.ANSIC) {
		return time.Parse(time.ANSIC, v)
	} else if tLen == len(time.UnixDate) {
		return time.Parse(time.UnixDate, v)
	} else if tLen == len(time.RubyDate) {
		t, err := time.Parse(time.RFC850, v)
		if err != nil {
			t, err = time.Parse(time.RubyDate, v)
		}
		return t, err
	} else if tLen == len(time.RFC822Z) {
		return time.Parse(time.RFC822Z, v)
	} else if tLen == len(time.RFC1123) {
		return time.Parse(time.RFC1123, v)
	} else if tLen == len(time.RFC1123Z) {
		return time.Parse(time.RFC1123Z, v)
	} else if tLen == len(time.RFC3339) {
		return time.Parse(time.RFC3339, v)
	} else if tLen == len(time.RFC3339Nano) {
		return time.Parse(time.RFC3339Nano, v)
	} else if tLen == len("2025-03-28T18:59:24") {
		timeArray := strings.Split(v, "T")
		if len(timeArray) == 2 {
			return time.Parse(time.DateTime, timeArray[0]+" "+timeArray[1])
		}
		timeArray = strings.Split(v, " ")
		if len(timeArray) == 2 {
			return time.Parse(time.DateTime, v)
		}
	}

	return time.Time{}, fmt.Errorf("no found")
}

func toTimeFromString(v string) (time.Time, bool) {
	t, err := toTimeFromNormal(v)
	if err == nil {
		return t, true
	}

	tLen := len(v)

	if tLen == 10 {
		if cond.IsNumeric(v) {
			mcInt, _ := Int64(v)
			t = time.Unix(mcInt, 0)
			err = nil
			return t, true
		}
		t, err = time.ParseInLocation(FullDateForm, v, time.Local)
	} else if tLen == len(String(milliTime())) { //毫秒
		if cond.IsNumeric(v) {
			mcTempStr := v[0 : len(v)-3]
			mcInt, _ := Int64(mcTempStr)
			t = time.Unix(mcInt, 0)
			err = nil
			return t, true
		}
	} else if tLen == 19 { //毫秒
		t, err = time.ParseInLocation(fullTimeForm, v, time.Local)
		if err != nil {
			t, err = time.Parse(time.RFC822, v)
		}
	} else if tLen == len("2019-12-10T11:18:18.979878") ||
		tLen == len("2019-12-10T11:18:18.9798786") { //毫秒
		tempArr := strings.Split(v, ".")
		if len(tempArr) == 2 {
			timeTemp := tempArr[0]
			timeTemp = strings.Replace(timeTemp, "T", " ", 1)
			t, err = time.ParseInLocation(fullTimeForm, timeTemp, time.Local)
			if err != nil {
				t, err = time.Parse(time.RFC822, v)
			}
		}
	} else {
		if tLen > 19 {
			tempArr := strings.Split(v, ".")
			if len(tempArr) == 2 {
				timeTemp := tempArr[0]
				timeTemp = strings.Replace(timeTemp, "T", " ", 1)
				t, err = time.ParseInLocation(fullTimeForm, timeTemp, time.Local)
				if err == nil {
					return t, true
				}
			}
		}
		t, err = time.Parse(time.RFC1123, v)
	}

	if err != nil {
		{ //2023-04-14T10:09:00Z
			timePattern := "^(\\d{4})-(\\d{2})-(\\d{2})T(\\d{2}):(\\d{2}):(\\d{2})Z$"
			isFind, err := regexp.MatchString(timePattern, v)
			if err == nil {
				if isFind {
					regPattern, _ := regexp.Compile(timePattern)
					patternList := regPattern.FindAllStringSubmatch(v, -1)
					if len(patternList) == 1 {
						if len(patternList[0]) == 7 {
							v1 := fmt.Sprintf("%s-%s-%sT%s:%s:%s+00:00", patternList[0][1],
								patternList[0][2], patternList[0][3],
								patternList[0][4], patternList[0][5], patternList[0][6])
							return toTimeFromString(v1)
						}
					}
					return t, false
				}
			}
		}

		return t, false
	}
	return t, true
}
