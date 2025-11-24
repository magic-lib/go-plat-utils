package cond_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"testing"
	"time"
)

func TestIsUUID(t *testing.T) {
	isUUID := cond.IsUUID("e4ff48d4-ea6b-45b6-9217-35bc23e8a57f")
	fmt.Println(isUUID)
}
func TestIsZero(t *testing.T) {
	timeStr := "0001-01-01 00:00:00"
	layout := "2006-01-02 15:04:05"
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		fmt.Printf("加载时区失败：%v\n", err)
		return
	}
	time2, _ := time.ParseInLocation(layout, timeStr, loc)
	time3, _ := time.Parse(layout, timeStr)

	retBool := cond.IsZero(time2)
	if retBool {
		fmt.Println("sssss empty time")
	}
	retBool = cond.IsZero(time3)
	if retBool {
		fmt.Println("sssss empty time")
	}
}
