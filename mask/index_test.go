package mask_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/mask"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestMask(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"Phone", []any{""}, []any{""}, mask.Phone},
		{"Phone", []any{"1"}, []any{"1"}, mask.Phone},
		{"Phone", []any{"22"}, []any{"*2"}, mask.Phone},
		{"Phone", []any{"333"}, []any{"*33"}, mask.Phone},
		{"Phone", []any{"4444"}, []any{"**44"}, mask.Phone},
		{"Phone", []any{"55555"}, []any{"**555"}, mask.Phone},
		{"Phone", []any{"666666"}, []any{"***666"}, mask.Phone},
		{"Phone", []any{"7777777"}, []any{"***7777"}, mask.Phone},
		{"Phone", []any{"88888888"}, []any{"8***8888"}, mask.Phone},
		{"Phone", []any{"999999999"}, []any{"99***9999"}, mask.Phone},
		{"Phone", []any{"0000000000"}, []any{"00****0000"}, mask.Phone},
		{"Phone", []any{"13722223345"}, []any{"137****3345"}, mask.Phone},
		{"Email", []any{"zhangsan@go-mall.com"}, []any{"zh****an@go-mall.com"}, mask.Email},
		{"Email", []any{"you@go-mall.com"}, []any{"y*u@go-mall.com"}, mask.Email},
		{"Email", []any{"dear@go-mall.com"}, []any{"d**r@go-mall.com"}, mask.Email},
		{"RealName", []any{"张三"}, []any{"张*"}, mask.RealName},
		{"RealName", []any{"赵丽颖"}, []any{"赵*颖"}, mask.RealName},
		{"AutoCharacter", []any{"12345678", 0.5, "*"}, []any{"12****78"}, mask.AutoCharacter},
		{"AutoCharacter", []any{"1234567854354", 0.4, "*"}, []any{"1234*****4354"}, mask.AutoCharacter},
		{"AutoCharacter", []any{"1234567890", 0.3, "*"}, []any{"1234***890"}, mask.AutoCharacter},
	}
	utils.TestFunction(t, testCases, nil)

	aa := "aaaabbbb"
	bb := mask.AutoCharacter(aa, 0.5, "*")
	fmt.Println(aa, bb)

}

func TestIsSameNumber(t *testing.T) {
	cases := []struct {
		newNumber string
		oldNumber string
		want      bool
	}{
		// 完全相同
		{"13800138000", "13800138000", true},
		{"", "", true},
		{"abc", "abc", true},
		// 掩码匹配：oldNumber 中数字位必须一致，非数字位（掩码）可匹配任意字符
		{"13800138000", "138****8000", true},
		{"13800138000", "138********", true},
		{"13800138000", "********8000", false},
		{"15012345678", "1**********", true},
		{"138xx30000", "138***0000", true}, // 掩码位可匹配任意字符（包括字母）
		{"12345", "1234*", true},
		// 掩码但数字位不匹配
		{"13900138000", "138****8000", false},
		{"13800138001", "138****8000", false},
		{"12345", "12346", false},
		{"11111111111", "22222222222", false},
		// 长度不同
		{"13800138000", "1380013800", false},
		{"13800138000", "138001380001", false},
	}
	for _, c := range cases {
		if got := mask.IsMatch(c.newNumber, c.oldNumber, "*"); got != c.want {
			t.Errorf("IsMatch(%q, %q)=%v, want %v", c.newNumber, c.oldNumber, got, c.want)
		}
	}
}
