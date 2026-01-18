package utils_test

import (
	"fmt"
	"github.com/chuckpreslar/inflect"
	"github.com/magic-lib/go-plat-utils/conv"
	"github.com/magic-lib/go-plat-utils/utils"
	"github.com/serenize/snaker"
	"testing"
)

func TestString(t *testing.T) {
	kk := snaker.CamelToSnake("aaaaaMMaaDDD")
	fmt.Println(kk)
	kk = utils.VarNameConverter("aaaaaMMaaDDD", utils.Snake)
	fmt.Println(kk)
	kk = snaker.SnakeToCamel("aa_bb_cc")
	fmt.Println(kk)
	kk = utils.VarNameConverter("aa_bb_cc", utils.Pascal)
	fmt.Println(kk)
	kk = snaker.SnakeToCamelLower("aa_bb_cc")
	fmt.Println(kk)
	kk = utils.VarNameConverter("aa_bb_cc", utils.Camel)
	fmt.Println(kk)
	kk = inflect.Singularize("books") //去掉s
	fmt.Println(kk)
	kk = inflect.Singularize("bless") //去掉s
	fmt.Println(kk)
}
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
	for i := 0; i < 10000; i++ {
		kk := utils.RandomStringInt(4)
		fmt.Println(kk)
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
