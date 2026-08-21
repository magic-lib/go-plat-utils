package cond_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/cond"
	"github.com/wI2L/jsondiff"
	"testing"
	"time"
)

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

func TestIsZero2(t *testing.T) {
	type inner struct {
		A int
		B string
	}
	type testCase struct {
		name string
		in   any
		want bool
	}
	cases := []testCase{
		{"nil", nil, true},
		{"空字符串", "", true},
		{"空白字符串", "   ", false},
		{"非空字符串", "abc", false},
		{"int 0", 0, true},
		{"int 非0", 1, false},
		{"int64 0", int64(0), true},
		{"uint 0", uint(0), true},
		{"float 0", 0.0, true},
		{"float 非0", 1.5, false},
		{"bool false", false, true},
		{"bool true", true, false},
		{"nil 指针", (*int)(nil), true},
		{"非 nil 指针", func() *int { p := new(int); return p }(), false},
		{"空切片", []int{}, true},
		{"非空切片", []int{1}, false},
		{"nil 切片", []string(nil), true},
		{"nil map", map[string]int(nil), true},
		{"空 map", map[string]int{}, true},
		{"零值 time.Time", time.Time{}, true},
		{"非零 time.Time", time.Now(), false},
		{"零值结构体", inner{}, true},
		{"非零结构体", inner{A: 1}, false},
	}
	for _, c := range cases {
		if got := cond.IsZero(c.in); got != c.want {
			t.Errorf("%s IsZero(%v) = %v, want %v", c.name, c.in, got, c.want)
		}
	}
}

func TestInsertIgnore2(t *testing.T) {
	tt := time.Now()
	mm := tt.String()
	nn := cond.IsTime(mm)
	fmt.Println(mm, nn)
}

func TestIsUUID(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"e4ff48d4-ea6b-45b6-9217-35bc23e8a57f", true},  // 全小写
		{"E4FF48D4-EA6B-45B6-9217-35BC23E8A57F", true},  // 全大写
		{"e4ff48d4-ea6b-45b6-9217-35bc23E8a57f", false}, // 混用 -> 拒绝
		{"E4FF48D4-EA6B-45B6-9217-35bc23e8a57f", false}, // 混用 -> 拒绝
		{"e4ff48d4-ea6b-45b6-9217-35bc23e8a57", false},  // 长度不足
	}
	for _, c := range cases {
		if got := cond.IsUUID(c.in); got != c.want {
			t.Errorf("IsUUID(%q)=%v, want %v", c.in, got, c.want)
		}
	}
}
func TestIsSameJSON(t *testing.T) {
	cases := []struct {
		jsonA string
		jsonB string
		want  bool
	}{
		{"1", "1.0", true},
		{"1", "1", true},
		{"{\"a\":1}", "{\"a\":1}", true},
		{"{\"a\":2}", "{\"a\":1}", false},
		{"{\"a\":1,\"b\":2, \"c\":[1,2,3]}", "{\"b\":2, \"a\":1, \"d\":[1,2,3]}", false},
	}
	for _, c := range cases {
		if got := cond.IsSameJson(c.jsonA, c.jsonB); got != c.want {
			t.Errorf("IsSameJSON(%q, %q)=%v, want %v", c.jsonA, c.jsonB, got, c.want)
		}
	}
}
