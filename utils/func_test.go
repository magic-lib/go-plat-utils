package utils_test

import (
	"fmt"
	"github.com/magic-lib/go-plat-utils/id-generator/id"
	"github.com/magic-lib/go-plat-utils/utils"
	"testing"
)

func AA(str string, bb string, cc ...string) string {
	return str
}
func BB(str string, bb string) string {
	return str
}

func TestAesCbc(t *testing.T) {
	testCases := []*utils.TestStruct{
		{"变长参数，少参数", []any{AA, "aa"}, []any{"", false}, nil},
		{"变长参数，参数刚好少一个", []any{AA, "aa", "bb"}, []any{"aa", true}, nil},
		{"变长参数，参数刚好相等", []any{AA, "aa", "bb", "cc"}, []any{"aa", true}, nil},
		{"变长参数，多参数", []any{AA, "aa", "bb", "cc", "dd"}, []any{"aa", true}, nil},
		{"定长参数，少参数", []any{BB, "aa"}, []any{"", false}, nil},
		{"定长参数，参数刚好", []any{BB, "aa", "bb"}, []any{"aa", true}, nil},
		{"定长参数，参数多", []any{BB, "aa", "bb", "cc"}, []any{"aa", true}, nil},
	}
	utils.TestFunction(t, testCases, func(function any, args ...any) (string, bool) {
		ret, err := utils.FuncExecute(function, args...)
		if err != nil {
			return "", false
		}
		return ret[0].(string), true
	})
}

func TestUUID(t *testing.T) {
	aa := id.GetUUID("sssss")
	fmt.Println(aa)
}
