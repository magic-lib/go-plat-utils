package utils_test

import (
	"context"
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
		{"变长参数，少参数", []any{AA, "aa"}, []any{"aa", true}, nil},
		{"变长参数，参数刚好少一个", []any{AA, "aa", "bb"}, []any{"aa", true}, nil},
		{"变长参数，参数刚好相等", []any{AA, "aa", "bb", "cc"}, []any{"aa", true}, nil},
		{"变长参数，多参数", []any{AA, "aa", "bb", "cc", "dd"}, []any{"aa", true}, nil},
		{"定长参数，少参数", []any{BB, "aa"}, []any{"aa", true}, nil},
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
func TestArgType(t *testing.T) {
	ret1, b1, err1 := utils.FuncInTypeList(AA)
	ret2, b2, err2 := utils.FuncInTypeList(BB)
	fmt.Println(ret1, err1, ret2, err2, b1, b2)
}

func TestUUID(t *testing.T) {
	aa := id.GetUUID("sssss")
	fmt.Println(aa)
}

func GetUserTags(ctx context.Context, accountUserReq int) (string, error) {
	fmt.Println("rrrr")
	return "55555", nil
}

var newUserTags utils.ContextTypedHandler[int, string] = GetUserTags

func TestContextMethodToAnyHandler(t *testing.T) {
	var methodFun any = GetUserTags
	var newMethodFun any = newUserTags
	a, ok := utils.ContextMethodToAnyHandler[int, string](methodFun)
	fmt.Println(a, ok)
	b, ok2 := utils.ContextMethodToAnyHandler[int, string](newMethodFun)
	fmt.Println(b, ok2)
	if ok2 == nil {
		mm, err := b(nil, 1)
		fmt.Println(mm, err)
	}
}
