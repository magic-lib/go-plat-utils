package conf

import (
	"time"
)

var (
	sysTimeLocation = "UTC"
)

// SetTimeLocation 设置时区
func SetTimeLocation(location string) {
	sysTimeLocation = location
}

// TimeLocation 获得时区
func TimeLocation() *time.Location {
	if time.Local != nil {
		return time.Local
	}
	if cst, err := time.LoadLocation(sysTimeLocation); err == nil {
		return cst
	}
	return nil
}
