package cond_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/wI2L/jsondiff"
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

func TestIsJson(t *testing.T) {
	a := `{"accno":"0750000000","mid":0,"nid":"460660/99/3","name":"","from":"","bcode":""}`
	b := `{"mid":0,"accno":"0750000000","nid":"460660/99/3","name":"","from":"","bcode":""}`

	// 只判断是否相等（忽略key顺序）
	patch, err := jsondiff.CompareJSON([]byte(a), []byte(b))
	fmt.Println("相等？", len(patch) == 0, err) // true
}
