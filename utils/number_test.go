package utils_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func TestPercent(t *testing.T) {
	var a int = 3
	var b float64 = 2.5

	kk := utils.Percent(a, b, 2)
	fmt.Println(kk)
}
func TestRandomInt64(t *testing.T) {
	for i := 0; i < 10000; i++ {
		data := utils.RandomIntInRange(10, 20)
		if data < 10 || data > 20 {
			t.Error("error", data)
		}
	}

	for i := 0; i < 10000; i++ {
		data := utils.RandomInt(5)
		if data < 10000 || data > 99999 {
			t.Error("error", data)
		}
	}
	for i := 0; i < 10000; i++ {
		data := utils.RandomInt(18)
		kk := conv.String(data)
		if len(kk) != 18 {
			t.Error("error", data)
		}
	}

	//testCases := []*utils.TestStruct{
	//	{"string 1 to int 1", []any{10, 20}, []any{1}, utils.RandomInt},
	//	{"string 1 to int 1", []any{10, 20}, []any{1}, utils.RandomInt},
	//	{"string 1 to int 1", []any{10, 20}, []any{1}, utils.RandomInt},
	//	{"string 1 to int 1", []any{10, 20}, []any{1}, utils.RandomInt},
	//	{"string 1 to int 1", []any{10, 20}, []any{1}, utils.RandomInt},
	//	{"string 1 to int 1", []any{10, 20}, []any{1}, utils.RandomInt},
	//}
	//utils.TestFunction(t, testCases, nil)
}
