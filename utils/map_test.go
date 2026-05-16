package utils_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/utils"
	"net/url"
	"testing"
)

type AAA struct {
	Name string `json:"bName" db:"aName"`
	Age  string `json:"age"`
	Key  string `db:"key"`
	Key1 string `db:"key1" json:"bKey1"`
	Key2 string
}

func TestTag(t *testing.T) {
	err := fmt.Errorf("aaaaaa")
	mm := fmt.Errorf("fdsfs: %w", err)
	fmt.Println(mm.Error())
}
func TestMapToUrlParams(t *testing.T) {
	oneMap := map[string]any{
		"a":   1,
		"b":   "2",
		"c":   `/ ? # & = @ : , ; + - * () * ~!@#$%^&*()_+,./';:"?><`,
		"d":   true,
		"aaa": "wo 是谁",
	}
	aaa := utils.MapToUrlParams(oneMap)
	fmt.Println(aaa)

	bbb, err := url.QueryUnescape(aaa)
	fmt.Println(bbb, err)

	aa := url.QueryEscape("+ ")
	bb := url.PathEscape("+ ")
	fmt.Println(aa, bb)
}
